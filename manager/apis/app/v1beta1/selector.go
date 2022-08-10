// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Selector is a label query over a set of resources in the specified cluster.
type Selector struct {
	// Cluster name
	// +optional
	ClusterName string `json:"clusterName"`

	// WorkloadSelector enables to connect the resource to the application
	// Application labels should match the labels in the selector.
	// +required
	WorkloadSelector metav1.LabelSelector `json:"workloadSelector"`
}
