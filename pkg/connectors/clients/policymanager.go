// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"io"

	pb "fybrik.io/fybrik/pkg/connectors/protobuf"
	openapiclientmodels "fybrik.io/fybrik/pkg/taxonomy/model/base"
)

// PolicyManager is an interface of a facade to connect to a policy manager.
type PolicyManager interface {
	//pb.PolicyManagerServiceServer
	GetPoliciesDecisions(in *openapiclientmodels.PolicyManagerRequest, creds string) (*openapiclientmodels.PolicyManagerResponse, error)
	io.Closer
}

func MergePoliciesDecisions(in ...*pb.PoliciesDecisions) *pb.PoliciesDecisions {
	result := &pb.PoliciesDecisions{}

	for _, decisions := range in {
		result.ComponentVersions = append(result.ComponentVersions, decisions.ComponentVersions...)
		result.GeneralDecisions = append(result.GeneralDecisions, decisions.GeneralDecisions...)
		result.DatasetDecisions = append(result.DatasetDecisions, decisions.DatasetDecisions...)
	}

	result = compactPolicyDecisions(result)
	return result
}

// compactPolicyDecisions compacts policy decisions by merging decisions of same dataset identifier and same operation.
func compactPolicyDecisions(in *pb.PoliciesDecisions) *pb.PoliciesDecisions {
	if in == nil {
		return nil
	}

	result := &pb.PoliciesDecisions{
		ComponentVersions: in.ComponentVersions,
		DatasetDecisions:  []*pb.DatasetDecision{},
		GeneralDecisions:  compactOperationDecisions(in.GeneralDecisions),
	}

	// Group and flatten decisions by dataset id
	decisionsByIDKeys := []string{} // for determitistric results
	decisionsByID := map[string]*pb.DatasetDecision{}
	for _, datasetDecision := range in.DatasetDecisions {
		datasetID := datasetDecision.Dataset.DatasetId
		if _, exists := decisionsByID[datasetID]; !exists {
			decisionsByIDKeys = append(decisionsByIDKeys, datasetID)
			decisionsByID[datasetID] = &pb.DatasetDecision{
				Dataset: datasetDecision.Dataset,
			}
		}
		decisionsByID[datasetID].Decisions = append(decisionsByID[datasetID].Decisions, datasetDecision.Decisions...)
	}

	// Compact DatasetDecisions
	for _, key := range decisionsByIDKeys {
		datasetDecision := decisionsByID[key]
		result.DatasetDecisions = append(result.DatasetDecisions, &pb.DatasetDecision{
			Dataset:   datasetDecision.Dataset,
			Decisions: compactOperationDecisions(datasetDecision.Decisions),
		})
	}

	return result
}

func compactOperationDecisions(in []*pb.OperationDecision) []*pb.OperationDecision {
	if len(in) == 0 {
		return nil
	}

	type operationKeyType [2]interface{}

	// Group and flatten decisions for a specific dataset id by operation
	decisionsByOperationKeys := []operationKeyType{} // for determitistric results
	decisionsByOperation := map[operationKeyType]*pb.OperationDecision{}
	for _, operationDecision := range in {
		key := operationKeyType{operationDecision.Operation.Type, operationDecision.Operation.Destination}
		if _, exists := decisionsByOperation[key]; !exists {
			decisionsByOperationKeys = append(decisionsByOperationKeys, key)
			decisionsByOperation[key] = &pb.OperationDecision{
				Operation: operationDecision.Operation,
			}
		}
		decisionsByOperation[key].EnforcementActions = append(decisionsByOperation[key].EnforcementActions, operationDecision.EnforcementActions...)
		decisionsByOperation[key].UsedPolicies = append(decisionsByOperation[key].UsedPolicies, operationDecision.UsedPolicies...)
	}

	decisions := make([]*pb.OperationDecision, 0, len(decisionsByOperation))
	for _, key := range decisionsByOperationKeys {
		decision := decisionsByOperation[key]
		decisions = append(decisions, decision)
	}

	return decisions
}
