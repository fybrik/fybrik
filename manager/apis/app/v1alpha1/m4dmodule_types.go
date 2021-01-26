// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	connectors "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DependencyType indicates what type of pre-requisit is required
// +kubebuilder:validation:Enum=module;connector;feature
type DependencyType string

const (
	// Module indicates a reliance on another module
	Module DependencyType = "module"

	// Connector - example for connecting to data catalog, policy compiler, external credential manager
	Connector DependencyType = "connector"

	// Feature indicates a dependency on an optional control plane capability
	Feature DependencyType = "feature"
)

// ModuleFlow indicates what data flow is performed by the module
// +kubebuilder:validation:Enum=copy;read;write
type ModuleFlow string

const (
	// Copy moves data from one location to another - i.e implicit copy
	Copy ModuleFlow = "copy"

	// Write is accessed from within an application, typically through an SDK
	Write ModuleFlow = "write"

	// Read is accessed from within an application, typically through an SDK
	Read ModuleFlow = "read"
)

// CredentialManagementType indicates whether this module queries the SecretProvider by itself to get
// credentials, or whether it assumes that the data-mesh will inject them.
// +kubebuilder:validation:Enum=secret-provider;automatic
type CredentialManagementType string

const (
	// SecretProvider is set when the module uses the Secret Provider
	SecretProvider CredentialManagementType = "secret-provider"

	// Automatic is set when credential management is handled elsewhere
	Automatic CredentialManagementType = "automatic"
)

// ModuleInOut specifies the protocol and format of the data input and output by the module - if any
type ModuleInOut struct {

	// Flow for which this interface is supported
	// +required
	Flow ModuleFlow `json:"flow"`

	// Source specifies the input data protocol and format
	// +optional
	Source *InterfaceDetails `json:"source,omitempty"`

	// Sink specifies the output data protocol and format
	// +optional
	Sink *InterfaceDetails `json:"sink,omitempty"`
}

// Dependency details another component on which this module relies - i.e. a pre-requisit
type Dependency struct {

	// Type provides information used in determining how to instantiate the component
	// +required
	Type DependencyType `json:"type"`

	// Name is the name of the dependent component
	// +required
	Name string `json:"name"`
}

type SupportedAction struct {
	// +required
	ID string `json:"id,omitempty"`
	// +optional
	Level connectors.EnforcementAction_EnforcementActionLevel `json:"level,omitempty"`
}

// Capability declares what this module knows how to do and the types of data it knows how to handle
type Capability struct {

	// Type provides information used in determining how to instantiate the component
	// +required
	CredentialsManagedBy CredentialManagementType `json:"credentials-managed-by"`

	// Copy should have one or more instances in the list, and its content should have source and sink
	// Read should have one or more instances in the list, each with source populated
	// Write should have one or more instances in the list, each with sink populated
	// TODO - In the future if we have a module type that doesn't interface directly with data then this list could be empty
	// +required
	// +kubebuilder:validation:Min=1
	SupportedInterfaces []ModuleInOut `json:"supportedInterfaces"`

	// API indicates to the application how to access/write the data
	// +optional
	API *InterfaceDetails `json:"api,omitempty"`

	// Actions are the data transformations that the module supports
	// +optional
	Actions []SupportedAction `json:"actions,omitempty"`
}

// ResourceStatusIndicator is used to determine the status of an orchestrated resource
type ResourceStatusIndicator struct {
	// Kind provides information about the resource kind
	// +required
	Kind string `json:"kind"`

	// SuccessCondition specifies a condition that indicates that the resource is ready
	// It uses kubernetes label selection syntax (https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
	// +required
	SuccessCondition string `json:"successCondition"`

	// FailureCondition specifies a condition that indicates the resource failure
	// It uses kubernetes label selection syntax (https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
	// +optional
	FailureCondition string `json:"failureCondition"`

	// ErrorMessage specifies the resource field to check for an error, e.g. status.errorMsg
	// +optional
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// M4DModuleSpec contains the info common to all modules,
// which are one of the components that process, load, write, audit, monitor the data used by
// the data scientist's application.
type M4DModuleSpec struct {
	// Flows is a list of the types of capabilities supported by the module - copy, read, write
	// +required
	Flows []ModuleFlow `json:"flows"`

	// Other components that must be installed in order for this module to work
	// +optional
	Dependencies []Dependency `json:"dependencies,omitempty"`

	// Capabilities declares what this module knows how to do and the types of data it knows how to handle
	// +required
	Capabilities Capability `json:"capabilities"`

	// Reference to a Helm chart that allows deployment of the resources required for this module
	// +required
	Chart ChartSpec `json:"chart"`

	// StatusIndicators allow to check status of a non-standard resource that can not be computed by helm/kstatus
	// +optional
	StatusIndicators []ResourceStatusIndicator `json:"statusIndicators,omitempty"`
}

// ChartSpec specifies chart name and values
type ChartSpec struct {
	// Name of helm chart
	// +required
	Name string `json:"name"`

	// Values to pass to helm chart insallation
	// +optional
	Values map[string]string `json:"values,omitempty"`
}

// +kubebuilder:object:root=true

// M4DModule is a description of an injectable component.
// the parameters it requires, as well as the specification of how to instantiate such a component.
// It is used as metadata only.  There is no status nor reconciliation.
type M4DModule struct {

	// Metadata should include name, namespace, label, annotations.
	// annotations should include author, summary, description
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec M4DModuleSpec `json:"spec"`
}

// +kubebuilder:object:root=true

// M4DModuleList contains a list of M4DModule
type M4DModuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []M4DModule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&M4DModule{}, &M4DModuleList{})
}
