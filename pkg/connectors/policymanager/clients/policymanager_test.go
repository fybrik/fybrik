// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"fybrik.io/fybrik/pkg/connectors/policymanager/clients"
	pb "fybrik.io/fybrik/pkg/connectors/protobuf"
)

func getTemplate(datasetID string, operation *pb.AccessOperation, actions ...*pb.EnforcementAction) *pb.PoliciesDecisions {
	return &pb.PoliciesDecisions{DatasetDecisions: []*pb.DatasetDecision{{
		Dataset: &pb.DatasetIdentifier{DatasetId: datasetID},
		Decisions: []*pb.OperationDecision{
			{
				Operation:          operation,
				EnforcementActions: actions,
			},
		},
	}}}
}

var _ = Describe("PolicyManager", func() {
	Describe("merge policy decisions", func() {

		removeColumn1 := &pb.EnforcementAction{Name: "remove column", Id: "remove-ID", Level: pb.EnforcementAction_COLUMN,
			Args: map[string]string{"column_name": "col1"}}
		removeColumn2 := &pb.EnforcementAction{Name: "remove column", Id: "remove-ID", Level: pb.EnforcementAction_COLUMN,
			Args: map[string]string{"column_name": "col2"}}
		redactColumn1 := &pb.EnforcementAction{Name: "redact column", Id: "redact-ID", Level: pb.EnforcementAction_COLUMN,
			Args: map[string]string{"column_name": "col1"}}

		Context("on same dataset", func() {

			Context("with same operation (read)", func() {

				Context("has multiple actions on the same column", func() {

					It("should put the actions in the same EnforcementActions slice", func() {
						left := getTemplate("1", &pb.AccessOperation{Type: pb.AccessOperation_READ}, removeColumn1)
						right := getTemplate("1", &pb.AccessOperation{Type: pb.AccessOperation_READ}, redactColumn1)
						expected := getTemplate("1", &pb.AccessOperation{Type: pb.AccessOperation_READ}, removeColumn1, redactColumn1)
						Expect(clients.MergePoliciesDecisions(left, right)).To(Equal(expected))
					})
				})

				Context("has same action type but on different columns", func() {
					It("should put the actions in the same EnforcementActions slice", func() {
						left := getTemplate("1", &pb.AccessOperation{Type: pb.AccessOperation_READ}, removeColumn1)
						right := getTemplate("1", &pb.AccessOperation{Type: pb.AccessOperation_READ}, removeColumn2)
						expected := getTemplate("1", &pb.AccessOperation{Type: pb.AccessOperation_READ}, removeColumn1, removeColumn2)
						Expect(clients.MergePoliciesDecisions(left, right)).To(Equal(expected))
					})
				})
			})

			Context("with multiple operations (read, write)", func() {

				Context("has same action on the same column", func() {
					It("should result in two decisions for the dataset", func() {
						left := getTemplate("1", &pb.AccessOperation{Type: pb.AccessOperation_READ}, removeColumn1)
						right := getTemplate("1", &pb.AccessOperation{Type: pb.AccessOperation_WRITE}, removeColumn1)
						expected := &pb.PoliciesDecisions{DatasetDecisions: []*pb.DatasetDecision{{
							Dataset: &pb.DatasetIdentifier{DatasetId: "1"},
							Decisions: []*pb.OperationDecision{
								{
									Operation:          &pb.AccessOperation{Type: pb.AccessOperation_READ},
									EnforcementActions: []*pb.EnforcementAction{removeColumn1},
								},
								{
									Operation:          &pb.AccessOperation{Type: pb.AccessOperation_WRITE},
									EnforcementActions: []*pb.EnforcementAction{removeColumn1},
								},
							},
						}}}
						Expect(clients.MergePoliciesDecisions(left, right)).To(Equal(expected))
					})
				})

			})
		})

		Context("on two datasets", func() {
			It("should keep as separate dataset decisions", func() {
				left := getTemplate("1", &pb.AccessOperation{Type: pb.AccessOperation_READ}, removeColumn1)
				right := getTemplate("2", &pb.AccessOperation{Type: pb.AccessOperation_READ}, removeColumn1)
				expected := &pb.PoliciesDecisions{DatasetDecisions: []*pb.DatasetDecision{
					{
						Dataset: &pb.DatasetIdentifier{DatasetId: "1"},
						Decisions: []*pb.OperationDecision{
							{
								Operation:          &pb.AccessOperation{Type: pb.AccessOperation_READ},
								EnforcementActions: []*pb.EnforcementAction{removeColumn1},
							},
						},
					},
					{
						Dataset: &pb.DatasetIdentifier{DatasetId: "2"},
						Decisions: []*pb.OperationDecision{
							{
								Operation:          &pb.AccessOperation{Type: pb.AccessOperation_READ},
								EnforcementActions: []*pb.EnforcementAction{removeColumn1},
							},
						},
					},
				}}
				Expect(clients.MergePoliciesDecisions(left, right)).To(Equal(expected))
			})
		})

	})

})
