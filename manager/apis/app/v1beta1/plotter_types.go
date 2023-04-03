// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

// AssetDetails is a list of assets used in the fybrikapplication. In addition to assets declared in
// fybrikapplication, AssetDetails list also contains assets that are allocated by the control-plane
// in order to serve fybrikapplication
type AssetDetails struct {
	// AdvertisedAssetID links this asset to asset from fybrikapplication and is used by user facing services
	// +optional
	AdvertisedAssetID string `json:"advertisedAssetId,omitempty"`

	// +required
	DataStore DataStore `json:"assetDetails"`
}

// StepArgument describes a step: it could be assetID
// or an endpoint of another step
type StepArgument struct {
	// AssetID identifies the source asset of this step
	// +optional
	AssetID string `json:"assetId,omitempty"`

	// API holds information for accessing a module instance
	// +optional
	API *datacatalog.ResourceDetails `json:"api,omitempty"`
}

// StepParameters holds the parameters to the module
// that is deployed in this step
type StepParameters struct {
	// +optional
	Arguments []*StepArgument `json:"args,omitempty"`

	// +optional
	API *datacatalog.ResourceDetails `json:"api,omitempty"`

	// Actions are the data transformations that the module supports
	// +optional
	Actions []taxonomy.Action `json:"action,omitempty"`
}

// DataFlowStep contains details on a single data flow step
type DataFlowStep struct {
	// Name of the step
	// +required
	Name string `json:"name"`

	// Name of the cluster this step is executed on
	// +required
	Cluster string `json:"cluster"`

	// Template is the name of the template to execute the step
	// The full details of the template can be extracted from Plotter.spec.templates
	// list field.
	// +required
	Template string `json:"template"`

	// Step parameters
	// TODO why not flatten the parameters into this data flow step
	// +optional
	Parameters *StepParameters `json:"parameters,omitempty"`
}

// SubFlowTrigger indicates the trigger for this subflow
// +kubebuilder:validation:Enum=workload;init;timer
type SubFlowTrigger string

// TODO: These will come from the taxonomy in the future.
const (
	// Init flow trigger
	InitTrigger SubFlowTrigger = "init"

	// Workload flow trigger
	WorkloadTrigger SubFlowTrigger = "workload"

	// Timer flow trigger
	TimerTrigger SubFlowTrigger = "timer"
)

// Subflows is a list of data flows which are originated from the same data asset
// but are triggered differently (e.g., one upon init
// trigger and one upon workload trigger)
type SubFlow struct {
	// Name of the SubFlow
	// +required
	Name string `json:"name"`

	// Type of the flow (e.g. read)
	// +required
	FlowType taxonomy.DataFlow `json:"flowType"`

	// Triggers
	// +required
	Triggers []SubFlowTrigger `json:"triggers"`

	// Steps defines a series of sequential/parallel data flow steps
	// The first dimension represents parallel data flows. The second sequential components
	// within the same parallel data flow.
	// +required
	Steps [][]DataFlowStep `json:"steps" protobuf:"bytes,11,opt,name=steps"`
}

// Flows is the list of data flows driven from fybrikapplication:
// Each element in the list holds the flow of the data requested in fybrikapplication.
type Flow struct {
	// Name of the flow
	// +required
	Name string `json:"name"`

	// Type of the flow (e.g. read)
	// +required
	FlowType taxonomy.DataFlow `json:"flowType"`

	// AssetID indicates the data set being used in this data flow
	// +required
	AssetID string `json:"assetId"`

	// +required
	SubFlows []SubFlow `json:"subFlows"`
}

// ModuleInfo is a copy of FybrikModule Custom Resource.  It contains information
// to instantiate resource of type FybrikModule.
type ModuleInfo struct {
	// Name of the module
	// +required
	Name string `json:"name"`

	// May be one of service, config or plugin
	// Service: Means that the control plane deploys the component that performs the capability
	// Config: Another pre-installed service performs the capability and the module deployed configures
	// it for the particular workload or dataset
	// Plugin: Indicates that this module performs a capability as part of another service or module rather than as a stand-alone module
	// +required
	Type string `json:"type"`

	// Chart contains the information needed to use helm to install the capability
	// +required
	Chart ChartSpec `json:"chart"`

	// Scope indicates at what level the capability is used: workload, asset, cluster
	// If not indicated it is assumed to be asset
	// +optional
	Scope CapabilityScope `json:"scope,omitempty"`

	// Module capability
	// +required
	Capability taxonomy.Capability `json:"capability"`
}

// Template contains basic information about the required modules to serve the fybrikapplication
// e.g., the module helm chart name.
type Template struct {
	// Name of the template
	// +required
	Name string `json:"name,omitempty"`

	// Modules is a list of dependent modules. e.g., if a plugin module is used
	// then the service module is used in should appear first in the modules list of the
	// same template. If the modules list contains more than one module, the first module in the list
	// is referred to as the "primary module" of which all the parameters to this template are
	// sent to.
	// +required
	Modules []ModuleInfo `json:"modules"`
}

// FlowStatus includes information to be reported back to the FybrikApplication resource
// It holds the status per data flow
type FlowStatus struct {
	// ObservedState includes information about the current flow
	// It includes readiness and error indications, as well as user instructions
	// +optional
	ObservedState ObservedState `json:"status,omitempty"`

	// +required
	SubFlows map[string]ObservedState `json:"subFlows"`
}

// PlotterSpec defines the desired state of Plotter, which is applied in a multi-clustered environment.
// Plotter declares what needs to be installed and where (as blueprints running on remote clusters)
// which provides the Data Scientist's application with secure and governed access to the data requested in the
// FybrikApplication.
type PlotterSpec struct {
	// Selector enables to connect the resource to the application
	// Application labels should match the labels in the selector.
	// For some flows the selector may not be used.
	// +optional
	Selector Selector `json:"appSelector,omitempty"`

	// Application context to be transferred to the modules
	AppInfo taxonomy.AppInfo `json:"appInfo,omitempty"`

	// Assets is a map holding information about the assets
	// The key is the assetID
	// +required
	Assets map[string]AssetDetails `json:"assets"`

	// +required
	Flows []Flow `json:"flows"`

	// ModulesNamespace is the namespace where modules should be allocated
	// +required
	ModulesNamespace string `json:"modulesNamespace"`

	// Templates is a map holding the templates used in this plotter steps
	// The key is the template name
	// +required
	Templates map[string]Template `json:"templates"`
}

// PlotterStatus defines the observed state of Plotter
// This includes readiness, error message, and indicators received from blueprint
// resources owned by the Plotter for cleanup and status monitoring
type PlotterStatus struct {
	// ObservedState includes information to be reported back to the FybrikApplication resource
	// It includes readiness and error indications, as well as user instructions
	// +optional
	ObservedState ObservedState `json:"observedState,omitempty"`

	// ObservedGeneration is taken from the Plotter metadata.  This is used to determine during reconcile
	// whether reconcile was called because the desired state changed, or whether status of the allocated blueprints should be checked.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Flows is a map containing the status for each flow
	// the key is the flow name
	// +optional
	Flows map[string]FlowStatus `json:"flows,omitempty"`

	// Assets is a map containing the status per asset.
	// The key of this map is assetId
	// +optional
	Assets map[string]ObservedState `json:"assets,omitempty"`

	// +optional
	Blueprints map[string]MetaBlueprint `json:"blueprints,omitempty"`

	// Conditions represent the possible error and failure conditions
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`

	// + optional
	ReadyTimestamp *metav1.Time `json:"readyTimestamp,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.observedState.ready`
// +kubebuilder:printcolumn:name="ReadySince",type=string,JSONPath=`.status.readyTimestamp`

// Plotter is the Schema for the plotters API
type Plotter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   PlotterSpec   `json:"spec"`
	Status PlotterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PlotterList contains a list of Plotter resources
type PlotterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Plotter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Plotter{}, &PlotterList{})
}
