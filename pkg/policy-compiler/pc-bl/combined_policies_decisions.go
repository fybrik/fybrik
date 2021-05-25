// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package policycompilerbl

import (
	"strconv"

	pb "github.com/mesh-for-data/mesh-for-data/pkg/connectors/protobuf"
)

func GetCombinedPoliciesDecisions(firstDecisions *pb.PoliciesDecisions, secondDecisions *pb.PoliciesDecisions) *pb.PoliciesDecisions {
	// Create a map of maps
	combinedDecisions := make(map[string]map[[2]string]*pb.OperationDecision)

	// Populate using firstDecisions
	PopulateMapWithDecisions(firstDecisions, combinedDecisions)

	// Now populate with secondDecisions
	PopulateMapWithDecisions(secondDecisions, combinedDecisions)

	// Now create the object to return
	toReturn := &pb.PoliciesDecisions{DatasetDecisions: []*pb.DatasetDecision{}, ComponentVersions: append(firstDecisions.GetComponentVersions(), secondDecisions.GetComponentVersions()...), GeneralDecisions: append(firstDecisions.GetGeneralDecisions(), secondDecisions.GetGeneralDecisions()...)}

	for datasetIdentifier := range combinedDecisions {
		newDatasetIdentifier := pb.DatasetIdentifier{DatasetId: datasetIdentifier}

		newDatasetDecision := pb.DatasetDecision{Dataset: &newDatasetIdentifier, Decisions: []*pb.OperationDecision{}}

		for _, operationDecision := range combinedDecisions[datasetIdentifier] {
			newDatasetDecision.Decisions = append(newDatasetDecision.GetDecisions(), operationDecision)
		}
		toReturn.DatasetDecisions = append(toReturn.GetDatasetDecisions(), &newDatasetDecision)
	}
	return toReturn
}

func PopulateMapWithDecisions(decisions *pb.PoliciesDecisions, combinedDecisions map[string]map[[2]string]*pb.OperationDecision) {
	for _, datasetDecision := range decisions.GetDatasetDecisions() {
		currDatasetIdentifier := datasetDecision.GetDataset().GetDatasetId()

		if _, ok := combinedDecisions[currDatasetIdentifier]; !ok {
			// Add the key
			combinedDecisions[currDatasetIdentifier] = make(map[[2]string]*pb.OperationDecision)
		}

		// Iterate over operationDecisions
		for _, operationDecision := range datasetDecision.GetDecisions() {
			currAccessOperation := [2]string{strconv.Itoa(int(operationDecision.GetOperation().GetType())), operationDecision.GetOperation().GetDestination()}

			if _, ok := combinedDecisions[currDatasetIdentifier][currAccessOperation]; !ok {
				// Add the key
				combinedDecisions[currDatasetIdentifier][currAccessOperation] = &pb.OperationDecision{Operation: operationDecision.GetOperation(), EnforcementActions: operationDecision.GetEnforcementActions(), UsedPolicies: operationDecision.GetUsedPolicies()}
			} else {
				tempOperationDecision := combinedDecisions[currDatasetIdentifier][currAccessOperation]
				tempOperationDecision.EnforcementActions = append(tempOperationDecision.GetEnforcementActions(), operationDecision.GetEnforcementActions()...)

				tempOperationDecision.UsedPolicies = append(tempOperationDecision.GetUsedPolicies(), operationDecision.GetUsedPolicies()...)
			}
		}
	}
}
