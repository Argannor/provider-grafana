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

package folder

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

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
	apisv1alpha1 "github.com/argannor/provider-grafana/apis/v1alpha1"
	"github.com/argannor/provider-grafana/internal/features"
)

const (
	errNotFolder    = "managed resource is not a Folder custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"
	errCredsFormat  = "credentials are not formatted as base64 encoded 'username:password' pair"
	errOrgIdNotInt  = "orgId is not an integer"
	errIdNotInt     = "folder ID is not an integer"

	errNewClient          = "cannot create new Service"
	errFailedGetFolder    = "cannot get Folder from Grafana API"
	errFailedCreateFolder = "cannot create Folder"
	errFailedUpdateFolder = "cannot update Folder"
	errFailedDeleteFolder = "cannot delete Folder"
)

var (
	newService = func(config *grafana.TransportConfig) (common.GrafanaAPI, error) {
		client := *grafana.NewHTTPClientWithConfig(nil, config)
		return common.NewGrafanaAPI(client), nil
	}
)

// Setup adds a controller that reconciles Folder managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.FolderGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.FolderGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
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
		For(&v1alpha1.Folder{}).
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
	cr, ok := mg.(*v1alpha1.Folder)
	if !ok {
		return nil, errors.New(errNotFolder)
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
	cr, ok := mg.(*v1alpha1.Folder)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotFolder)
	}

	// orgId as int64
	orgId, err := strconv.ParseInt(*(cr.Spec.ForProvider.OrgID), 10, 64)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errOrgIdNotInt)
	}

	atGrafana, err := c.GetFolder(orgId, cr)

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errFailedGetFolder)
	}

	if atGrafana == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	upToDate := isUpToDate(cr, atGrafana)

	copyToStatus(atGrafana, cr, *cr.Spec.ForProvider.OrgID)
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
	cr, ok := mg.(*v1alpha1.Folder)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotFolder)
	}

	// orgId as int64
	spec := cr.Spec.ForProvider
	orgId, err := strconv.ParseInt(*(spec.OrgID), 10, 64)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errOrgIdNotInt)
	}

	command := &models.CreateFolderCommand{
		ParentUID: common.DefaultString(spec.ParentFolderUID, ""),
		Title:     common.DefaultString(spec.Title, ""),
		UID:       common.DefaultString(spec.UID, ""),
	}

	_, err = c.service.CreateFolder(orgId, command)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errFailedCreateFolder)
	}

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Folder)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotFolder)
	}

	// orgId as int64
	spec := cr.Spec.ForProvider
	orgId, err := strconv.ParseInt(*spec.OrgID, 10, 64)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errOrgIdNotInt)
	}

	command := &models.UpdateFolderCommand{
		Title:   common.DefaultString(spec.Title, ""),
		Version: *cr.Status.AtProvider.Version,
		// Overwrite?
	}

	response, err := c.service.UpdateFolder(orgId, *cr.Status.AtProvider.UID, command)

	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errFailedUpdateFolder)
	}

	copyToStatus(response, cr, *spec.OrgID)

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Folder)
	if !ok {
		return errors.New(errNotFolder)
	}

	// orgId as int64
	spec := cr.Spec.ForProvider
	orgId, err := strconv.ParseInt(*(spec.OrgID), 10, 64)
	if err != nil {
		return errors.Wrap(err, errOrgIdNotInt)
	}

	_, err = c.service.DeleteFolder(orgId, *cr.Status.AtProvider.UID)

	return errors.Wrap(err, errFailedDeleteFolder)
}

func copyToStatus(response *models.Folder, cr *v1alpha1.Folder, orgId string) {
	id := fmt.Sprintf("%s:%s", orgId, response.UID)
	cr.Status.AtProvider.ID = &id
	cr.Status.AtProvider.OrgID = &orgId
	cr.Status.AtProvider.UID = &response.UID
	cr.Status.AtProvider.Title = &response.Title
	cr.Status.AtProvider.ParentFolderUID = &response.ParentUID
	cr.Status.AtProvider.URL = &response.URL
	cr.Status.AtProvider.Version = &response.Version
}

func isUpToDate(cr *v1alpha1.Folder, atGrafana *models.Folder) bool {
	spec := cr.Spec.ForProvider
	upToDate := true

	upToDate = upToDate && common.CompareOptional(spec.Title, atGrafana.Title, "")

	return upToDate
}

func (c *external) GetFolder(orgId int64, cr *v1alpha1.Folder) (*models.Folder, error) {
	switch status := cr.Status.AtProvider; {
	case status.UID != nil:
		return c.service.GetFolderByUid(orgId, *status.UID)
	case status.ID != nil:
		id := strings.Split(*cr.Status.AtProvider.ID, ":")[1]
		idAsInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, errIdNotInt)
		}
		return c.service.GetFolderById(orgId, idAsInt)
	default:
		return c.service.GetFolderByName(orgId, *cr.Spec.ForProvider.Title, cr.Spec.ForProvider.ParentFolderUID)
	}
}
