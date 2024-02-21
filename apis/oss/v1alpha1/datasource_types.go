// SPDX-FileCopyrightText: 2023 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

/*
Copyright 2022 Upbound Inc.
*/

package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

type DataSourceInitParameters struct {

	// (String) The method by which Grafana will access the data source: proxy or direct. Defaults to proxy.
	// The method by which Grafana will access the data source: `proxy` or `direct`. Defaults to `proxy`.
	AccessMode *string `json:"accessMode,omitempty" tf:"access_mode,omitempty"`

	// (Boolean) Whether to enable basic auth for the data source. Defaults to false.
	// Whether to enable basic auth for the data source. Defaults to `false`.
	BasicAuthEnabled *bool `json:"basicAuthEnabled,omitempty" tf:"basic_auth_enabled,omitempty"`

	// (String) Basic auth username. Defaults to “.
	// Basic auth username. Defaults to “.
	BasicAuthUsername *string `json:"basicAuthUsername,omitempty" tf:"basic_auth_username,omitempty"`

	// (String)  The name of the database to use on the selected data source server. Defaults to “.
	// (Required by some data source types) The name of the database to use on the selected data source server. Defaults to “.
	DatabaseName *string `json:"databaseName,omitempty" tf:"database_name,omitempty"`

	// (Boolean) Whether to set the data source as default. This should only be true to a single data source. Defaults to false.
	// Whether to set the data source as default. This should only be `true` to a single data source. Defaults to `false`.
	IsDefault *bool `json:"isDefault,omitempty" tf:"is_default,omitempty"`

	// (String) Serialized JSON string containing the json data. This attribute can be used to pass configuration options to the data source. To figure out what options a datasource has available, see its docs or inspect the network data when saving it from the Grafana UI. Note that keys in this map are usually camelCased.
	// Serialized JSON string containing the json data. This attribute can be used to pass configuration options to the data source. To figure out what options a datasource has available, see its docs or inspect the network data when saving it from the Grafana UI. Note that keys in this map are usually camelCased.
	JSONDataEncoded *string `json:"jsonDataEncoded,omitempty" tf:"json_data_encoded,omitempty"`

	// (String) A unique name for the data source.
	// A unique name for the data source.
	Name *string `json:"name,omitempty" tf:"name,omitempty"`

	// (String) The Organization ID. If not set, the Org ID defined in the provider block will be used.
	// The Organization ID. If not set, the Org ID defined in the provider block will be used.
	// +crossplane:generate:reference:type=github.com/argannor/provider-grafana/apis/oss/v1alpha1.Organization
	// +crossplane:generate:reference:refFieldName=OrganizationRef
	// +crossplane:generate:reference:selectorFieldName=OrganizationSelector
	OrgID *string `json:"orgId,omitempty" tf:"org_id,omitempty"`

	// Reference to a Organization in oss to populate orgId.
	// +kubebuilder:validation:Optional
	OrganizationRef *v1.Reference `json:"organizationRef,omitempty" tf:"-"`

	// Selector for a Organization in oss to populate orgId.
	// +kubebuilder:validation:Optional
	OrganizationSelector *v1.Selector `json:"organizationSelector,omitempty" tf:"-"`

	// (String) The data source type. Must be one of the supported data source keywords.
	// The data source type. Must be one of the supported data source keywords.
	Type *string `json:"type,omitempty" tf:"type,omitempty"`

	// (String) Unique identifier. If unset, this will be automatically generated.
	// Unique identifier. If unset, this will be automatically generated.
	UID *string `json:"uid,omitempty" tf:"uid,omitempty"`

	// (String) The URL for the data source. The type of URL required varies depending on the chosen data source type.
	// The URL for the data source. The type of URL required varies depending on the chosen data source type.
	URL *string `json:"url,omitempty" tf:"url,omitempty"`

	// (String)  The username to use to authenticate to the data source. Defaults to “.
	// (Required by some data source types) The username to use to authenticate to the data source. Defaults to “.
	Username *string `json:"username,omitempty" tf:"username,omitempty"`
}

type DataSourceObservation struct {

	// (String) The method by which Grafana will access the data source: proxy or direct. Defaults to proxy.
	// The method by which Grafana will access the data source: `proxy` or `direct`. Defaults to `proxy`.
	AccessMode *string `json:"accessMode,omitempty" tf:"access_mode,omitempty"`

	// (Boolean) Whether to enable basic auth for the data source. Defaults to false.
	// Whether to enable basic auth for the data source. Defaults to `false`.
	BasicAuthEnabled *bool `json:"basicAuthEnabled,omitempty" tf:"basic_auth_enabled,omitempty"`

	// (String) Basic auth username. Defaults to “.
	// Basic auth username. Defaults to “.
	BasicAuthUsername *string `json:"basicAuthUsername,omitempty" tf:"basic_auth_username,omitempty"`

	// (String)  The name of the database to use on the selected data source server. Defaults to “.
	// (Required by some data source types) The name of the database to use on the selected data source server. Defaults to “.
	DatabaseName *string `json:"databaseName,omitempty" tf:"database_name,omitempty"`

	// (String) The ID of this resource.
	ID *string `json:"id,omitempty" tf:"id,omitempty"`

	// (Boolean) Whether to set the data source as default. This should only be true to a single data source. Defaults to false.
	// Whether to set the data source as default. This should only be `true` to a single data source. Defaults to `false`.
	IsDefault *bool `json:"isDefault,omitempty" tf:"is_default,omitempty"`

	// (String) Serialized JSON string containing the json data. This attribute can be used to pass configuration options to the data source. To figure out what options a datasource has available, see its docs or inspect the network data when saving it from the Grafana UI. Note that keys in this map are usually camelCased.
	// Serialized JSON string containing the json data. This attribute can be used to pass configuration options to the data source. To figure out what options a datasource has available, see its docs or inspect the network data when saving it from the Grafana UI. Note that keys in this map are usually camelCased.
	JSONDataEncoded *string `json:"jsonDataEncoded,omitempty" tf:"json_data_encoded,omitempty"`

	// (String) A unique name for the data source.
	// A unique name for the data source.
	Name *string `json:"name,omitempty" tf:"name,omitempty"`

	// (String) The Organization ID. If not set, the Org ID defined in the provider block will be used.
	// The Organization ID. If not set, the Org ID defined in the provider block will be used.
	OrgID *string `json:"orgId,omitempty" tf:"org_id,omitempty"`

	// (String) The data source type. Must be one of the supported data source keywords.
	// The data source type. Must be one of the supported data source keywords.
	Type *string `json:"type,omitempty" tf:"type,omitempty"`

	// (String) Unique identifier. If unset, this will be automatically generated.
	// Unique identifier. If unset, this will be automatically generated.
	UID *string `json:"uid,omitempty" tf:"uid,omitempty"`

	// (String) The URL for the data source. The type of URL required varies depending on the chosen data source type.
	// The URL for the data source. The type of URL required varies depending on the chosen data source type.
	URL *string `json:"url,omitempty" tf:"url,omitempty"`

	// (String)  The username to use to authenticate to the data source. Defaults to “.
	// (Required by some data source types) The username to use to authenticate to the data source. Defaults to “.
	Username *string `json:"username,omitempty" tf:"username,omitempty"`
}

type DataSourceParameters struct {

	// (String) The method by which Grafana will access the data source: proxy or direct. Defaults to proxy.
	// The method by which Grafana will access the data source: `proxy` or `direct`. Defaults to `proxy`.
	// +kubebuilder:validation:Optional
	AccessMode *string `json:"accessMode,omitempty" tf:"access_mode,omitempty"`

	// (Boolean) Whether to enable basic auth for the data source. Defaults to false.
	// Whether to enable basic auth for the data source. Defaults to `false`.
	// +kubebuilder:validation:Optional
	BasicAuthEnabled *bool `json:"basicAuthEnabled,omitempty" tf:"basic_auth_enabled,omitempty"`

	// (String) Basic auth username. Defaults to “.
	// Basic auth username. Defaults to “.
	// +kubebuilder:validation:Optional
	BasicAuthUsername *string `json:"basicAuthUsername,omitempty" tf:"basic_auth_username,omitempty"`

	// (String)  The name of the database to use on the selected data source server. Defaults to “.
	// (Required by some data source types) The name of the database to use on the selected data source server. Defaults to “.
	// +kubebuilder:validation:Optional
	DatabaseName *string `json:"databaseName,omitempty" tf:"database_name,omitempty"`

	// (Map of String, Sensitive) Custom HTTP headers
	// Custom HTTP headers
	// +kubebuilder:validation:Optional
	HTTPHeadersSecretRef *v1.SecretReference `json:"httpHeadersSecretRef,omitempty" tf:"-"`

	// (Boolean) Whether to set the data source as default. This should only be true to a single data source. Defaults to false.
	// Whether to set the data source as default. This should only be `true` to a single data source. Defaults to `false`.
	// +kubebuilder:validation:Optional
	IsDefault *bool `json:"isDefault,omitempty" tf:"is_default,omitempty"`

	// (String) Serialized JSON string containing the json data. This attribute can be used to pass configuration options to the data source. To figure out what options a datasource has available, see its docs or inspect the network data when saving it from the Grafana UI. Note that keys in this map are usually camelCased.
	// Serialized JSON string containing the json data. This attribute can be used to pass configuration options to the data source. To figure out what options a datasource has available, see its docs or inspect the network data when saving it from the Grafana UI. Note that keys in this map are usually camelCased.
	// +kubebuilder:validation:Optional
	JSONDataEncoded *string `json:"jsonDataEncoded,omitempty" tf:"json_data_encoded,omitempty"`

	// (String) A unique name for the data source.
	// A unique name for the data source.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Name is immutable"
	Name *string `json:"name,omitempty" tf:"name,omitempty"`

	// (String) The Organization ID. If not set, the Org ID defined in the provider block will be used.
	// The Organization ID. If not set, the Org ID defined in the provider block will be used.
	// +crossplane:generate:reference:type=github.com/argannor/provider-grafana/apis/oss/v1alpha1.Organization
	// +crossplane:generate:reference:refFieldName=OrganizationRef
	// +crossplane:generate:reference:selectorFieldName=OrganizationSelector
	// +crossplane:generate:reference:extractor=github.com/argannor/provider-grafana/apis/oss/v1alpha1.OrgId()
	// +kubebuilder:validation:Optional
	OrgID *string `json:"orgId,omitempty" tf:"org_id,omitempty"`

	// Reference to a Organization in oss to populate orgId.
	// +kubebuilder:validation:Optional
	OrganizationRef *v1.Reference `json:"organizationRef,omitempty" tf:"-"`

	// Selector for a Organization in oss to populate orgId.
	// +kubebuilder:validation:Optional
	OrganizationSelector *v1.Selector `json:"organizationSelector,omitempty" tf:"-"`

	// (String, Sensitive) Serialized JSON string containing the secure json data. This attribute can be used to pass secure configuration options to the data source. To figure out what options a datasource has available, see its docs or inspect the network data when saving it from the Grafana UI. Note that keys in this map are usually camelCased.
	// Serialized JSON string containing the secure json data. This attribute can be used to pass secure configuration options to the data source. To figure out what options a datasource has available, see its docs or inspect the network data when saving it from the Grafana UI. Note that keys in this map are usually camelCased.
	// +kubebuilder:validation:Optional
	SecureJSONDataEncodedSecretRef *v1.SecretKeySelector `json:"secureJsonDataEncodedSecretRef,omitempty" tf:"-"`

	// (String) The data source type. Must be one of the supported data source keywords.
	// The data source type. Must be one of the supported data source keywords.
	// +kubebuilder:validation:Optional
	Type *string `json:"type,omitempty" tf:"type,omitempty"`

	// (String) Unique identifier. If unset, this will be automatically generated.
	// Unique identifier. If unset, this will be automatically generated.
	// +kubebuilder:validation:Optional
	UID *string `json:"uid,omitempty" tf:"uid,omitempty"`

	// (String) The URL for the data source. The type of URL required varies depending on the chosen data source type.
	// The URL for the data source. The type of URL required varies depending on the chosen data source type.
	// +kubebuilder:validation:Optional
	URL *string `json:"url,omitempty" tf:"url,omitempty"`

	// (String)  The username to use to authenticate to the data source. Defaults to “.
	// (Required by some data source types) The username to use to authenticate to the data source. Defaults to “.
	// +kubebuilder:validation:Optional
	Username *string `json:"username,omitempty" tf:"username,omitempty"`
}

// DataSourceSpec defines the desired state of DataSource
type DataSourceSpec struct {
	v1.ResourceSpec `json:",inline"`
	ForProvider     DataSourceParameters `json:"forProvider"`
	// THIS IS A BETA FIELD. It will be honored
	// unless the Management Policies feature flag is disabled.
	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// The same fields are also added to the terraform ignore_changes hook, to
	// avoid updating them after creation. This is useful for fields that are
	// required on creation, but we do not desire to update them after creation,
	// for example because of an external controller is managing them, like an
	// autoscaler.
	InitProvider DataSourceInitParameters `json:"initProvider,omitempty"`
}

// DataSourceStatus defines the observed state of DataSource.
type DataSourceStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        DataSourceObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// DataSource is the Schema for the DataSources API. Official documentation https://grafana.com/docs/grafana/latest/datasources/HTTP API https://grafana.com/docs/grafana/latest/developers/http_api/data_source/ The required arguments for this resource vary depending on the type of data source selected (via the 'type' argument).
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,grafana}
type DataSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.name) || (has(self.initProvider) && has(self.initProvider.name))",message="spec.forProvider.name is a required parameter"
	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.type) || (has(self.initProvider) && has(self.initProvider.type))",message="spec.forProvider.type is a required parameter"
	Spec   DataSourceSpec   `json:"spec"`
	Status DataSourceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DataSourceList contains a list of DataSources
type DataSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataSource `json:"items"`
}

// DataSource type metadata.
var (
	DataSourceKind             = reflect.TypeOf(DataSource{}).Name()
	DataSourceGroupKind        = schema.GroupKind{Group: Group, Kind: DataSourceKind}.String()
	DataSourceKindAPIVersion   = DataSourceKind + "." + SchemeGroupVersion.String()
	DataSourceGroupVersionKind = SchemeGroupVersion.WithKind(DataSourceKind)
)

func init() {
	SchemeBuilder.Register(&DataSource{}, &DataSourceList{})
}
