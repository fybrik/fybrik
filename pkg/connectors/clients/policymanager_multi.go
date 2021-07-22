// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"fmt"

	"emperror.dev/errors"
	pb "github.com/mesh-for-data/mesh-for-data/pkg/connectors/protobuf"
	"go.uber.org/multierr"
)

// NewMultiPolicyManager creates a PolicyManager facade that combines results from multiple policy managers
// You must call .Close() when you are done using the created instance
func NewMultiPolicyManager(managers ...PolicyManager) PolicyManager {
	return &multiPolicyManager{
		managers: managers,
	}
}

var _ PolicyManager = (*multiPolicyManager)(nil)

type multiPolicyManager struct {
	pb.UnimplementedPolicyManagerServiceServer

	managers []PolicyManager
}

func (m *multiPolicyManager) GetPoliciesDecisions(ctx context.Context, in *pb.ApplicationContext) (*pb.PoliciesDecisions, error) {
	var allErr error
	decisionsList := []*pb.PoliciesDecisions{}

	for _, manager := range m.managers {
		decisions, err := manager.GetPoliciesDecisions(ctx, in)
		if !multierr.AppendInto(&allErr, err) {
			if decisions != nil {
				decisionsList = append(decisionsList, decisions)
			}
		}
	}

	// if len(decisionsList) == 0 {
	// 	return nil, fmt.Errorf("received no policy manager decisions")
	// }

	result := MergePoliciesDecisions(decisionsList...)
	return result, errors.Wrap(allErr, fmt.Sprintf("multi policy manager returned %d errors", len(multierr.Errors(allErr))))
}

func (m *multiPolicyManager) Close() error {
	var err error
	for _, manager := range m.managers {
		multierr.AppendInto(&err, manager.Close())
	}
	return err
}
