/*
Copyright 2022 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package organization

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/google/go-cmp/cmp"
	"github.com/grafana/grafana-openapi-client-go/client/orgs"
	"github.com/grafana/grafana-openapi-client-go/client/users"
	"github.com/grafana/grafana-openapi-client-go/models"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/argannor/provider-grafana/apis/oss/v1alpha1"
	apisv1alpha1 "github.com/argannor/provider-grafana/apis/v1alpha1"
	"github.com/argannor/provider-grafana/internal/features"
	grafana "github.com/grafana/grafana-openapi-client-go/client"
)

const (
	errNotOrganization = "managed resource is not a Organization custom resource"
	errTrackPCUsage    = "cannot track ProviderConfig usage"
	errGetPC           = "cannot get ProviderConfig"
	errGetCreds        = "cannot get credentials"
	errCredsFormat     = "credentials are not formatted as base64 encoded 'username:password' pair"

	errNewClient = "cannot create new Service"

	errGetOrg         = "cannot get organization"
	errGetOrgUsers    = "cannot get users of organization"
	errUnexpectedRole = "unexpected role"
	errCreateOrg      = "cannot create organization"
	errDeleteOrg      = "cannot delete organization"
	errOrgNotFound    = "cannot find organization"
	errUpdateUser     = "cannot update user"
	errOrgNameChanged = "organization name cannot be changed"
)

// A NoOpService does nothing.
type NoOpService struct{}

var (
	newService = func(config *grafana.TransportConfig) (grafana.GrafanaHTTPAPI, error) {
		return *grafana.NewHTTPClientWithConfig(nil, config), nil
	}
)

type OrgUser struct {
	ID    int64
	Email string
	Role  string
}

type UserChange struct {
	Type ChangeType
	User OrgUser
}

type ChangeType int8

const (
	Add ChangeType = iota
	Update
	Remove
)

// Setup adds a controller that reconciles Organization managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.OrganizationGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1alpha1.StoreConfigGroupVersionKind))
	}

	logger := o.Logger.WithValues("controller", name)
	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.OrganizationGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			newServiceFn: newService,
			logger:       logger}),
		managed.WithLogger(logger),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.Organization{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn func(config *grafana.TransportConfig) (grafana.GrafanaHTTPAPI, error)
	logger       logging.Logger
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Organization)
	if !ok {
		return nil, errors.New(errNotOrganization)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &apisv1alpha1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	cd := pc.Spec.Credentials
	data, err := resource.CommonCredentialExtractor(ctx, cd.Source, c.kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(data))
	decodedCredentials, err := io.ReadAll(decoder)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}
	parts := strings.Split(string(decodedCredentials), ":")
	if len(parts) != 2 {
		return nil, errors.New(errCredsFormat)
	}

	clientCfg := grafana.DefaultTransportConfig()
	clientCfg = clientCfg.WithHost(fmt.Sprintf("%s:%d", pc.Spec.Host, pc.Spec.Port))
	clientCfg = clientCfg.WithSchemes(pc.Spec.Schemes)
	clientCfg.BasicAuth = url.UserPassword(parts[0], parts[1])

	svc, err := c.newServiceFn(clientCfg)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &external{service: svc, logger: c.logger}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	service grafana.GrafanaHTTPAPI
	logger  logging.Logger
}

type grafanaRole string

func (r grafanaRole) SetUsersInParameters(parameters *v1alpha1.OrganizationParameters, users []*string) error {
	switch r {
	case "Admin":
		parameters.Admins = users
	case "Editor":
		parameters.Editors = users
	case "Viewer":
		parameters.Viewers = users
	case "None":
		parameters.UsersWithoutAccess = users
	default:
		return errors.New(fmt.Sprintf("%s: %s", errUnexpectedRole, r))
	}
	return nil
}

type ApiError interface {
	error
	IsCode(code int) bool
}

func isCode(err error, code int) bool {
	if err == nil {
		return false
	}
	var oasError ApiError
	isOasError := errors.As(err, &oasError)
	if isOasError {
		return oasError.IsCode(code)
	}
	return false
}

func (c *external) observeActualParameters(cr *v1alpha1.Organization) (*v1alpha1.OrganizationParameters, error) {
	resp, err := c.service.Clone().WithOrgID(0).Orgs.GetOrgByName(*cr.Spec.ForProvider.Name)

	if isCode(err, 404) {
		return nil, nil
	}

	if err != nil {
		return nil, errors.Wrap(err, errGetOrg)
	}

	cr.Status.AtProvider.OrgID = &resp.Payload.ID
	idAsString := fmt.Sprintf("%d", resp.Payload.ID)
	cr.Status.AtProvider.ID = &idAsString
	cr.Status.AtProvider.Name = cr.Spec.ForProvider.Name
	cr.Status.AtProvider.AdminUser = cr.Spec.ForProvider.AdminUser
	cr.Status.AtProvider.CreateUsers = cr.Spec.ForProvider.CreateUsers

	userResp, err := c.service.Clone().WithOrgID(resp.Payload.ID).Orgs.GetOrgUsers(resp.Payload.ID)

	if err != nil {
		return nil, errors.Wrap(err, errGetOrgUsers)
	}

	actual := v1alpha1.OrganizationParameters{}
	actual.Name = &resp.Payload.Name
	roles := []grafanaRole{"Admin", "Editor", "Viewer", "None"}
	for _, role := range roles {
		var users []*string
		for _, user := range userResp.Payload {
			if user.Role == string(role) {
				users = append(users, &user.Email)
			}
		}
		err = role.SetUsersInParameters(&actual, users)
		if err != nil {
			return &actual, err
		}
	}

	cr.Status.AtProvider.Admins = actual.Admins
	cr.Status.AtProvider.Editors = actual.Editors
	cr.Status.AtProvider.Viewers = actual.Viewers
	cr.Status.AtProvider.UsersWithoutAccess = actual.UsersWithoutAccess

	return &actual, nil
}

func (c *external) usersEqualIgnoreOrder(a, b []*string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, user := range a {
		found := false
		for _, otherUser := range b {
			if *user == *otherUser {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Organization)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotOrganization)
	}

	actual, err := c.observeActualParameters(cr)
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	if actual == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	upToDate := true

	nameUpToDate := *actual.Name == *cr.Spec.ForProvider.Name
	upToDate = upToDate && nameUpToDate
	upToDate = upToDate && c.usersEqualIgnoreOrder(cr.Spec.ForProvider.Admins, actual.Admins)
	upToDate = upToDate && c.usersEqualIgnoreOrder(cr.Spec.ForProvider.Editors, actual.Editors)
	upToDate = upToDate && c.usersEqualIgnoreOrder(cr.Spec.ForProvider.Viewers, actual.Viewers)
	upToDate = upToDate && c.usersEqualIgnoreOrder(cr.Spec.ForProvider.UsersWithoutAccess, actual.UsersWithoutAccess)

	delta := cmp.Diff(cr.Spec.ForProvider, *actual)

	return managed.ExternalObservation{
		// Return false when the external resource does not exist. This lets
		// the managed resource reconciler know that it needs to call Create to
		// (re)create the resource, or that it has successfully been deleted.
		ResourceExists: true,

		// Return false when the external resource exists, but it not up to date
		// with the desired managed resource state. This lets the managed
		// resource reconciler know that it needs to call Update.
		ResourceUpToDate: upToDate,

		Diff: delta,

		// Return any details that may be required to connect to the external
		// resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Organization)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotOrganization)
	}

	response, err := c.service.Clone().WithOrgID(0).Orgs.CreateOrg(&models.CreateOrgCommand{
		Name: *cr.Spec.ForProvider.Name,
	})

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateOrg)
	}

	cr.Status.AtProvider.OrgID = response.Payload.OrgID
	idAsString := fmt.Sprintf("%d", response.Payload.OrgID)
	cr.Status.AtProvider.ID = &idAsString

	err = c.updateUsers(cr, v1alpha1.OrganizationParameters{}, response.Payload.OrgID)

	// TODO: according to the documentation we should not return an error if the resource already exists, but we need
	//   to ensure, that the existing resource should be adopted somehow according to
	//   https://github.com/crossplane/crossplane-runtime/issues/27
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateOrg)
}

func (c *external) updateUsers(cr *v1alpha1.Organization, actual v1alpha1.OrganizationParameters, orgID *int64) error {
	var err error
	client := c.service.Clone()
	changes := userChanges(mapUsers(actual), mapUsers(cr.Spec.ForProvider))
	changes, err = c.addUserIdsToChanges(&cr.Spec.ForProvider, changes, *orgID)
	if err != nil {
		return errors.Wrap(err, errUpdateUser)
	}
	for _, change := range changes {
		u := change.User
		switch change.Type {
		case Add:
			_, err = client.Orgs.AddOrgUser(*orgID, &models.AddOrgUserCommand{LoginOrEmail: u.Email, Role: u.Role})
		case Update:
			params := orgs.NewUpdateOrgUserParams().WithOrgID(*orgID).WithUserID(u.ID).WithBody(&models.UpdateOrgUserCommand{Role: u.Role})
			_, err = client.Orgs.UpdateOrgUser(params)
		case Remove:
			_, err = client.Orgs.RemoveOrgUser(u.ID, *orgID)
		}
		if err != nil && !strings.Contains(err.Error(), "409") {
			return errors.Wrap(err, errUpdateUser)
		}
	}
	return nil
}

func mapUsers(p v1alpha1.OrganizationParameters) map[string]OrgUser {
	users := make(map[string]OrgUser)
	for _, email := range p.Admins {
		users[*email] = OrgUser{Email: *email, Role: "Admin"}
	}
	for _, email := range p.Editors {
		users[*email] = OrgUser{Email: *email, Role: "Editor"}
	}
	for _, email := range p.Viewers {
		users[*email] = OrgUser{Email: *email, Role: "Viewer"}
	}
	for _, email := range p.UsersWithoutAccess {
		users[*email] = OrgUser{Email: *email, Role: "None"}
	}
	return users
}

func (c *external) getAllUsers() ([]*models.UserSearchHitDTO, error) {
	var allUsers []*models.UserSearchHitDTO
	var page int64 = 1
	params := users.NewSearchUsersParams().WithDefaults()
	client := c.service.Clone()

	for {
		resp, err := client.Users.SearchUsers(params.WithPage(&page), nil)
		if err != nil {
			return nil, err
		}

		allUsers = append(allUsers, resp.Payload...)
		if len(resp.Payload) != int(*params.Perpage) {
			break
		}
		page++
	}
	return allUsers, nil
}

func (c *external) createUser(user string) (int64, error) {
	client := c.service.Clone()
	n := 64
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		return 0, err
	}
	pass := string(bytes[:n])
	u := models.AdminCreateUserForm{
		Name:     user,
		Login:    user,
		Email:    user,
		Password: pass,
	}
	resp, err := client.AdminUsers.AdminCreateUser(&u)
	if err != nil {
		return 0, err
	}
	return resp.Payload.ID, err
}

// nolint: gocyclo
func (c *external) addUserIdsToChanges(d *v1alpha1.OrganizationParameters, changes []UserChange, orgId int64) ([]UserChange, error) {
	gUserMap := make(map[string]int64)
	gUsers, err := c.getAllUsers()
	if err != nil {
		return nil, err
	}
	for _, u := range gUsers {
		gUserMap[u.Email] = u.ID
	}
	output := make([]UserChange, 0)
	create := true
	if d.CreateUsers != nil {
		create = *d.CreateUsers
	}
	for _, change := range changes {
		id, ok := gUserMap[change.User.Email]
		if !ok && change.Type == Remove {
			c.logger.Info(fmt.Sprintf("can't remove user %s from organization %d because it no longer exists in grafana", change.User.Email, orgId))
			continue
		}
		if !ok && !create {
			return nil, fmt.Errorf("error adding user %s. User does not exist in Grafana", change.User.Email)
		}
		if !ok && create {
			id, err = c.createUser(change.User.Email)
			if err != nil {
				return nil, err
			}
		}
		change.User.ID = id
		output = append(output, change)
	}
	return output, nil
}

func userChanges(stateUsers, configUsers map[string]OrgUser) []UserChange {
	var changes []UserChange
	for _, user := range configUsers {
		sUser, ok := stateUsers[user.Email]
		if !ok {
			// User doesn't exist in Grafana's state for the organization, should be added.
			changes = append(changes, UserChange{Add, user})
			continue
		}
		if sUser.Role != user.Role {
			// Update the user as they're configured with a different role than
			// what is in Grafana's state.
			changes = append(changes, UserChange{Update, user})
		}
	}
	for _, user := range stateUsers {
		if _, ok := configUsers[user.Email]; !ok {
			// User exists in Grafana's state for the organization, but isn't
			// present in the organization configuration, should be removed.
			changes = append(changes, UserChange{Remove, user})
		}
	}
	return changes
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Organization)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotOrganization)
	}

	actual, err := c.observeActualParameters(cr)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}
	if actual == nil {
		return managed.ExternalUpdate{}, errors.New(errOrgNotFound)
	}

	nameUpToDate := *actual.Name == *cr.Spec.ForProvider.Name
	if !nameUpToDate {
		return managed.ExternalUpdate{}, errors.New(errOrgNameChanged)
	}

	usersUpToDate := true
	usersUpToDate = usersUpToDate && c.usersEqualIgnoreOrder(cr.Spec.ForProvider.Admins, actual.Admins)
	usersUpToDate = usersUpToDate && c.usersEqualIgnoreOrder(cr.Spec.ForProvider.Editors, actual.Editors)
	usersUpToDate = usersUpToDate && c.usersEqualIgnoreOrder(cr.Spec.ForProvider.Viewers, actual.Viewers)
	usersUpToDate = usersUpToDate && c.usersEqualIgnoreOrder(cr.Spec.ForProvider.UsersWithoutAccess, actual.UsersWithoutAccess)

	if !usersUpToDate {
		err = c.updateUsers(cr, *actual, cr.Status.AtProvider.OrgID)
	}

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, err
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Organization)
	if !ok {
		return errors.New(errNotOrganization)
	}

	orgID := cr.Status.AtProvider.OrgID
	if orgID == nil {
		return nil
	}

	client := c.service.Clone().WithOrgID(*orgID)

	currentUser, err := client.SignedInUser.GetSignedInUser()
	if err != nil {
		return errors.Wrap(err, errDeleteOrg)
	}

	if currentUser.Payload.OrgID == *orgID {
		err = c.switchToLowestOrgId()
	}
	if err != nil {
		return errors.Wrap(err, errDeleteOrg)
	}

	_, err = c.service.WithOrgID(0).Orgs.DeleteOrgByID(*orgID)
	return errors.Wrap(err, errDeleteOrg)
}

func (c *external) switchToLowestOrgId() error {
	orgs, err := c.getAllOrgs()
	if err != nil {
		return err
	}
	var orgId int64
	orgId = 9999999
	for _, org := range orgs {
		if org.ID < orgId {
			orgId = org.ID
		}
	}
	_, err = c.service.SignedInUser.UserSetUsingOrg(orgId)
	return err
}

func (c *external) getAllOrgs() ([]*models.OrgDTO, error) {
	var allOrgs []*models.OrgDTO
	var page int64 = 1
	params := orgs.NewSearchOrgsParams().WithDefaults()
	client := c.service.Clone()
	for {
		resp, err := client.Orgs.SearchOrgs(params.WithPage(&page), nil)
		if err != nil {
			return nil, err
		}

		allOrgs = append(allOrgs, resp.Payload...)
		if len(resp.Payload) != int(*params.Perpage) {
			break
		}
		page++
	}
	return allOrgs, nil
}
