// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import corev1 "k8s.io/api/core/v1"

// EvaluatorInterface is an interface for config policies' evaluator
type EvaluatorInterface interface {
	SetupWithInfrastructureManager(mgr *InfrastructureManager)
	Evaluate(in *EvaluatorInput) (EvaluatorOutput, error)
}

// DefaultDecision creates a Decision object with some defaults e.g. any cluster is available
func DefaultDecision(data *Infrastructure) Decision {
	anyCluster := []string{}
	for _, cluster := range data.Clusters {
		anyCluster = append(anyCluster, cluster.Name)
	}
	return Decision{
		Deploy: corev1.ConditionUnknown,
		DeploymentRestrictions: Restrictions{
			Clusters:           anyCluster,
			ModuleRestrictions: map[string]string{},
		},
	}
}
