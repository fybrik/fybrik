// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	api "fybrik.io/fybrik/manager/apis/app/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

// DefaultConfig implements EvaluatorInterface
// It provides a default configuration as an alternative to evaluation of the written rego policies
type DefaultConfig struct {
}

// NewDefaultConfig constructs a new DefaultConfig object
func NewDefaultConfig() *DefaultConfig {
	return &DefaultConfig{}
}

// Evaluate replaces hard-coded decisions in manager by default configuration
// The following logic is implemented:
/* 	Read capability is deployed in a read-type scenario.
    Read capability is deployed at the workload scope.
	Write capability is not yet supported and thus won’t be deployed.
	Transforms should always be performed close to the data.
	Copy is deployed if a user has requested it explicitly.
	Copy is deployed if there is no read module that supports the asset format.
	Copy is deployed in a read scenario if dataset resides in a different geography and governance actions are required.
*/
func (r *DefaultConfig) Evaluate(in *EvaluatorInput) (EvaluatorOutput, error) {
	decisions := map[api.CapabilityType]Decision{}
	// Read capability is deployed in a read-type scenario.
	deployRead := corev1.ConditionFalse
	if in.AssetRequirements.Usage[api.ReadFlow] {
		deployRead = corev1.ConditionTrue
	}
	decisions[api.Read] = Decision{Deploy: deployRead, Clusters: []string{in.Workload.Cluster.Name},
		Restrictions: map[string]string{"capabilities.scope": "workload"}}
	decisions[api.Write] = Decision{Deploy: corev1.ConditionFalse}

	copyDecision := DefaultDecision(in)
	if in.AssetRequirements.Usage[api.CopyFlow] {
		copyDecision.Deploy = corev1.ConditionTrue
	}

	clustersInRegion := []string{}
	for _, cluster := range in.Clusters {
		if cluster.Metadata.Region == in.AssetMetadata.Geography {
			clustersInRegion = append(clustersInRegion, cluster.Name)
		}
	}
	if deployRead == corev1.ConditionTrue && len(in.GovernanceActions) > 0 && in.Workload.Cluster.Metadata.Region != in.AssetMetadata.Geography {
		copyDecision.Deploy = corev1.ConditionTrue
		copyDecision.Clusters = clustersInRegion
	}

	transformDecision := DefaultDecision(in)
	transformDecision.Clusters = clustersInRegion

	decisions[api.Transform] = transformDecision
	decisions[api.Copy] = copyDecision

	return EvaluatorOutput{Valid: true, DatasetID: in.AssetRequirements.DatasetID, ConfigDecisions: decisions}, nil
}
