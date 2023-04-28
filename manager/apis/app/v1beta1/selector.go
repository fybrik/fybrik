// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1beta1

import (
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Selector is a label query over a set of resources in the specified cluster.
type Selector struct {
	// Cluster name
	// +optional
	ClusterName string `json:"clusterName"`

	// WorkloadSelector enables to connect the resource to a user application.
	// Application labels should match the labels in the selector.
	// +required
	WorkloadSelector metav1.LabelSelector `json:"workloadSelector"`

	// Namespaces where user application might run
	// +optional
	Namespaces []string `json:"namespaces"`

	// IPBlocks define policy on particular IPBlocks.
	// the structure of the IPBlock is defined at
	// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#ipblock-v1-networking-k8s-io
	// +optional
	IPBlocks []*netv1.IPBlock `json:"ipBlocks,omitempty"`
}
