// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CatalogRequirements contain the specifics for catalogging the data asset
type CatalogRequirements struct {
	// CatalogService specifies the datacatalog service that will be used for catalogging the data into.
	// +optional
	CatalogService string `json:"service,omitempty"`

	// CatalogID specifies the catalog where the data will be cataloged.
	// +optional
	CatalogID string `json:"catalogID,omitempty"`
}

// CopyRequirements include the requirements for the data copy operation
type CopyRequirements struct {
	// Required indicates that the data must be copied.
	// +optional
	Required bool `json:"required,omitempty"`

	// Catalog indicates that the data asset must be cataloged.
	// +optional
	Catalog CatalogRequirements `json:"catalog,omitempty"`
}

// DataRequirements structure contains a list of requirements (interface, need to catalog the dataset, etc.)
type DataRequirements struct {
	// Interface indicates the protocol and format expected by the data user
	// +required
	Interface InterfaceDetails `json:"interface"`

	// CopyRequrements include the requirements for copying the data
	// +optional
	Copy CopyRequirements `json:"copy,omitempty"`
}

// DataContext indicates data set chosen by the Data Scientist to be used by his application,
// and includes information about the data format and technologies used by the application
// to access the data.
type DataContext struct {
	// DataSetID is a unique identifier of the dataset chosen from the data catalog for processing by the data user application.
	// +required
	// +kubebuilder:validation:MinLength=1
	DataSetID string `json:"dataSetID"`

	// CatalogService represents the catalog service for accessing the requested dataset.
	// If not specified, the enterprise catalog service will be used.
	// +optional
	CatalogService string `json:"catalogService,omitempty"`
	// Requirements from the system
	// +required
	Requirements DataRequirements `json:"requirements"`
}

// ApplicationDetails provides information about the Data Scientist's application, which is deployed separately.
// The information provided is used to determine if the data should be altered in any way prior to its use,
// based on policies and rules defined in an external data policy manager.
type ApplicationDetails map[string]string

// FybrikApplicationSpec defines data flows needed by the application, the purpose and other contextual information about the application.
// Read flow - if selector is populated, fybrik builds a data plane for reading the specified data sets
// Ingest flow - if no selector, and data/copy/required is true then the data specified is copied into a bucket allocated by fybrik and is cataloged in the data catalog
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
	AppInfo ApplicationDetails `json:"appInfo"`

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

// DatasetDetails holds details of the provisioned storage
type DatasetDetails struct {
	// Reference to a Dataset resource containing the request to provision storage
	DatasetRef string `json:"datasetRef,omitempty"`
	// Reference to a secret where the credentials are stored
	SecretRef string `json:"secretRef,omitempty"`
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
	Endpoint EndpointSpec `json:"endpoint,omitempty"`
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
	// It allows FybrikApplication controller to manage buckets in case the spec has been modified, an error has occurred, or a delete event has been received.
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
// +kubebuilder:subresource:status
type FybrikApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FybrikApplicationSpec   `json:"spec,omitempty"`
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
