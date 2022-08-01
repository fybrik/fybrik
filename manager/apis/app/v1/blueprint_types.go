// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"fybrik.io/fybrik/pkg/model/taxonomy"
)

// AssetContext defines the input parameters for modules that access an asset
type AssetContext struct {
	// List of datastores associated with the asset
	// +optional
	Arguments []*DataStore `json:"args,omitempty"`

	// AssetID identifies the asset to be used for accessing the data when it is ready
	// It is copied from the FybrikApplication resource
	// +required
	AssetID string `json:"assetID"`

	// Transformations are different types of processing that may be done to the data as it is copied.
	// +optional
	Transformations []taxonomy.Action `json:"transformations,omitempty"`

	// Capability of the module
	// +required
	Capability taxonomy.Capability `json:"capability"`
}

type ApplicationDetails struct {
	// Application selector is used to identify the user workload.
	// It is obtained from FybrikApplication spec.
	// +optional
	WorkloadSelector metav1.LabelSelector `json:"selector,omitempty"`

	// Application context such as intent, role, etc.
	// +optional
	Context taxonomy.AppInfo `json:"context,omitempty"`
}

// ModuleArguments are the parameters passed to a component that runs in the data path
type ModuleArguments struct {
	// Assets define asset related arguments, such as data source, transformations, etc.
	// +optional
	Assets []AssetContext `json:"assets,omitempty"`
}

// BlueprintModule is a copy of a FybrikModule Custom Resource.  It contains the information necessary
// to instantiate a datapath component, including the parameters relevant for the particular workload.
type BlueprintModule struct {
	// Name of the FybrikModule on which this is based
	// +required
	Name string `json:"name"`

	// Chart contains the location of the helm chart with info detailing how to deploy
	// +required
	Chart ChartSpec `json:"chart"`

	// Arguments are the input parameters for a specific instance of a module.
	// +optional
	Arguments ModuleArguments `json:"arguments,omitempty"`

	// assetIDs indicate the assets processed by this module.  Included so we can track asset status
	// as well as module status in the future.
	// +optional
	AssetIDs []string `json:"assetIds,omitempty"`
}

// BlueprintSpec defines the desired state of Blueprint, which defines the components of the workload's data path
// that run in a particular cluster.
// In a single cluster environment there is one blueprint per workload (FybrikApplication).
// In a multi-cluster environment there is one Blueprint per cluster per workload (FybrikApplication).
type BlueprintSpec struct {
	// Cluster indicates the cluster on which the Blueprint runs
	// +required
	Cluster string `json:"cluster"`

	// ModulesNamespace is the namespace where modules should be allocated
	// +required
	ModulesNamespace string `json:"modulesNamespace"`

	// Modules is a map which contains modules that indicate the data path components that run in this cluster
	// The map key is moduleInstanceName which is the unique name for the deployed instance related to this workload
	// +required
	Modules map[string]BlueprintModule `json:"modules"`

	// ApplicationContext is a context of the origin FybrikApplication (labels, properties, etc.)
	// +optional
	Application *ApplicationDetails `json:"application,omitempty"`
}

// BlueprintStatus defines the observed state of Blueprint
// This includes readiness, error message, and indicators for the Kubernetes
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

	// ModulesState is a map which holds the status of each module
	// its key is the moduleInstanceName which is the unique name for the deployed instance related to this workload
	// +optional
	ModulesState map[string]ObservedState `json:"modules,omitempty"`

	// Releases map each release to the observed generation of the blueprint containing this release.
	// At the end of reconcile, each release should be mapped to the latest blueprint version or be uninstalled.
	// +optional
	Releases map[string]int64 `json:"releases,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.observedState.ready`

// Blueprint is the Schema for the blueprints API
type Blueprint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   BlueprintSpec   `json:"spec"`
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
		Status: BlueprintStatus{
			ModulesState: map[string]ObservedState{},
		},
	}
	return metaBlueprint
}
