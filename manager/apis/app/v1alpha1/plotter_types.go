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
	// Blueprints structure represents remote blueprints mapped by the identifier of a cluster in which they will be running
	Blueprints map[string]BlueprintSpec `json:"blueprints"`
}

// PlotterStatus defines the observed state of Plotter
// This includes readiness, error message, and indicators received from blueprint
// resources owned by the Plotter for cleanup and status monitoring
type PlotterStatus struct {
	// ObservedState includes information to be reported back to the M4DApplication resource
	// It includes readiness and error indications, as well as user instructions
	// +optional
	ObservedState ObservedState `json:"observedState,omitempty"`

	// ObservedGeneration is taken from the Plotter metadata.  This is used to determine during reconcile
	// whether reconcile was called because the desired state changed, or whether status of the allocated blueprints should be checked.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// + optional
	Blueprints map[string]MetaBlueprint `json:"blueprints,omitempty"`

	// + optional
	ReadyTimestamp *metav1.Time `json:"readyTimestamp,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.observedState.ready`
// +kubebuilder:printcolumn:name="ReadySince",type=string,JSONPath=`.status.readyTimestamp`

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
