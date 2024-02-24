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

package dashboard

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	providerV1alpha1 "github.com/argannor/provider-grafana/apis/v1alpha1"

	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/util/json"

	"github.com/argannor/provider-grafana/internal/controller/common"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	grafana "github.com/grafana/grafana-openapi-client-go/client"
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
	apisv1beta1 "github.com/argannor/provider-grafana/apis/v1beta1"
	"github.com/argannor/provider-grafana/internal/features"
)

const (
	errNotDashboard = "managed resource is not a Dashboard custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"
	errCredsFormat  = "credentials are not formatted as base64 encoded 'username:password' pair"
	errOrgIdNotInt  = "orgId is not an integer"
	errNoTitle      = "configJson does not contain a title for the dashboard"

	errNewClient             = "cannot create new Service"
	errFailedGetDashboard    = "cannot get Dashboard from Grafana API"
	errFailedCreateDashboard = "cannot create Dashboard"
	errFailedUpdateDashboard = "cannot update Dashboard"
	errFailedDeleteDashboard = "cannot delete Dashboard"

	errUnmarshalJson            = "cannot unmarshal JSON data"
	errInvalidDashboardResponse = "cannot parse dashboard response"
)

var (
	newService = func(config *grafana.TransportConfig) (common.GrafanaAPI, error) {
		client := *grafana.NewHTTPClientWithConfig(nil, config)
		return common.NewGrafanaAPI(client), nil
	}
)

// Setup adds a controller that reconciles Dashboard managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.DashboardGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), providerV1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.DashboardGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1beta1.ProviderConfigUsage{}),
			newServiceFn: newService,
			logger:       o.Logger}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.Dashboard{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        resource.Tracker
	logger       logging.Logger
	newServiceFn func(config *grafana.TransportConfig) (common.GrafanaAPI, error)
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Dashboard)
	if !ok {
		return nil, errors.New(errNotDashboard)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &apisv1beta1.ProviderConfig{}
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

	return &external{service: svc, logger: c.logger, kube: c.kube}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	service common.GrafanaAPI
	logger  logging.Logger
	kube    client.Client
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Dashboard)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotDashboard)
	}

	// orgId as int64
	orgId, err := strconv.ParseInt(*(cr.Spec.ForProvider.OrgID), 10, 64)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errOrgIdNotInt)
	}

	atGrafana, err := c.GetDashboard(orgId, cr)

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errFailedGetDashboard)
	}

	if atGrafana == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	upToDate := isUpToDate(cr, atGrafana)

	err = copyToStatusFromMeta(atGrafana, cr, *cr.Spec.ForProvider.OrgID)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	return managed.ExternalObservation{
		// Return false when the external resource does not exist. This lets
		// the managed resource reconciler know that it needs to call Create to
		// (re)create the resource, or that it has successfully been deleted.
		ResourceExists: true,

		// Return false when the external resource exists, but it not up to date
		// with the desired managed resource state. This lets the managed
		// resource reconciler know that it needs to call Update.
		ResourceUpToDate: upToDate,

		// Return any details that may be required to connect to the external
		// resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Dashboard)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDashboard)
	}

	// orgId as int64
	spec := cr.Spec.ForProvider
	orgId, err := strconv.ParseInt(*(spec.OrgID), 10, 64)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errOrgIdNotInt)
	}

	configJson, err := parseConfigJson(spec.ConfigJSON)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errUnmarshalJson)
	}

	command := &models.SaveDashboardCommand{
		Dashboard: configJson,
		IsFolder:  false,
		Message:   common.DefaultString(spec.Message, ""),
		Overwrite: common.DefaultBool(spec.Overwrite, false),
	}
	setFolderId(spec.Folder, command)

	_, err = c.service.CreateOrUpdateDashboard(orgId, command)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errFailedCreateDashboard)
	}

	cr.Status.AtProvider.ConfigJSON = cr.Spec.ForProvider.ConfigJSON

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func setFolderId(folder *string, command *models.SaveDashboardCommand) {
	if folder == nil {
		return
	}
	// if folder matches uuid regex, set as FolderUID
	if _, err := uuid.Parse(*folder); err == nil {
		command.FolderUID = *folder
	} else {
		// else set as FolderID
		folderId, err := strconv.ParseInt(*folder, 10, 64)
		if err == nil {
			// nolint: staticcheck
			command.FolderID = folderId
		}
	}
}

func parseConfigJson(configJson *string) (map[string]interface{}, error) {
	if configJson == nil {
		return nil, nil
	}
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(*configJson), &config); err != nil {
		return nil, errors.Wrap(err, errUnmarshalJson)
	}
	return config, nil

}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Dashboard)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDashboard)
	}

	// orgId as int64
	spec := cr.Spec.ForProvider
	orgId, err := strconv.ParseInt(*spec.OrgID, 10, 64)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errOrgIdNotInt)
	}

	configJson, err := parseConfigJson(spec.ConfigJSON)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUnmarshalJson)
	}
	configJson["id"] = cr.Status.AtProvider.DashboardID
	configJson["uid"] = cr.Status.AtProvider.UID

	command := &models.SaveDashboardCommand{
		Dashboard: configJson,
		IsFolder:  false,
		Message:   common.DefaultString(spec.Message, ""),
		Overwrite: common.DefaultBool(spec.Overwrite, false),
	}
	setFolderId(spec.Folder, command)

	response, err := c.service.CreateOrUpdateDashboard(orgId, command)

	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errFailedUpdateDashboard)
	}

	copyToStatus(response, cr, *spec.OrgID)
	cr.Status.AtProvider.ConfigJSON = cr.Spec.ForProvider.ConfigJSON

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Dashboard)
	if !ok {
		return errors.New(errNotDashboard)
	}

	// orgId as int64
	spec := cr.Spec.ForProvider
	orgId, err := strconv.ParseInt(*(spec.OrgID), 10, 64)
	if err != nil {
		return errors.Wrap(err, errOrgIdNotInt)
	}

	_, err = c.service.DeleteDashboard(orgId, *cr.Status.AtProvider.UID)

	return errors.Wrap(err, errFailedDeleteDashboard)
}

func copyToStatus(response *models.PostDashboardOKBody, cr *v1alpha1.Dashboard, orgId string) {
	id := fmt.Sprintf("%s:%s", orgId, *response.UID)
	cr.Status.AtProvider.ID = &id
	cr.Status.AtProvider.OrgID = &orgId
	cr.Status.AtProvider.UID = response.UID
	cr.Status.AtProvider.Folder = &response.FolderUID
	cr.Status.AtProvider.DashboardID = response.ID
	cr.Status.AtProvider.URL = response.URL
	cr.Status.AtProvider.Version = response.Version
}

func copyToStatusFromMeta(response *models.DashboardFullWithMeta, cr *v1alpha1.Dashboard, orgId string) error {
	dashboard, err := dashboardInDashboardFullWithMetaFromJSON(&response.Dashboard)
	if err != nil {
		return err
	}
	id := fmt.Sprintf("%s:%s", orgId, dashboard.UID)
	cr.Status.AtProvider.ID = &id
	cr.Status.AtProvider.OrgID = &orgId
	cr.Status.AtProvider.UID = &dashboard.UID
	cr.Status.AtProvider.Folder = &response.Meta.FolderUID
	cr.Status.AtProvider.DashboardID = &dashboard.ID
	cr.Status.AtProvider.URL = &response.Meta.URL
	cr.Status.AtProvider.Version = &dashboard.Version
	return nil
}

type dashboardInDashboardFullWithMeta struct {
	UID     string `json:"uid,omitempty"`
	ID      int64  `json:"id,omitempty"`
	Version int64  `json:"version,omitempty"`
}

func dashboardInDashboardFullWithMetaFromJSON(dashboard *models.JSON) (*dashboardInDashboardFullWithMeta, error) {
	if dashboard == nil {
		return nil, nil
	}
	asMap, ok := (*dashboard).(map[string]interface{})
	if !ok {
		return nil, errors.New(errInvalidDashboardResponse)
	}
	return &dashboardInDashboardFullWithMeta{
		UID:     asMap["uid"].(string),
		ID:      common.AsInt64(asMap["id"]),
		Version: common.AsInt64(asMap["version"]),
	}, nil
}

func isUpToDate(cr *v1alpha1.Dashboard, atGrafana *models.DashboardFullWithMeta) bool {
	// These fmt statements should be removed in the real implementation.
	spec := cr.Spec.ForProvider
	upToDate := true

	upToDate = upToDate && common.CompareOptional(spec.Folder, atGrafana.Meta.FolderUID, "")

	if cr.Status.AtProvider.UID != nil {
		// if the UID is still nil, we didn't have the chance to set the status yet, so we can't compare
		upToDate = upToDate && cr.Status.AtProvider.ConfigJSON != nil && common.CompareOptional(spec.ConfigJSON, *cr.Status.AtProvider.ConfigJSON, "")
	} else {
		// unfortunately we can't set it in the Create method, so we need to do it here, and only if it is during
		// observation after creation - otherwise it would interfere with change detection. During Update, the
		// status will be updated accordingly.
		cr.Status.AtProvider.ConfigJSON = cr.Spec.ForProvider.ConfigJSON
	}

	return upToDate
}

func (c *external) GetDashboard(orgId int64, cr *v1alpha1.Dashboard) (*models.DashboardFullWithMeta, error) {
	if cr.Status.AtProvider.UID != nil {
		return c.service.GetDashboardByUid(orgId, *cr.Status.AtProvider.UID)
	} else {
		configJson, err := parseConfigJson(cr.Spec.ForProvider.ConfigJSON)
		if err != nil {
			return nil, err
		}
		title, found := configJson["title"]
		if !found {
			return nil, errors.New(errNoTitle)
		}
		return c.service.GetDashboardByName(orgId, title.(string), cr.Spec.ForProvider.Folder)
	}
}
