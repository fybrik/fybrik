// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
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
	Transformations []pb.EnforcementAction `json:"transformations,omitempty"`
}

// ReadModuleArgs define the input parameters for modules that read data from location A
type ReadModuleArgs struct {
	// Source of the read path module
	// +required
	Source DataStore `json:"source"`

	// AssetName represents the asset name to be used for accessing the data when it is ready
	// It is copied from the m4dapplication resource
	// +required
	AssetName string `json:"assetName"`

	// Transformations are different types of processing that may be done to the data
	// +optional
	Transformations []pb.EnforcementAction `json:"transformations,omitempty"`
}

// WriteModuleArgs define the input parameters for modules that write data to location B
type WriteModuleArgs struct {
	// Destination is the data store to which the data will be written
	// +required
	Destination DataStore `json:"destination"`

	// Transformations are different types of processing that may be done to the data as it is written.
	// +optional
	Transformations []pb.EnforcementAction `json:"transformations,omitempty"`
}

// ModuleArguments are the parameters passed to a component that runs in the data path
// In the future might support output args as well
// The arguments passed depend on the type of module
type ModuleArguments struct {

	// ONE AND ONLY ONE OF THE FOLLOWING FIELDS SHOULD BE POPULATED
	// Flow is the selector for this union
	// +required
	Flow ModuleFlow `json:"flow"`

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

	// Resources contains the location of the helm chart with info detailing how to deploy
	// +required
	Resources []string `json:"resources"`
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

	// Selector enables to connect the resource to the application
	// Should match the selector of the owner - M4DApplication CRD.
	Selector metav1.LabelSelector `json:"selector"`

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
	// Ready represents that the modules have been orchestrated successfully and the data is ready for usage
	// +optional
	Ready bool `json:"ready,omitempty"`
	// Error indicates that there has been an error to orchestrate the modules and provides the error message
	// +optional
	Error string `json:"error,omitempty"`
	// DataAccessInstructions indicate how the data user or his application may access the data.
	// Instructions are available upon successful orchestration.
	// +optional
	DataAccessInstructions string `json:"dataAccessInstructions,omitempty"`
	// ObservedGeneration is taken from the Blueprint metadata.  This is used to determine during reconcile
	// whether reconcile was called because the desired state changed, or whether status of the allocated resources should be checked.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced

// Blueprint is the Schema for the blueprints API
type Blueprint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BlueprintSpec   `json:"spec,omitempty"`
	Status BlueprintStatus `json:"status,omitempty"`
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
