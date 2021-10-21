// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import (
	api "fybrik.io/fybrik/manager/apis/app/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

// DefaultConfig implements EvaluatorInterface
// It provides a default configuration as an alternative to evaluation of the written rego policies
type DefaultConfig struct {
	InfrastructureManager
	Data *Infrastructure
}

// NewDefaultConfig constructs a new DefaultConfig object
func NewDefaultConfig(manager InfrastructureManager) *DefaultConfig {
	return &DefaultConfig{InfrastructureManager: manager, Data: nil}
}

func (r *DefaultConfig) SetInfrastructureDetails() error {
	var err error
	r.Data, err = r.SetInfrastructure()
	return err
}

// DefaultDecision creates a Decision object with some defaults e.g. any cluster is available
func (r *DefaultConfig) DefaultDecision(in *EvaluatorInput) Decision {
	anyCluster := []string{in.Workload.Cluster.Name}
	for _, cluster := range r.Data.Clusters {
		anyCluster = append(anyCluster, cluster.Name)
	}
	return Decision{Deploy: corev1.ConditionUnknown, Clusters: anyCluster}
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
	if r.Data == nil {
		if err := r.SetInfrastructureDetails(); err != nil {
			return EvaluatorOutput{Valid: false}, err
		}
	}
	decisions := map[api.CapabilityType]Decision{}
	// Read capability is deployed in a read-type scenario.
	deployRead := corev1.ConditionFalse
	if in.AssetRequirements.Usage[api.ReadFlow] {
		deployRead = corev1.ConditionTrue
	}
	decisions[api.Read] = Decision{Deploy: deployRead, Clusters: []string{in.Workload.Cluster.Name},
		Restrictions: map[string]string{"capabilities.scope": "workload"}}
	decisions[api.Write] = Decision{Deploy: corev1.ConditionFalse}

	copyDecision := r.DefaultDecision(in)
	if in.AssetRequirements.Usage[api.CopyFlow] {
		copyDecision.Deploy = corev1.ConditionTrue
	}

	clustersInRegion := []string{}
	for _, cluster := range r.Data.Clusters {
		if cluster.Metadata.Region == in.AssetMetadata.Geography {
			clustersInRegion = append(clustersInRegion, cluster.Name)
		}
	}
	if deployRead == corev1.ConditionTrue && len(in.GovernanceActions) > 0 && in.Workload.Cluster.Metadata.Region != in.AssetMetadata.Geography {
		copyDecision.Deploy = corev1.ConditionTrue
		copyDecision.Clusters = clustersInRegion
	}

	transformDecision := r.DefaultDecision(in)
	transformDecision.Clusters = clustersInRegion

	decisions[api.Transform] = transformDecision
	decisions[api.Copy] = copyDecision

	return EvaluatorOutput{Valid: true, DatasetID: in.AssetRequirements.DatasetID, ConfigDecisions: decisions}, nil
}
