// SPDX-FileCopyrightText: 2023 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

/*
Copyright 2022 Upbound Inc.
*/

package v1alpha1

import (
	"reflect"

	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

type DashboardInitParameters struct {

	// (String) The complete dashboard model JSON.
	// The complete dashboard model JSON.
	ConfigJSON *string `json:"configJson,omitempty" tf:"config_json,omitempty"`

	// (String) The id or UID of the folder to save the dashboard in.
	// The id or UID of the folder to save the dashboard in.
	// +crossplane:generate:reference:type=github.com/argannor/provider-grafana/apis/oss/v1alpha1.Folder
	// +crossplane:generate:reference:extractor=github.com/argannor/provider-grafana/apis/oss/v1alpha1.UIDExtractor()
	// +crossplane:generate:reference:refFieldName=FolderRef
	// +crossplane:generate:reference:selectorFieldName=FolderSelector
	Folder *string `json:"folder,omitempty" tf:"folder,omitempty"`

	// Reference to a Folder in oss to populate folder.
	// +kubebuilder:validation:Optional
	FolderRef *v1.Reference `json:"folderRef,omitempty" tf:"-"`

	// Selector for a Folder in oss to populate folder.
	// +kubebuilder:validation:Optional
	FolderSelector *v1.Selector `json:"folderSelector,omitempty" tf:"-"`

	// (String) Set a commit message for the version history.
	// Set a commit message for the version history.
	Message *string `json:"message,omitempty" tf:"message,omitempty"`

	// (String) The Organization ID. If not set, the Org ID defined in the provider block will be used.
	// The Organization ID. If not set, the Org ID defined in the provider block will be used.
	// +crossplane:generate:reference:type=github.com/argannor/provider-grafana/apis/oss/v1alpha1.Organization
	// +crossplane:generate:reference:refFieldName=OrganizationRef
	// +crossplane:generate:reference:selectorFieldName=OrganizationSelector
	// +crossplane:generate:reference:extractor=github.com/argannor/provider-grafana/apis/oss/v1alpha1.OrgId()
	OrgID *string `json:"orgId,omitempty" tf:"org_id,omitempty"`

	// Reference to a Organization in oss to populate orgId.
	// +kubebuilder:validation:Optional
	OrganizationRef *v1.Reference `json:"organizationRef,omitempty" tf:"-"`

	// Selector for a Organization in oss to populate orgId.
	// +kubebuilder:validation:Optional
	OrganizationSelector *v1.Selector `json:"organizationSelector,omitempty" tf:"-"`

	// (Boolean) Set to true if you want to overwrite existing dashboard with newer version, same dashboard title in folder or same dashboard uid.
	// Set to true if you want to overwrite existing dashboard with newer version, same dashboard title in folder or same dashboard uid.
	Overwrite *bool `json:"overwrite,omitempty" tf:"overwrite,omitempty"`
}

type DashboardObservation struct {

	// (String) The complete dashboard model JSON.
	// The complete dashboard model JSON.
	ConfigJSON *string `json:"configJson,omitempty" tf:"config_json,omitempty"`

	// (Number) The numeric ID of the dashboard computed by Grafana.
	// The numeric ID of the dashboard computed by Grafana.
	DashboardID *int64 `json:"dashboardId,omitempty" tf:"dashboard_id,omitempty"`

	// (String) The id or UID of the folder to save the dashboard in.
	// The id or UID of the folder to save the dashboard in.
	Folder *string `json:"folder,omitempty" tf:"folder,omitempty"`

	// (String) The ID of this resource.
	ID *string `json:"id,omitempty" tf:"id,omitempty"`

	// (String) Set a commit message for the version history.
	// Set a commit message for the version history.
	Message *string `json:"message,omitempty" tf:"message,omitempty"`

	// (String) The Organization ID. If not set, the Org ID defined in the provider block will be used.
	// The Organization ID. If not set, the Org ID defined in the provider block will be used.
	OrgID *string `json:"orgId,omitempty" tf:"org_id,omitempty"`

	// (Boolean) Set to true if you want to overwrite existing dashboard with newer version, same dashboard title in folder or same dashboard uid.
	// Set to true if you want to overwrite existing dashboard with newer version, same dashboard title in folder or same dashboard uid.
	Overwrite *bool `json:"overwrite,omitempty" tf:"overwrite,omitempty"`

	// (String) The unique identifier of a dashboard. This is used to construct its URL. It's automatically generated if not provided when creating a dashboard. The uid allows having consistent URLs for accessing dashboards and when syncing dashboards between multiple Grafana installs.
	// The unique identifier of a dashboard. This is used to construct its URL. It's automatically generated if not provided when creating a dashboard. The uid allows having consistent URLs for accessing dashboards and when syncing dashboards between multiple Grafana installs.
	UID *string `json:"uid,omitempty" tf:"uid,omitempty"`

	// (String) The full URL of the dashboard.
	// The full URL of the dashboard.
	URL *string `json:"url,omitempty" tf:"url,omitempty"`

	// (Number) Whenever you save a version of your dashboard, a copy of that version is saved so that previous versions of your dashboard are not lost.
	// Whenever you save a version of your dashboard, a copy of that version is saved so that previous versions of your dashboard are not lost.
	Version *int64 `json:"version,omitempty" tf:"version,omitempty"`

	// (Number) The version that was returned by Grafana on the last write operation carried out by this provider.
	ManagedVersion *int64 `json:"managedVersion,omitempty" tf:"managed_version,omitempty"`
}

type DashboardParameters struct {

	// (String) The complete dashboard model JSON.
	// The complete dashboard model JSON.
	// +kubebuilder:validation:Optional
	ConfigJSON *string `json:"configJson,omitempty" tf:"config_json,omitempty"`

	// (String) The id or UID of the folder to save the dashboard in.
	// The id or UID of the folder to save the dashboard in.
	// +crossplane:generate:reference:type=github.com/argannor/provider-grafana/apis/oss/v1alpha1.Folder
	// +crossplane:generate:reference:extractor=github.com/argannor/provider-grafana/apis/oss/v1alpha1.UIDExtractor()
	// +crossplane:generate:reference:refFieldName=FolderRef
	// +crossplane:generate:reference:selectorFieldName=FolderSelector
	// +kubebuilder:validation:Optional
	Folder *string `json:"folder,omitempty" tf:"folder,omitempty"`

	// Reference to a Folder in oss to populate folder.
	// +kubebuilder:validation:Optional
	FolderRef *v1.Reference `json:"folderRef,omitempty" tf:"-"`

	// Selector for a Folder in oss to populate folder.
	// +kubebuilder:validation:Optional
	FolderSelector *v1.Selector `json:"folderSelector,omitempty" tf:"-"`

	// (String) Set a commit message for the version history.
	// Set a commit message for the version history.
	// +kubebuilder:validation:Optional
	Message *string `json:"message,omitempty" tf:"message,omitempty"`

	// (String) The Organization ID. If not set, the Org ID defined in the provider block will be used.
	// The Organization ID. If not set, the Org ID defined in the provider block will be used.
	// +crossplane:generate:reference:type=github.com/argannor/provider-grafana/apis/oss/v1alpha1.Organization
	// +crossplane:generate:reference:refFieldName=OrganizationRef
	// +crossplane:generate:reference:selectorFieldName=OrganizationSelector
	// +crossplane:generate:reference:extractor=github.com/argannor/provider-grafana/apis/oss/v1alpha1.OrgId()
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Name is immutable"
	// +kubebuilder:validation:Optional
	OrgID *string `json:"orgId,omitempty" tf:"org_id,omitempty"`

	// Reference to a Organization in oss to populate orgId.
	// +kubebuilder:validation:Optional
	OrganizationRef *v1.Reference `json:"organizationRef,omitempty" tf:"-"`

	// Selector for a Organization in oss to populate orgId.
	// +kubebuilder:validation:Optional
	OrganizationSelector *v1.Selector `json:"organizationSelector,omitempty" tf:"-"`

	// (Boolean) Set to true if you want to overwrite existing dashboard with newer version, same dashboard title in folder or same dashboard uid.
	// Set to true if you want to overwrite existing dashboard with newer version, same dashboard title in folder or same dashboard uid.
	// +kubebuilder:validation:Optional
	Overwrite *bool `json:"overwrite,omitempty" tf:"overwrite,omitempty"`
}

// DashboardSpec defines the desired state of Dashboard
type DashboardSpec struct {
	v1.ResourceSpec `json:",inline"`
	ForProvider     DashboardParameters `json:"forProvider"`
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
	InitProvider DashboardInitParameters `json:"initProvider,omitempty"`
}

// DashboardStatus defines the observed state of Dashboard.
type DashboardStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        DashboardObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// Dashboard is the Schema for the Dashboards API. Manages Grafana dashboards. Official documentation https://grafana.com/docs/grafana/latest/dashboards/HTTP API https://grafana.com/docs/grafana/latest/developers/http_api/dashboard/
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,grafana}
type Dashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.configJson) || (has(self.initProvider) && has(self.initProvider.configJson))",message="spec.forProvider.configJson is a required parameter"
	Spec   DashboardSpec   `json:"spec"`
	Status DashboardStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DashboardList contains a list of Dashboards
type DashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dashboard `json:"items"`
}

// Dashboard type metadata.
var (
	DashboardKind             = reflect.TypeOf(Dashboard{}).Name()
	DashboardGroupKind        = schema.GroupKind{Group: Group, Kind: DashboardKind}.String()
	DashboardKindAPIVersion   = DashboardKind + "." + SchemeGroupVersion.String()
	DashboardGroupVersionKind = SchemeGroupVersion.WithKind(DashboardKind)
)

func init() {
	SchemeBuilder.Register(&Dashboard{}, &DashboardList{})
}

func UIDExtractor() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		paved, err := fieldpath.PaveObject(mg)
		if err != nil {
			return ""
		}
		r, err := paved.GetString("spec.forProvider.uid")
		if err == nil && r != "" {
			return r
		}
		// UID is optional, so it can be in atProvider if it's not in forProvider
		r, err = paved.GetString("status.atProvider.uid")
		if err != nil {
			return ""
		}
		return r
	}
}
