// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"fybrik.io/fybrik/pkg/serde"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CopyModuleArgs define the input parameters for modules that copy data from location A to location B
// Credentials are stored in a credential management system such as vault
type CopyModuleArgs struct {

	// Source is the where the data currently resides
	// +required
	Source DataStore `json:"source"`

	// Destination is the data store to which the data will be copied
	// +required
	Destination DataStore `json:"destination"`

	// Transformations are different types of processing that may be done to the data as it is copied.
	// +optional
	Transformations []serde.Arbitrary `json:"transformations,omitempty"`
}

// ReadModuleArgs define the input parameters for modules that read data from location A
type ReadModuleArgs struct {
	// Source of the read path module
	// +required
	Source DataStore `json:"source"`

	// AssetID identifies the asset to be used for accessing the data when it is ready
	// It is copied from the FybrikApplication resource
	// +required
	AssetID string `json:"assetID"`

	// Transformations are different types of processing that may be done to the data
	// +optional
	Transformations []serde.Arbitrary `json:"transformations,omitempty"`
}

// WriteModuleArgs define the input parameters for modules that write data to location B
type WriteModuleArgs struct {
	// Destination is the data store to which the data will be written
	// +required
	Destination DataStore `json:"destination"`

	// Transformations are different types of processing that may be done to the data as it is written.
	// +optional
	Transformations []serde.Arbitrary `json:"transformations,omitempty"`
}

// ModuleArguments are the parameters passed to a component that runs in the data path
// In the future might support output args as well
// The arguments passed depend on the type of module
type ModuleArguments struct {
	// CopyArgs are parameters specific to modules that copy data from one data store to another.
	// +optional
	Copy *CopyModuleArgs `json:"copy,omitempty"`

	// ReadArgs are parameters that are specific to modules that enable an application to read data
	// +optional
	Read []ReadModuleArgs `json:"read,omitempty"`

	// WriteArgs are parameters that are specific to modules that enable an application to write data
	// +optional
	Write []WriteModuleArgs `json:"write,omitempty"`
}

// CapabilityAPI enables the separation of APIs by capability, since a single module may
// be able to perform multiple capabilities. (ex: read, write, ...).
// Future: Capabilities are defined in the taxonomy.
type CapabilityAPI struct {
	// Capability indicates the capability associated with the arguments and API (ex: read, write, ...)
	// +required
	Capability string `json:"capability"`

	// API indicates to the application how to access the module for the given capability
	// +required
	API ModuleAPI `json:"api,omitempty"`
}

// BlueprintModule is a copy of a FybrikModule Custom Resource.  It contains the information necessary
// to instantiate a datapath component, including the parameters relevant for the particular workload.
type BlueprintModule struct {
	// Name of the fybrikmodule on which this is based
	// +required
	Name string `json:"name"`

	// InstanceName is the unique name for the deployed instance related to this workload
	// +required
	InstanceName string `json:"instanceName"`

	// ServiceModule is the name of the service being configured by this module.
	// Only used if this module is configuring another model rather than performing the function on its own
	// +optional
	ServiceModule string `json:"serviceModule"`

	// Chart contains the location of the helm chart with info detailing how to deploy
	// +required
	Chart ChartSpec `json:"chart"`

	// Arguments define the API used for external components and the workload to access a module's capabilities,
	// and the parameters passed to it for a given workload.
	// +optional
	APIs []CapabilityAPI `json:"apis,omitempty"`

	// Arguments are the input parameters for a specific instance of a module.
	// +optional
	Arguments ModuleArguments `json:"arguments,omitempty"`
}

// BlueprintSpec defines the desired state of Blueprint, which defines the components of the workload's data path
// that run in a particular cluster.  In a single cluster environment there is one blueprint.  In a multi-cluster
// environment there is one Blueprint per cluster per workload (FybrikApplication).
type BlueprintSpec struct {
	// Cluster indicates the cluster on which the Blueprint runs
	// +required
	Cluster string `json:"cluster"`

	// Modules is a list of modules that indicate the data path components that run in this cluster
	// +required
	Modules []BlueprintModule `json:"modules"`
}

// BlueprintStatus defines the observed state of Blueprint
// This includes readiness, error message, and indicators forthe Kubernetes
// resources owned by the Blueprint for cleanup and status monitoring
type BlueprintStatus struct {
	// ObservedState includes information to be reported back to the FybrikApplication resource
	// It includes readiness and error indications, as well as user instructions
	// +optional
	ObservedState ObservedState `json:"observedState,omitempty"`

	// ObservedGeneration is taken from the Blueprint metadata.  This is used to determine during reconcile
	// whether reconcile was called because the desired state changed, or whether status of the allocated resources should be checked.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Releases map each release to the observed generation of the blueprint containing this release.
	// At the end of reconcile, each release should be mapped to the latest blueprint version or be uninstalled.
	// +optional
	Releases map[string]int64 `json:"releases,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.observedState.ready`

// Blueprint is the Schema for the blueprints API
type Blueprint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BlueprintSpec   `json:"spec,omitempty"`
	Status BlueprintStatus `json:"status,omitempty"`
}

// MetaBlueprint defines blueprint metadata (name, namespace) and status
type MetaBlueprint struct {
	// +required
	Name string `json:"name"`

	// +required
	Namespace string `json:"namespace"`

	// +required
	Status BlueprintStatus `json:"status"`
}

// +kubebuilder:object:root=true

// BlueprintList contains a list of Blueprint
type BlueprintList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Blueprint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Blueprint{}, &BlueprintList{})
}

// CreateMetaBlueprint creates MetaBlueprint structure of the given blueprint
func CreateMetaBlueprint(blueprint *Blueprint) MetaBlueprint {
	metaBlueprint := MetaBlueprint{
		Name:      blueprint.GetName(),
		Namespace: blueprint.GetNamespace(),
		Status:    blueprint.Status,
	}
	return metaBlueprint
}

// CreateMetaBlueprintWithoutState creates the MetaBlueprint structure with an empty state
func CreateMetaBlueprintWithoutState(blueprint *Blueprint) MetaBlueprint {
	metaBlueprint := MetaBlueprint{
		Name:      blueprint.GetName(),
		Namespace: blueprint.GetNamespace(),
		Status:    BlueprintStatus{},
	}
	return metaBlueprint
}

const (
	BlueprintNamespaceLabel = "app.fybrik.io/blueprintNamespace"
	BlueprintNameLabel      = "app.fybrik.io/blueprintName"
)
