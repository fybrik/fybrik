// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PlotterSpec defines the desired state of Plotter, which is applied in a multi-clustered environment. Plotter installs the runtime environment
// (as blueprints running on remote clusters) which provides the Data Scientist's application with secure and governed access to the data requested in the
// M4DApplication.
type PlotterSpec struct {

	// Selector enables to connect the resource to the application
	// Should match the selector of the owner - M4DApplication CRD.
	Selector metav1.LabelSelector `json:"selector"`

	// +required
	// +kubebuilder::validation:MinItems:=1
	// Blueprints structure represents remote blueprints mapped by the identifier of a cluster in which they will be running
	Blueprints map[string]BlueprintSpec `json:"blueprints"`
}

// PlotterStatus defines the observed state of Plotter
// This includes readiness, error message, and indicators received from blueprint
// resources owned by the Plotter for cleanup and status monitoring
type PlotterStatus struct {
	// Ready represents that the installed blueprints have been orchestrated successfully and the data is ready for usage
	// +optional
	Ready bool `json:"ready,omitempty"`
	// Error indicates that there has been an error to orchestrate the modules of some blueprint and provides the error message
	// +optional
	Error string `json:"error,omitempty"`
	// DataAccessInstructions indicate how the data user or his application may access the data.
	// Instructions are available upon successful orchestration.
	// +optional
	DataAccessInstructions string `json:"dataAccessInstructions,omitempty"`
	// ObservedGeneration is taken from the Plotter metadata.  This is used to determine during reconcile
	// whether reconcile was called because the desired state changed, or whether status of the allocated blueprints should be checked.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced

// Plotter is the Schema for the plotters API
type Plotter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PlotterSpec   `json:"spec,omitempty"`
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
