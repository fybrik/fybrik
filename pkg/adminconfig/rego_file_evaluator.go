// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import (
	"context"

	"github.com/open-policy-agent/opa/rego"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"sigs.k8s.io/yaml"

	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/logging"
)

// RegoPolicyEvaluator implements EvaluatorInterface
type RegoPolicyEvaluator struct {
	Log   zerolog.Logger
	Query rego.PreparedEvalQuery
}

// NewRegoPolicyEvaluator constructs a new RegoPolicyEvaluator object
func NewRegoPolicyEvaluator(log zerolog.Logger, query rego.PreparedEvalQuery) *RegoPolicyEvaluator {
	return &RegoPolicyEvaluator{
		Log:   log,
		Query: query,
	}
}

// Evaluate method evaluates the rego files based on the dynamic input object
func (r *RegoPolicyEvaluator) Evaluate(in *EvaluatorInput) (EvaluatorOutput, error) {
	log := r.Log.With().Str(utils.FybrikAppUUID, in.Workload.UUID).Logger()
	input, err := r.prepareInputForOPA(in)

	if err != nil {
		return EvaluatorOutput{Valid: false}, errors.Wrap(err, "failed to prepare an input for OPA")
	}
	// Run the evaluation with the new input
	rs, err := r.Query.Eval(context.Background(), rego.EvalInput(input))
	if err != nil {
		return EvaluatorOutput{Valid: false}, errors.Wrap(err, "failed to evaluate a query")
	}
	logging.LogStructure("Admin policy evaluation", &rs, log, false, true)
	// merge decisions and build an output object for the manager
	out := EvaluatorOutput{
		DatasetID:       in.Request.DatasetID,
		PolicySetID:     in.Workload.PolicySetID,
		UUID:            in.Workload.UUID,
		ConfigDecisions: DecisionPerCapabilityMap{},
		Policies:        []DecisionPolicy{},
	}
	err = r.getOPADecisions(in, rs, &out)
	return out, err
}

// prepares an input in OPA format
func (r *RegoPolicyEvaluator) prepareInputForOPA(in *EvaluatorInput) (map[string]interface{}, error) {
	log := r.Log.With().Str(utils.FybrikAppUUID, in.Workload.UUID).Logger()
	logging.LogStructure("Evaluator Input", in, log, false, false)
	var input map[string]interface{}
	bytes, err := yaml.Marshal(in)
	if err != nil {
		return input, errors.Wrap(err, "failed to marshal the input structure")
	}
	err = yaml.Unmarshal(bytes, &input)
	return input, errors.Wrap(err, "failed  to unmarshal the input structure")
}

// getOPADecisions parses the OPA decisions and merges decisions for the same capability
func (r *RegoPolicyEvaluator) getOPADecisions(in *EvaluatorInput, rs rego.ResultSet, out *EvaluatorOutput) error {
	out.Valid = false
	if len(rs) == 0 {
		return errors.New("invalid opa evaluation - an empty result set has been received")
	}
	for _, result := range rs {
		for _, expr := range result.Expressions {
			bytes, err := yaml.Marshal(expr.Value)
			if err != nil {
				return err
			}
			evalStruct := EvaluationOutputStructure{}
			if err = yaml.Unmarshal(bytes, &evalStruct); err != nil {
				return errors.Wrap(err, "Unexpected OPA response structure")
			}
			if !r.processConfigDecisions(&evalStruct, in, out) {
				return nil
			}
			r.processOptimizeDecisions(&evalStruct, in, out)
		}
	}
	out.Valid = true
	return nil
}

// merge config decisions
// return true if there is no conflict
func (r *RegoPolicyEvaluator) processConfigDecisions(evalStruct *EvaluationOutputStructure, in *EvaluatorInput, out *EvaluatorOutput) bool {
	log := r.Log.With().Str(utils.FybrikAppUUID, in.Workload.UUID).Logger()
	for _, rule := range evalStruct.Config {
		capability := rule.Capability
		newDecision := rule.Decision
		// filter by policySetID
		if newDecision.Policy.PolicySetID != "" && in.Workload.PolicySetID != "" && newDecision.Policy.PolicySetID != in.Workload.PolicySetID {
			continue
		}
		// apply defaults for undefined fields
		if newDecision.Deploy == "" {
			newDecision.Deploy = StatusUnknown
		}
		// a single decision should be made for a capability
		decision, exists := out.ConfigDecisions[capability]
		out.Policies = append(out.Policies, newDecision.Policy)
		if !exists {
			out.ConfigDecisions[capability] = newDecision
		} else {
			valid, mergedDecision := r.merge(newDecision, decision)
			if !valid {
				log.Error().Msg("Conflict while merging OPA decisions")
				logging.LogStructure("Conflicting decisions", out, log, true, true)
				return false
			}
			out.ConfigDecisions[capability] = mergedDecision
		}
	}
	return true
}

func (r *RegoPolicyEvaluator) processOptimizeDecisions(evalStruct *EvaluationOutputStructure, in *EvaluatorInput, out *EvaluatorOutput) {
	out.OptimizationStrategy = []AttributeOptimization{}
	if len(evalStruct.Optimize) > 0 {
		// choose the first optimization strategy
		// TODO(shlomitk1): add priorities to optimization strategies
		rule := evalStruct.Optimize[0]
		out.OptimizationStrategy = append(out.OptimizationStrategy, rule.Strategy...)
		out.Policies = append(out.Policies, rule.Policy)
	}
}

// This function merges two decisions for the same capability using the following logic:
// deploy: true/false take precedence over undefined, true and false result in a conflict.
// restrictions: new pairs <key, value> are added, if both exist - compatibility is checked.
// policy: concatenation of IDs and descriptions.
func (r *RegoPolicyEvaluator) merge(newDecision, oldDecision Decision) (bool, Decision) {
	mergedDecision := Decision{}
	// merge deployment decisions
	deploy := oldDecision.Deploy
	if deploy == StatusUnknown {
		deploy = newDecision.Deploy
	} else if newDecision.Deploy != StatusUnknown {
		if newDecision.Deploy != deploy {
			return false, mergedDecision
		}
	}
	mergedDecision.Deploy = deploy
	// merge restrictions
	mergedDecision.DeploymentRestrictions = oldDecision.DeploymentRestrictions
	mergedDecision.DeploymentRestrictions.Clusters = append(mergedDecision.DeploymentRestrictions.Clusters,
		newDecision.DeploymentRestrictions.Clusters...)
	mergedDecision.DeploymentRestrictions.Modules = append(mergedDecision.DeploymentRestrictions.Modules,
		newDecision.DeploymentRestrictions.Modules...)
	mergedDecision.DeploymentRestrictions.StorageAccounts = append(mergedDecision.DeploymentRestrictions.StorageAccounts,
		newDecision.DeploymentRestrictions.StorageAccounts...)
	// policies are appended to the output, no need to merge
	mergedDecision.Policy = DecisionPolicy{}
	return true, mergedDecision
}
