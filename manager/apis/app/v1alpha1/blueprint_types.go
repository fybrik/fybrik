// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"github.com/mesh-for-data/mesh-for-data/pkg/serde"
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
	// It is copied from the M4DApplication resource
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

// FlowStep is one step indicates an instance of a module in the blueprint,
// It includes the name of the module template (spec) and the parameters received by the component instance
// that is initiated by the orchestrator.
type FlowStep struct {

	// Name is the name of the instance of the module.
	// For example, if the application is named "notebook" and an implicitcopy
	// module is deemed necessary.  The FlowStep name would be notebook-implicitcopy.
	// +required
	Name string `json:"name"`

	// +required
	// Template is the name of the specification in the Blueprint describing how to instantiate
	// a component indicated by the module.  It is the name of a M4DModule CRD.
	// For example: implicit-copy-db2wh-to-s3-latest
	Template string `json:"template"`

	// Arguments are the input parameters for a specific instance of a module.
	// +optional
	Arguments ModuleArguments `json:"arguments,omitempty"`
}

// ComponentTemplate is a copy of a M4DModule Custom Resource.  It contains the information necessary
// to instantiate a component in a FlowStep, which provides the functionality described by the module.  There are 3 different module types.
type ComponentTemplate struct {

	// Name of the template
	// +required
	Name string `json:"name"`

	// Kind of k8s resource
	// +required
	Kind string `json:"kind"`

	// Chart contains the location of the helm chart with info detailing how to deploy
	// +required
	Chart ChartSpec `json:"chart"`
}

// DataFlow indicates the flow of the data between the components
// Currently we assume this is linear and thus use steps, but other more complex graphs could be defined
// as per how it is done in argo workflow
type DataFlow struct {

	// +required
	Name string `json:"name"`

	// +required
	// +kubebuilder::validation:MinItems:=1
	Steps []FlowStep `json:"steps"`
}

// BlueprintSpec defines the desired state of Blueprint, which is the runtime environment
// which provides the Data Scientist's application with secure and governed access to the data requested in the
// M4DApplication.
// The blueprint uses an "argo like" syntax which indicates the components and the flow of data between them as steps
// TODO: Add an indication of the communication relationships between the components
type BlueprintSpec struct {
	// +required
	Entrypoint string `json:"entrypoint"`

	// +required
	Flow DataFlow `json:"flow"`

	// +required
	// +kubebuilder::validation:MinItems:=1
	Templates []ComponentTemplate `json:"templates"`
}

// BlueprintStatus defines the observed state of Blueprint
// This includes readiness, error message, and indicators forthe Kubernetes
// resources owned by the Blueprint for cleanup and status monitoring
type BlueprintStatus struct {
	// ObservedState includes information to be reported back to the M4DApplication resource
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
	BlueprintNamespaceLabel = "app.m4d.ibm.com/blueprintNamespace"
	BlueprintNameLabel      = "app.m4d.ibm.com/blueprintName"
)
