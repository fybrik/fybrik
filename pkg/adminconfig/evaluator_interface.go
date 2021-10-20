// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import corev1 "k8s.io/api/core/v1"

// EvaluatorInterface is an interface for config policies' evaluator
type EvaluatorInterface interface {
	Evaluate(in *EvaluatorInput) (EvaluatorOutput, error)
}

// DefaultDecision creates a Decision object with some defaults e.g. any cluster is available
func DefaultDecision(in *EvaluatorInput) Decision {
	anyCluster := []string{in.Workload.Cluster.Name}
	for _, cluster := range in.Clusters {
		anyCluster = append(anyCluster, cluster.Name)
	}
	return Decision{Deploy: corev1.ConditionUnknown, Clusters: anyCluster}
}
