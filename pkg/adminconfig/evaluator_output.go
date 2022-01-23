// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import "fybrik.io/fybrik/pkg/model/taxonomy"

// EvaluatorOutput is an output of ConfigurationPoliciesEvaluator.
// Used by manager to decide which modules are deployed and in which cluster.
type EvaluatorOutput struct {
	// Valid is true when there is no conflict between the decisions, and false otherwise
	Valid bool
	// Dataset identifier
	DatasetID string
	// Unique fybrikapplication id used for logging
	UUID string
	// Policy set id used in the evaluation
	PolicySetID string
	// Decisions per capability (after being merged)
	ConfigDecisions taxonomy.DecisionPerCapabilityMap
}
