// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import (
	"context"
	"sync"

	"github.com/open-policy-agent/opa/rego"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"sigs.k8s.io/yaml"

	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/monitor"
)

// RegoPolicyEvaluator implements EvaluatorInterface
type RegoPolicyEvaluator struct {
	Log   zerolog.Logger
	Query rego.PreparedEvalQuery
	Mux   *sync.RWMutex
}

// NewRegoPolicyEvaluator constructs a new RegoPolicyEvaluator object
func NewRegoPolicyEvaluator() (*RegoPolicyEvaluator, error) {
	logger := logging.LogInit(logging.CONTROLLER, "ConfigPolicyEvaluator")
	// pre-compiling config policy files
	query, err := PrepareQuery()
	if err != nil {
		return nil, err
	}

	return &RegoPolicyEvaluator{
		Log:   logger,
		Query: query,
		Mux:   &sync.RWMutex{},
	}, nil
}

func NewRegoPolicyEvaluatorWithQuery(query rego.PreparedEvalQuery) *RegoPolicyEvaluator {
	return &RegoPolicyEvaluator{
		Log:   logging.LogInit(logging.CONTROLLER, "ConfigPolicyEvaluator"),
		Query: query,
		Mux:   &sync.RWMutex{},
	}
}

func (r *RegoPolicyEvaluator) OnError(err error) {
	r.Log.Error().Err(err).Msg("Error compiling the policies")
}

// Options for file monitor including the monitored directory and the relevant file extension
func (r *RegoPolicyEvaluator) GetOptions() monitor.FileMonitorOptions {
	return monitor.FileMonitorOptions{Path: RegoPolicyDirectory, Extension: ".rego"}
}

// notification event: policy files have been changed
func (r *RegoPolicyEvaluator) OnNotify() {
	query, err := PrepareQuery()
	if err != nil {
		r.OnError(err)
	}
	r.Mux.Lock()
	r.Query = query
	r.Mux.Unlock()
}

// Evaluate method evaluates the rego files based on the dynamic input object
func (r *RegoPolicyEvaluator) Evaluate(in *EvaluatorInput) (EvaluatorOutput, error) {
	logger := r.Log.With().Str(utils.FybrikAppUUID, in.Workload.UUID).Logger()
	input, err := r.prepareInputForOPA(in)

	if err != nil {
		return EvaluatorOutput{Valid: false}, errors.Wrap(err, "failed to prepare an input for OPA")
	}
	// Run the evaluation with the new input
	r.Mux.RLock()
	rs, err := r.Query.Eval(context.Background(), rego.EvalInput(input))
	r.Mux.RUnlock()
	if err != nil {
		return EvaluatorOutput{Valid: false}, errors.Wrap(err, "failed to evaluate a query")
	}
	logging.LogStructure("Admin policy evaluation", &rs, &logger, zerolog.DebugLevel, false, true)
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
	logging.LogStructure("Evaluator Input", in, &log, zerolog.DebugLevel, false, false)
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
			r.processOptimizeDecisions(&evalStruct, out)
		}
	}
	out.Valid = true
	return nil
}

// merge config decisions
// return true if there is no conflict
func (r *RegoPolicyEvaluator) processConfigDecisions(evalStruct *EvaluationOutputStructure, in *EvaluatorInput, out *EvaluatorOutput) bool {
	log := r.Log.With().Str(utils.FybrikAppUUID, in.Workload.UUID).Logger()
	for ind := range evalStruct.Config {
		rule := &evalStruct.Config[ind]
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
			valid, mergedDecision := r.merge(&newDecision, &decision)
			if !valid {
				log.Error().Msg("Conflict while merging OPA decisions")
				logging.LogStructure("Conflicting decisions", out, &log, zerolog.ErrorLevel, true, true)
				return false
			}
			out.ConfigDecisions[capability] = mergedDecision
		}
	}
	return true
}

func (r *RegoPolicyEvaluator) processOptimizeDecisions(evalStruct *EvaluationOutputStructure, out *EvaluatorOutput) {
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
func (r *RegoPolicyEvaluator) merge(newDecision, oldDecision *Decision) (bool, Decision) {
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
