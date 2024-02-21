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

package datasource

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
	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	grafana "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/pkg/errors"
	kubeV1 "k8s.io/api/core/v1"
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
	errNotDataSource = "managed resource is not a DataSource custom resource"
	errTrackPCUsage  = "cannot track ProviderConfig usage"
	errGetPC         = "cannot get ProviderConfig"
	errGetCreds      = "cannot get credentials"
	errCredsFormat   = "credentials are not formatted as base64 encoded 'username:password' pair"
	errOrgIdNotInt   = "orgId is not an integer"
	errNameChange    = "cannot change name of DataSource"

	errNewClient              = "cannot create new Service"
	errFailedGetDataSource    = "cannot get DataSource from Grafana API"
	errFailedGetHeadersSecret = "cannot get referenced HttpHeadersSecret"
	errFailedCreateDataSource = "cannot create DataSource"
	errFailedUpdateDataSource = "cannot update DataSource"
	errFailedDeleteDataSource = "cannot delete DataSource"
	errGetSecret              = "cannot get Secret"
	errGetSecureJsonData      = "cannot get referenced SecureJSONDataEncodedSecret"

	errUnmarshalJson       = "cannot unmarshal JSON data"
	errUnmarshalSecureJson = "cannot unmarshal secure JSON data"
)

var (
	newService = func(config *grafana.TransportConfig) (common.GrafanaAPI, error) {
		client := *grafana.NewHTTPClientWithConfig(nil, config)
		return common.NewGrafanaAPI(client), nil
	}
)

// Setup adds a controller that reconciles DataSource managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.DataSourceGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.DataSourceGroupVersionKind),
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
		For(&v1alpha1.DataSource{}).
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
	cr, ok := mg.(*v1alpha1.DataSource)
	if !ok {
		return nil, errors.New(errNotDataSource)
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
	cr, ok := mg.(*v1alpha1.DataSource)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotDataSource)
	}

	// orgId as int64
	orgId, err := strconv.ParseInt(*(cr.Spec.ForProvider.OrgID), 10, 64)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errOrgIdNotInt)
	}

	atGrafana, err := c.GetDataSource(orgId, cr)

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errFailedGetDataSource)
	}

	if atGrafana == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	var httpHeaderSecret *kubeV1.Secret
	if cr.Spec.ForProvider.HTTPHeadersSecretRef != nil {
		httpHeaderSecret, err = c.getSecret(ctx, *cr.Spec.ForProvider.HTTPHeadersSecretRef)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errFailedGetHeadersSecret)
		}
	}

	var secureJsonDataEncoded *string
	if cr.Spec.ForProvider.SecureJSONDataEncodedSecretRef != nil {
		secureJsonDataEncoded, err = c.getValueFromSecret(ctx, *cr.Spec.ForProvider.SecureJSONDataEncodedSecretRef)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errGetSecret)
		}
	}

	upToDate, err := isUpToDate(cr, atGrafana, orgId, httpHeaderSecret, secureJsonDataEncoded)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	copyToStatus(atGrafana, cr)

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
	cr, ok := mg.(*v1alpha1.DataSource)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDataSource)
	}

	// orgId as int64
	spec := cr.Spec.ForProvider
	orgId, err := strconv.ParseInt(*(spec.OrgID), 10, 64)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errOrgIdNotInt)
	}

	jsonData, secureJsonData, err := c.MakeJsonData(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	response, err := c.service.CreateDataSource(orgId, &models.AddDataSourceCommand{
		Access:          models.DsAccess(defaultString(spec.AccessMode, "proxy")),
		BasicAuth:       defaultBool(spec.BasicAuthEnabled, false),
		BasicAuthUser:   defaultString(spec.BasicAuthUsername, ""),
		Database:        defaultString(spec.DatabaseName, ""),
		IsDefault:       defaultBool(spec.IsDefault, false),
		JSONData:        *jsonData,
		Name:            defaultString(spec.Name, cr.Name),
		SecureJSONData:  *secureJsonData,
		Type:            defaultString(spec.Type, ""),
		UID:             defaultString(spec.UID, ""),
		URL:             defaultString(spec.URL, ""),
		User:            defaultString(spec.Username, ""),
		WithCredentials: false,
	})

	copyToStatus(response.Datasource, cr)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errFailedCreateDataSource)
	}

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.DataSource)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDataSource)
	}

	if *cr.Spec.ForProvider.Name != *cr.Status.AtProvider.Name {
		return managed.ExternalUpdate{}, errors.New(errNameChange)
	}

	// orgId as int64
	spec := cr.Spec.ForProvider
	orgId, err := strconv.ParseInt(*(spec.OrgID), 10, 64)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errOrgIdNotInt)
	}

	jsonData, secureJsonData, err := c.MakeJsonData(ctx, cr)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	response, err := c.service.UpdateDataSource(orgId, *cr.Status.AtProvider.ID, &models.UpdateDataSourceCommand{
		Access:          models.DsAccess(defaultString(spec.AccessMode, "proxy")),
		BasicAuth:       defaultBool(spec.BasicAuthEnabled, false),
		BasicAuthUser:   defaultString(spec.BasicAuthUsername, ""),
		Database:        defaultString(spec.DatabaseName, ""),
		IsDefault:       defaultBool(spec.IsDefault, false),
		JSONData:        *jsonData,
		Name:            defaultString(spec.Name, cr.Name),
		SecureJSONData:  *secureJsonData,
		Type:            defaultString(spec.Type, ""),
		UID:             defaultString(spec.UID, ""),
		URL:             defaultString(spec.URL, ""),
		User:            defaultString(spec.Username, ""),
		WithCredentials: false,
	})

	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errFailedUpdateDataSource)
	}

	copyToStatus(response.Datasource, cr)

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.DataSource)
	if !ok {
		return errors.New(errNotDataSource)
	}

	// orgId as int64
	spec := cr.Spec.ForProvider
	orgId, err := strconv.ParseInt(*(spec.OrgID), 10, 64)
	if err != nil {
		return errors.Wrap(err, errOrgIdNotInt)
	}

	_, err = c.service.DeleteDataSource(orgId, *cr.Status.AtProvider.ID)

	return errors.Wrap(err, errFailedDeleteDataSource)
}

func copyToStatus(response *models.DataSource, cr *v1alpha1.DataSource) {
	dataSourceId := fmt.Sprintf("%d", response.ID)
	orgIdAsString := fmt.Sprintf("%d", response.OrgID)
	cr.Status.AtProvider.ID = &dataSourceId
	cr.Status.AtProvider.OrgID = &orgIdAsString
	cr.Status.AtProvider.UID = &response.UID
	cr.Status.AtProvider.Name = &response.Name
	cr.Status.AtProvider.Username = &response.User
	cr.Status.AtProvider.IsDefault = &response.IsDefault
	cr.Status.AtProvider.BasicAuthEnabled = &response.BasicAuth
	cr.Status.AtProvider.BasicAuthUsername = &response.BasicAuthUser
	cr.Status.AtProvider.DatabaseName = &response.Database
	cr.Status.AtProvider.Type = &response.Type
	cr.Status.AtProvider.URL = &response.URL
}

// nolint: gocyclo
func isUpToDate(cr *v1alpha1.DataSource, atGrafana *models.DataSource, orgId int64, httpHeaderSecret *kubeV1.Secret, secureJsonDataEncoded *string) (bool, error) {
	// These fmt statements should be removed in the real implementation.
	spec := cr.Spec.ForProvider
	upToDate := true

	jd, err := makeJSONData(spec.JSONDataEncoded)
	if err != nil {
		return false, err
	}
	sjd, err := makeSecureJSONData(secureJsonDataEncoded)
	if err != nil {
		return false, err
	}
	httpHeaderMap := secretToStringMap(httpHeaderSecret)
	jsonData, secureJSONData := jsonDataWithHeaders(jd, sjd, httpHeaderMap)

	name := ""
	if spec.Name == nil {
		name = cr.Name
	} else {
		name = *spec.Name
	}

	upToDate = upToDate && name == atGrafana.Name
	upToDate = upToDate && *spec.Type == atGrafana.Type
	upToDate = upToDate && compareOptional(spec.AccessMode, string(atGrafana.Access), "proxy")
	upToDate = upToDate && compareOptional(spec.BasicAuthEnabled, atGrafana.BasicAuth, false)
	upToDate = upToDate && compareOptional(spec.BasicAuthUsername, atGrafana.BasicAuthUser, "")
	upToDate = upToDate && compareOptional(spec.DatabaseName, atGrafana.Database, "")
	upToDate = upToDate && compareOptional(spec.IsDefault, atGrafana.IsDefault, false)
	upToDate = upToDate && (spec.UID == nil || (*spec.UID == atGrafana.UID))
	upToDate = upToDate && compareOptional(spec.URL, atGrafana.URL, "")
	upToDate = upToDate && compareOptional(spec.Username, atGrafana.User, "")
	upToDate = upToDate && orgId == atGrafana.OrgID
	upToDate = upToDate && compareMap(jsonData, atGrafana.JSONData.(map[string]interface{}))
	// secure fields are not returned by the API, so we can't compare them
	upToDate = upToDate && compareMapKeys(secureJSONData, atGrafana.SecureJSONFields)
	// TODO: since the values are not included in the response, we can't check if they need to be updated. For this we
	//   would need to store a hash of the secret data in the status and compare against that. It needs to be stable
	//   against reordering of the keys and the values.

	return upToDate, err
}

func (c *external) GetDataSource(orgId int64, cr *v1alpha1.DataSource) (*models.DataSource, error) {
	if cr.Status.AtProvider.ID != nil {
		return c.service.GetDataSourceById(orgId, *cr.Status.AtProvider.ID)
	} else {
		return c.service.GetDataSourceByName(orgId, *cr.Spec.ForProvider.Name)
	}
}

func (c *external) MakeJsonData(ctx context.Context, cr *v1alpha1.DataSource) (*map[string]interface{}, *map[string]string, error) {
	jsonData, err := makeJSONData(cr.Spec.ForProvider.JSONDataEncoded)
	if err != nil {
		return nil, nil, err
	}

	var httpHeaderSecret *kubeV1.Secret
	if cr.Spec.ForProvider.HTTPHeadersSecretRef != nil {
		httpHeaderSecret, err = c.getSecret(ctx, *cr.Spec.ForProvider.HTTPHeadersSecretRef)
		if err != nil {
			return nil, nil, errors.Wrap(err, errFailedGetHeadersSecret)
		}
	}

	var secureJsonDataEncoded *string
	if cr.Spec.ForProvider.SecureJSONDataEncodedSecretRef != nil {
		secureJsonDataEncoded, err = c.getValueFromSecret(ctx, *cr.Spec.ForProvider.SecureJSONDataEncodedSecretRef)
		if err != nil {
			return nil, nil, errors.Wrap(err, errGetSecret)
		}
	}

	secureJSONData, err := makeSecureJSONData(secureJsonDataEncoded)
	if err != nil {
		return nil, nil, err
	}
	httpHeaderMap := secretToStringMap(httpHeaderSecret)
	jsonData, secureJSONData = jsonDataWithHeaders(jsonData, secureJSONData, httpHeaderMap)
	return &jsonData, &secureJSONData, err
}

func (c *external) getSecret(ctx context.Context, reference v1.SecretReference) (*kubeV1.Secret, error) {
	var secret kubeV1.Secret
	err := c.kube.Get(ctx, types.NamespacedName{Name: reference.Name, Namespace: reference.Namespace}, &secret)
	return &secret, err
}
