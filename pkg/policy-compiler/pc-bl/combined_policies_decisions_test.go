// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package policycompilerbl

import (
	"testing"

	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	tu "github.com/ibm/the-mesh-for-data/pkg/policy-compiler/testutil"
)

var usedPolicy1 = &pb.Policy{Description: "policy 1 description"}
var usedPolicy2 = &pb.Policy{Description: "policy 2 description"}

var dataset1 = &pb.DatasetIdentifier{DatasetId: "{\"id\": \"mock-datasetID-1\"}"}
var dataset2 = &pb.DatasetIdentifier{DatasetId: "{\"id\": \"mock-datasetID-2\"}"}

var column1 = "mock-col-1"
var column2 = "mock-col-2"

func checkCombinedPolicies(t *testing.T, decisions1, decisions2 *pb.PoliciesDecisions) {
	combinedPolicies := GetCombinedPoliciesDecisions(decisions1, decisions2)

	for _, datasetDecisions := range decisions1.DatasetDecisions {
		tu.VerifyContainsDatasetDecision(t, combinedPolicies, datasetDecisions)
	}
	for _, datasetDecisions := range decisions2.DatasetDecisions {
		tu.VerifyContainsDatasetDecision(t, combinedPolicies, datasetDecisions)
	}
	// here could be verification on order if we would work with it
	// TODO: check duplication of same policy/action
}

// TestSingleColumnSingleOperationCombine checks if multiple enforcement actions combined
// together performed on a single column is returned correctly in accordance with what is expected

func TestSingleColumnSingleOperationCombine(t *testing.T) {
	operationDecision1 := &pb.OperationDecision{Operation: &pb.AccessOperation{Type: pb.AccessOperation_READ},
		EnforcementActions: []*pb.EnforcementAction{tu.ConstructRemoveColumn(column1), tu.ConstructRedactColumn(column1)},
		UsedPolicies:       []*pb.Policy{usedPolicy1}}

	datasetDecison1 := &pb.DatasetDecision{Dataset: dataset1, Decisions: []*pb.OperationDecision{operationDecision1}}

	operationDecision2 := &pb.OperationDecision{Operation: &pb.AccessOperation{Type: pb.AccessOperation_READ},
		EnforcementActions: []*pb.EnforcementAction{tu.ConstructEncryptColumn(column1)},
		UsedPolicies:       []*pb.Policy{usedPolicy2}}

	datasetDecison2 := &pb.DatasetDecision{Dataset: dataset1, Decisions: []*pb.OperationDecision{operationDecision2}}

	checkCombinedPolicies(t, &pb.PoliciesDecisions{DatasetDecisions: []*pb.DatasetDecision{datasetDecison1}},
		&pb.PoliciesDecisions{DatasetDecisions: []*pb.DatasetDecision{datasetDecison2}})
}

// TestSingleDatasetSingleOperationCombine checks if multiple enforcement actions combined
// together and performed on multiple columns pertaining to one single dataset
// then if the policies are returned correctly in accordance with what is expected

func TestSingleDatasetSingleOperationCombine(t *testing.T) {
	operationDecision1 := &pb.OperationDecision{Operation: &pb.AccessOperation{Type: pb.AccessOperation_READ},
		EnforcementActions: []*pb.EnforcementAction{tu.ConstructRemoveColumn(column1)},
		UsedPolicies:       []*pb.Policy{usedPolicy1}}

	datasetDecison1 := &pb.DatasetDecision{Dataset: dataset1, Decisions: []*pb.OperationDecision{operationDecision1}}

	operationDecision2 := &pb.OperationDecision{Operation: &pb.AccessOperation{Type: pb.AccessOperation_READ},
		EnforcementActions: []*pb.EnforcementAction{tu.ConstructRemoveColumn(column2)},
		UsedPolicies:       []*pb.Policy{usedPolicy2}}

	datasetDecison2 := &pb.DatasetDecision{Dataset: dataset1, Decisions: []*pb.OperationDecision{operationDecision2}}

	checkCombinedPolicies(t, &pb.PoliciesDecisions{DatasetDecisions: []*pb.DatasetDecision{datasetDecison1}},
		&pb.PoliciesDecisions{DatasetDecisions: []*pb.DatasetDecision{datasetDecison2}})
}

// TestSingleColumnTwoOperationsCombine checks if two operations are combined
// together and performed on single columns  then if the copmbined  policies are returned
// correctly in accordance with what is expected

func TestSingleColumnTwoOperationsCombine(t *testing.T) {
	operationDecision1 := &pb.OperationDecision{Operation: &pb.AccessOperation{Type: pb.AccessOperation_COPY},
		EnforcementActions: []*pb.EnforcementAction{tu.ConstructRemoveColumn(column1)},
		UsedPolicies:       []*pb.Policy{usedPolicy1}}

	datasetDecison1 := &pb.DatasetDecision{Dataset: dataset1, Decisions: []*pb.OperationDecision{operationDecision1}}

	operationDecision2 := &pb.OperationDecision{Operation: &pb.AccessOperation{Type: pb.AccessOperation_READ},
		EnforcementActions: []*pb.EnforcementAction{tu.ConstructEncryptColumn(column1), tu.ConstructRedactColumn(column1)},
		UsedPolicies:       []*pb.Policy{usedPolicy2}}

	datasetDecison2 := &pb.DatasetDecision{Dataset: dataset1, Decisions: []*pb.OperationDecision{operationDecision2}}

	checkCombinedPolicies(t, &pb.PoliciesDecisions{DatasetDecisions: []*pb.DatasetDecision{datasetDecison1}},
		&pb.PoliciesDecisions{DatasetDecisions: []*pb.DatasetDecision{datasetDecison2}})
}

// TestTwoDatasetsSingleOperationCombine checks if single operations is performed on multiple
// datasets and then the policies are combined together then if the combined policies are
// returned correctly in accordance with what is expected

func TestTwoDatasetsSingleOperationCombine(t *testing.T) {
	operationDecision1 := &pb.OperationDecision{Operation: &pb.AccessOperation{Type: pb.AccessOperation_COPY},
		EnforcementActions: []*pb.EnforcementAction{tu.ConstructRemoveColumn(column1)},
		UsedPolicies:       []*pb.Policy{usedPolicy1}}

	datasetDecison1 := &pb.DatasetDecision{Dataset: dataset1, Decisions: []*pb.OperationDecision{operationDecision1}}

	datasetDecison2 := &pb.DatasetDecision{Dataset: dataset2, Decisions: []*pb.OperationDecision{operationDecision1}}

	checkCombinedPolicies(t, &pb.PoliciesDecisions{DatasetDecisions: []*pb.DatasetDecision{datasetDecison1}},
		&pb.PoliciesDecisions{DatasetDecisions: []*pb.DatasetDecision{datasetDecison2}})
}
