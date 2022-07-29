// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1beta1

import (
	"github.com/c2h5oh/datasize"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

// FlowRequirements include the requirements specific to the flow
// Note: Implicit copies done for data plane optimization by Fybrik do not use these parameters
type FlowRequirements struct {
	// Catalog indicates that the data asset must be cataloged, and in which catalog to register it
	// +optional
	Catalog string `json:"catalog,omitempty"`

	// Storage estimate indicates the estimated amount of storage in MB, GB, TB required when writing new data.
	// +optional
	StorageEstimate datasize.ByteSize `json:"storageEstimate,omitempty"`

	// IsNewDataSet if true indicates that the DataContext.DataSetID is user provided and not a full catalog / dataset ID.
	// Relevant when writing.
	// A unique ID from the catalog will be provided in the FybrikApplication Status after a new catalog entry is created.
	// +optional
	IsNewDataSet bool `json:"isNewDataSet,omitempty"`

	// Source asset metadata like asset name, owner, geography, etc
	// Relevant when writing new asset.
	// +optional
	ResourceMetadata *datacatalog.ResourceMetadata `json:"metadata,omitempty"`
}

// DataRequirements structure contains a list of requirements (interface, need to catalog the dataset, etc.)
type DataRequirements struct {
	// Interface indicates the protocol and format expected by the data user
	// +optional
	Interface *taxonomy.Interface `json:"interface,omitempty"`

	// FlowParams include the requirements for particular data flows
	// +optional
	FlowParams FlowRequirements `json:"flowParams,omitempty"`
}

// DataContext indicates data set being processed by the workload
// and includes information about the data format and technologies used to access the data.
type DataContext struct {
	// DataSetID is a unique identifier of the dataset chosen from the data catalog.
	// For data catalogs that support multiple sub-catalogs, it includes the catalog id and the dataset id.
	// When writing a new dataset it is the name provided by the user or workload generating it.
	// +required
	// +kubebuilder:validation:MinLength=1
	DataSetID string `json:"dataSetID"`

	// Flows indicates what is being done with the particular dataset - ex: read, write, copy (ingest), delete
	// This is optional for the purpose of backward compatibility.
	// If nothing is provided, read is assumed.
	// +optional
	Flow taxonomy.DataFlow `json:"flow,omitempty"`

	// Requirements from the system
	// +required
	Requirements DataRequirements `json:"requirements"`
}

// FybrikApplicationSpec defines data flows needed by the application, the purpose and other contextual information about the application.
// +fybrik:validation:object="fybrik_application"
type FybrikApplicationSpec struct {

	// Selector enables to connect the resource to the application
	// Application labels should match the labels in the selector.
	// +optional
	Selector Selector `json:"selector"`

	// SecretRef points to the secret that holds credentials for each system the user has been authenticated with.
	// The secret is deployed in FybrikApplication namespace.
	// +optional
	SecretRef string `json:"secretRef,omitempty"`

	// AppInfo contains information describing the reasons for the processing
	// that will be done by the application.
	// +required
	AppInfo taxonomy.AppInfo `json:"appInfo"`

	// Data contains the identifiers of the data to be used by the Data Scientist's application,
	// and the protocol used to access it and the format expected.
	// +required
	Data []DataContext `json:"data"`
}

// ErrorMessages that are reported to the user
const (
	InvalidAssetID              string = "the asset does not exist"
	ReadAccessDenied            string = "governance policies forbid access to the data"
	CopyNotAllowed              string = "copy of the data is required but can not be done according to the governance policies"
	WriteNotAllowed             string = "governance policies forbid writing of the data"
	StorageAccountUndefined     string = "no storage account has been defined"
	ModuleNotFound              string = "no module has been registered"
	InsufficientStorage         string = "no bucket was provisioned for implicit copy"
	InvalidClusterConfiguration string = "cluster configuration does not support the requirements"
	InvalidAssetDataStore       string = "the asset data store is not supported"
)

// ResourceReference contains resource identifier(name, namespace, kind)
type ResourceReference struct {
	// Name of the resource
	Name string `json:"name"`
	// Namespace of the resource
	Namespace string `json:"namespace"`
	// Kind of the resource (Blueprint, Plotter)
	Kind string `json:"kind"`
	// Version of FybrikApplication that has generated this resource
	AppVersion int64 `json:"appVersion"`
}

// SecretRef contains the details of a secret
type SecretRef struct {
	// Secret Namespace
	// +required
	Namespace string `json:"namespace"`
	// Secret name
	// +required
	Name string `json:"name"`
}

// DatasetDetails holds details of the provisioned storage
type DatasetDetails struct {
	// Reference to a Dataset resource containing the request to provision storage
	// +optional
	DatasetRef string `json:"datasetRef,omitempty"`

	// Reference to a secret where the credentials are stored
	// +optional
	SecretRef SecretRef `json:"secretRef,omitempty"`

	// Dataset information
	// +optional
	Details *DataStore `json:"details,omitempty"`

	// Resource Metadata
	// +optional
	ResourceMetadata *datacatalog.ResourceMetadata `json:"resourceMetadata,omitempty"`
}

// AssetState defines the observed state of an asset
type AssetState struct {
	// Conditions indicate the asset state (Ready, Deny, Error)
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`

	// CatalogedAsset provides a new asset identifier after being registered in the enterprise catalog
	// +optional
	CatalogedAsset string `json:"catalogedAsset,omitempty"`

	// Endpoint provides the endpoint spec from which the asset will be served to the application
	// +optional
	Endpoint taxonomy.Connection `json:"endpoint,omitempty"`
}

// FybrikApplicationStatus defines the observed state of FybrikApplication.
type FybrikApplicationStatus struct {
	// Ready is true if all specified assets are either ready to be used or are denied access.
	// +optional
	Ready bool `json:"ready,omitempty"`

	// ErrorMessage indicates that an error has happened during the reconcile, unrelated to a specific asset
	// +optional
	ErrorMessage string `json:"errorMessage,omitempty"`

	// AssetStates provides a status per asset
	// +optional
	AssetStates map[string]AssetState `json:"assetStates,omitempty"`

	// ObservedGeneration is taken from the FybrikApplication metadata.  This is used to determine during reconcile
	// whether reconcile was called because the desired state changed, or whether the Blueprint status changed.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// ValidatedGeneration is the version of the FyrbikApplication that has been validated with the taxonomy defined.
	// +optional
	ValidatedGeneration int64 `json:"validatedGeneration,omitempty"`

	// ValidApplication indicates whether the FybrikApplication is valid given the defined taxonomy
	// +optional
	ValidApplication corev1.ConditionStatus `json:"validApplication,omitempty"`

	// Generated resource identifier
	// +optional
	Generated *ResourceReference `json:"generated,omitempty"`

	// ProvisionedStorage maps a dataset (identified by AssetID) to the new provisioned bucket.
	// It allows FybrikApplication controller to manage buckets in case the spec has been modified, an error has occurred,
	// or a delete event has been received.
	// ProvisionedStorage has the information required to register the dataset once the owned plotter resource is ready
	// +optional
	ProvisionedStorage map[string]DatasetDetails `json:"provisionedStorage,omitempty"`
}

// FybrikApplication provides information about the application whose data is being operated on,
// the nature of the processing, and the data sets chosen for processing by the application.
// The FybrikApplication controller obtains instructions regarding any governance related changes that must
// be performed on the data, identifies the modules capable of performing such changes, and finally
// generates the Plotter which defines the secure runtime environment and all the components
// in it.  This runtime environment provides the application with access to the data requested
// in a secure manner and without having to provide any credentials for the data sets.  The credentials are obtained automatically
// by the manager from the credential management system.
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
type FybrikApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   FybrikApplicationSpec   `json:"spec"`
	Status FybrikApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FybrikApplicationList contains a list of FybrikApplication
type FybrikApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FybrikApplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FybrikApplication{}, &FybrikApplicationList{})
}

const (
	ApplicationClusterLabel   = "app.fybrik.io/app-cluster"
	ApplicationNamespaceLabel = "app.fybrik.io/app-namespace"
	ApplicationNameLabel      = "app.fybrik.io/app-name"
)
