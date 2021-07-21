// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StreamTransferSpec defines the desired state of StreamTransfer
type StreamTransferSpec struct {
	// Source data store for this batch job
	Source DataStore `json:"source"`

	// Destination data store for this batch job
	Destination DataStore `json:"destination"`

	// Transformations to be applied to the source data before writing to destination
	Transformation []Transformation `json:"transformation,omitempty"`

	// Interval in which the Micro batches of this stream should be triggered
	// The default is '5 seconds'.
	// +optional
	TriggerInterval string `json:"triggerInterval,omitempty"`

	// Image that should be used for the actual batch job. This is usually a datamover
	// image. This property will be defaulted by the webhook if not set.
	// +optional
	Image string `json:"image"`

	// Image pull policy that should be used for the actual job.
	// This property will be defaulted by the webhook if not set.
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`

	// Secret provider url that should be used for the actual job.
	// This property will be defaulted by the webhook if not set.
	// +optional
	SecretProviderURL string `json:"secretProviderURL,omitempty"`

	// Secret provider role that should be used for the actual job.
	// This property will be defaulted by the webhook if not set.
	// +optional
	SecretProviderRole string `json:"secretProviderRole,omitempty"`

	// If this batch job instance is run on a schedule the regular schedule can be suspended with this property.
	// This property will be defaulted by the webhook if not set.
	// +optional
	Suspend bool `json:"suspend,omitempty"`

	// If this batch job instance should have a finalizer or not.
	// This property will be defaulted by the webhook if not set.
	// +optional
	NoFinalizer bool `json:"noFinalizer,omitempty"`

	// Data flow type that specifies if this is a stream or a batch workflow
	// +optional
	DataFlowType DataFlowType `json:"flowType,omitempty"`

	// Data type of the data that is read from source (log data or change data)
	// +optional
	ReadDataType DataType `json:"readDataType,omitempty"`

	// Data type of how the data should be written to the target (log data or change data)
	// +optional
	WriteDataType DataType `json:"writeDataType,omitempty"`

	// Write operation that should be performed when writing (overwrite,append,update)
	// Caution: Some write operations are only available for batch and some only for stream.
	// +optional
	WriteOperation WriteOperation `json:"writeOperation,omitempty"`
}

// StreamTransferStatus defines the observed state of StreamTransfer
type StreamTransferStatus struct {
	// A pointer to the currently running job (or nil)
	// +optional
	Active *corev1.ObjectReference `json:"active,omitempty"`

	// +optional
	Status StreamStatus `json:"status,omitempty"`

	// +optional
	Error string `json:"error,omitempty"`
}

// +kubebuilder:validation:Enum=STARTING;RUNNING;STOPPED;FAILING
type StreamStatus string

// to be refined...
const (
	StreamStarting StreamStatus = "STARTING"
	StreamRunning  StreamStatus = "RUNNING"
	StreamStopped  StreamStatus = "STOPPED"
	StreamFailing  StreamStatus = "FAILING"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Source",type=string,JSONPath=`.spec.source.description`
// +kubebuilder:printcolumn:name="Destination",type=string,JSONPath=`.spec.destination.description`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:resource:scope=Namespaced

// StreamTransfer is the Schema for the streamtransfers API
type StreamTransfer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StreamTransferSpec   `json:"spec,omitempty"`
	Status StreamTransferStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StreamTransferList contains a list of StreamTransfer
type StreamTransferList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StreamTransfer `json:"items"`
}

const StreamtransferFinalizer = "streamtransfer.finalizers.ibm.com"
const StreamtransferBinary = "/stream"

func init() {
	SchemeBuilder.Register(&StreamTransfer{}, &StreamTransferList{})
}

// IsBeingDeleted returns true if a deletion timestamp is set
func (streamTransfer *StreamTransfer) IsBeingDeleted() bool {
	return !streamTransfer.ObjectMeta.DeletionTimestamp.IsZero()
}

func (streamTransfer *StreamTransfer) HasStarted() bool {
	return streamTransfer.Status.Active != nil
}

func (streamTransfer *StreamTransfer) HasFinalizer() bool {
	return controllerutil.ContainsFinalizer(streamTransfer, StreamtransferFinalizer)
}

func (streamTransfer *StreamTransfer) AddFinalizer() {
	controllerutil.AddFinalizer(streamTransfer, StreamtransferFinalizer)
}

func (streamTransfer *StreamTransfer) RemoveFinalizer() {
	controllerutil.RemoveFinalizer(streamTransfer, StreamtransferFinalizer)
}

func (streamTransfer *StreamTransfer) FinalizerPodName() string {
	return streamTransfer.Name + "-finalizer"
}

func (streamTransfer *StreamTransfer) FinalizerPodKey() client.ObjectKey {
	return client.ObjectKey{
		Namespace: streamTransfer.Namespace,
		Name:      streamTransfer.FinalizerPodName(),
	}
}

func (streamTransfer *StreamTransfer) ObjectKey() client.ObjectKey {
	return client.ObjectKey{
		Namespace: streamTransfer.Namespace,
		Name:      streamTransfer.Name,
	}
}

func (streamTransfer *StreamTransfer) GetImage() string {
	return streamTransfer.Spec.Image
}

func (streamTransfer *StreamTransfer) GetImagePullPolicy() corev1.PullPolicy {
	return streamTransfer.Spec.ImagePullPolicy
}
