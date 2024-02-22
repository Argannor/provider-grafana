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

type FolderInitParameters struct {

	// Reference to a Folder in oss to populate parentFolderUid.
	// +kubebuilder:validation:Optional
	FolderRef *v1.Reference `json:"folderRef,omitempty" tf:"-"`

	// Selector for a Folder in oss to populate parentFolderUid.
	// +kubebuilder:validation:Optional
	FolderSelector *v1.Selector `json:"folderSelector,omitempty" tf:"-"`

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

	// (String) The uid of the parent folder. If set, the folder will be nested. If not set, the folder will be created in the root folder. Note: This requires the nestedFolders feature flag to be enabled on your Grafana instance.
	// The uid of the parent folder. If set, the folder will be nested. If not set, the folder will be created in the root folder. Note: This requires the nestedFolders feature flag to be enabled on your Grafana instance.
	// +crossplane:generate:reference:type=github.com/argannor/provider-grafana/apis/oss/v1alpha1.Folder
	// +crossplane:generate:reference:extractor=github.com/argannor/provider-grafana/apis/oss/v1alpha1.UIDExtractor()
	// +crossplane:generate:reference:refFieldName=FolderRef
	// +crossplane:generate:reference:selectorFieldName=FolderSelector
	ParentFolderUID *string `json:"parentFolderUid,omitempty" tf:"parent_folder_uid,omitempty"`

	// (String) The title of the folder.
	// The title of the folder.
	Title *string `json:"title,omitempty" tf:"title,omitempty"`

	// (String) Unique identifier.
	// Unique identifier.
	UID *string `json:"uid,omitempty" tf:"uid,omitempty"`
}

type FolderObservation struct {

	// (String) The ID of this resource.
	ID *string `json:"id,omitempty" tf:"id,omitempty"`

	// (String) The Organization ID. If not set, the Org ID defined in the provider block will be used.
	// The Organization ID. If not set, the Org ID defined in the provider block will be used.
	OrgID *string `json:"orgId,omitempty" tf:"org_id,omitempty"`

	// (String) The uid of the parent folder. If set, the folder will be nested. If not set, the folder will be created in the root folder. Note: This requires the nestedFolders feature flag to be enabled on your Grafana instance.
	// The uid of the parent folder. If set, the folder will be nested. If not set, the folder will be created in the root folder. Note: This requires the nestedFolders feature flag to be enabled on your Grafana instance.
	ParentFolderUID *string `json:"parentFolderUid,omitempty" tf:"parent_folder_uid,omitempty"`

	// (String) The title of the folder.
	// The title of the folder.
	Title *string `json:"title,omitempty" tf:"title,omitempty"`

	// (String) Unique identifier.
	// Unique identifier.
	UID *string `json:"uid,omitempty" tf:"uid,omitempty"`

	// (String) The full URL of the folder.
	// The full URL of the folder.
	URL *string `json:"url,omitempty" tf:"url,omitempty"`

	// (Int64) The version of the folder.
	// The version of the folder.
	Version *int64 `json:"version,omitempty" tf:"version,omitempty"`
}

type FolderParameters struct {

	// Reference to a Folder in oss to populate parentFolderUid.
	// +kubebuilder:validation:Optional
	FolderRef *v1.Reference `json:"folderRef,omitempty" tf:"-"`

	// Selector for a Folder in oss to populate parentFolderUid.
	// +kubebuilder:validation:Optional
	FolderSelector *v1.Selector `json:"folderSelector,omitempty" tf:"-"`

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

	// (String) The uid of the parent folder. If set, the folder will be nested. If not set, the folder will be created in the root folder. Note: This requires the nestedFolders feature flag to be enabled on your Grafana instance.
	// The uid of the parent folder. If set, the folder will be nested. If not set, the folder will be created in the root folder. Note: This requires the nestedFolders feature flag to be enabled on your Grafana instance.
	// +crossplane:generate:reference:type=github.com/argannor/provider-grafana/apis/oss/v1alpha1.Folder
	// +crossplane:generate:reference:extractor=github.com/argannor/provider-grafana/apis/oss/v1alpha1.UIDExtractor()
	// +crossplane:generate:reference:refFieldName=FolderRef
	// +crossplane:generate:reference:selectorFieldName=FolderSelector
	// +kubebuilder:validation:Optional
	ParentFolderUID *string `json:"parentFolderUid,omitempty" tf:"parent_folder_uid,omitempty"`

	// (String) The title of the folder.
	// The title of the folder.
	// +kubebuilder:validation:Optional
	Title *string `json:"title,omitempty" tf:"title,omitempty"`

	// (String) Unique identifier.
	// Unique identifier.
	// +kubebuilder:validation:Optional
	UID *string `json:"uid,omitempty" tf:"uid,omitempty"`
}

// FolderSpec defines the desired state of Folder
type FolderSpec struct {
	v1.ResourceSpec `json:",inline"`
	ForProvider     FolderParameters `json:"forProvider"`
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
	InitProvider FolderInitParameters `json:"initProvider,omitempty"`
}

// FolderStatus defines the observed state of Folder.
type FolderStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        FolderObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// Folder is the Schema for the Folders API. Official documentation https://grafana.com/docs/grafana/latest/folders/manage-folders/HTTP API https://grafana.com/docs/grafana/latest/developers/http_api/folder/
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,grafana}
type Folder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.title) || (has(self.initProvider) && has(self.initProvider.title))",message="spec.forProvider.title is a required parameter"
	Spec   FolderSpec   `json:"spec"`
	Status FolderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FolderList contains a list of Folders
type FolderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Folder `json:"items"`
}

// DataSource type metadata.
var (
	FolderKind             = reflect.TypeOf(DataSource{}).Name()
	FolderGroupKind        = schema.GroupKind{Group: Group, Kind: FolderKind}.String()
	FolderKindAPIVersion   = FolderKind + "." + SchemeGroupVersion.String()
	FolderGroupVersionKind = SchemeGroupVersion.WithKind(FolderKind)
)

func init() {
	SchemeBuilder.Register(&Folder{}, &FolderList{})
}
