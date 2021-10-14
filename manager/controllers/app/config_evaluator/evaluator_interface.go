// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package config_evaluator

import corev1 "k8s.io/api/core/v1"

// ConfigEvaluatorInterface is an interface for config policies' evaluator
type ContextInConfigEvaluatorInterfaceterface interface {
	Evaluate(in *EvaluatorInput) (EvaluatorOutput, error)
}

// DefaultDecision creates a ConfigDecision object with some defaults e.g. any cluster is available, asset scope level, etc.
func DefaultDecision(in *EvaluatorInput) ConfigDecision {
	anyCluster := []string{in.Workload.Cluster.Name}
	for _, cluster := range in.Clusters {
		anyCluster = append(anyCluster, cluster.Name)
	}
	return ConfigDecision{Deploy: corev1.ConditionUnknown, Clusters: anyCluster,
		Restrictions: map[string]string{"capabilities.scope": "asset"}}
}
