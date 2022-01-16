// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/logging"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// Type definitions for parsing OPA response
// A list of decisions per capability, e.g. {"read": {"deploy": true}, "write": {"deploy": false}}
type RuleDecisionList []DecisionPerCapabilityMap

// A directory containing rego files that define admin config policies
const RegoPolicyDirectory = "/tmp/adminconfig/"

// RegoPolicyEvaluator implements EvaluatorInterface
type RegoPolicyEvaluator struct {
	Log          zerolog.Logger
	Query        rego.PreparedEvalQuery
	ReadyForEval bool
	Timestamp    time.Time
}

// NewRegoPolicyEvaluator constructs a new RegoPolicyEvaluator object
func NewRegoPolicyEvaluator(log zerolog.Logger) *RegoPolicyEvaluator {
	return &RegoPolicyEvaluator{
		Log:          log,
		Query:        rego.PreparedEvalQuery{},
		ReadyForEval: false,
		Timestamp:    time.Now(),
	}
}

// isRecent checks whether the compilation timestamp is not behind the file update/create time
func (r *RegoPolicyEvaluator) isRecent() (bool, error) {
	entries, err := os.ReadDir(RegoPolicyDirectory)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		fileInfo, err := entry.Info()
		if err != nil {
			return false, err
		}
		if fileInfo.ModTime().After(r.Timestamp) {
			return false, nil
		}
	}
	return true, nil
}

// prepareQuery prepares a query for OPA evaluation - data object and compiled modules.
// This function is called upon the change in rego files.
// Monitoring changes in rego files will be implemented in the future version.
func (r *RegoPolicyEvaluator) prepareQuery() (rego.PreparedEvalQuery, error) {
	// read and compile rego files
	files, err := os.ReadDir(RegoPolicyDirectory)
	if err != nil {
		return rego.PreparedEvalQuery{}, err
	}
	modules := map[string]string{}
	for _, info := range files {
		name := info.Name()
		if !strings.HasSuffix(name, ".rego") {
			continue
		}
		fileName := filepath.Join(RegoPolicyDirectory, name)
		module, err := os.ReadFile(filepath.Clean(fileName))
		if err != nil {
			return rego.PreparedEvalQuery{}, err
		}
		modules[name] = string(module)
	}
	r.Timestamp = time.Now()
	compiler, err := ast.CompileModules(modules)

	if err != nil {
		return rego.PreparedEvalQuery{}, errors.Wrap(err, "couldn't compile modules")
	}
	rego := rego.New(
		rego.Query("data.adminconfig.config"),
		rego.Compiler(compiler),
	)
	return rego.PrepareForEval(context.Background())
}

// Evaluate method evaluates the rego files based on the dynamic input object
func (r *RegoPolicyEvaluator) Evaluate(in *EvaluatorInput) (EvaluatorOutput, error) {
	log := r.Log.With().Str(utils.FybrikAppUUID, in.Workload.UUID).Logger()
	recompile := false
	if !r.ReadyForEval {
		recompile = true
	} else {
		recent, err := r.isRecent()
		if err != nil {
			log.Error().Err(err).Msg("")
			return EvaluatorOutput{Valid: false}, err
		}
		if !recent {
			log.Info().Msg("recompiling since newer files have been found")
			recompile = true
		}
	}
	if recompile {
		var err error
		if r.Query, err = r.prepareQuery(); err != nil {
			return EvaluatorOutput{Valid: false}, errors.Wrap(err, "failed to prepare a query")
		}
		r.ReadyForEval = true
	}
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
	decisions, valid, err := r.getOPADecisions(in, rs)
	if err != nil {
		return EvaluatorOutput{Valid: valid, DatasetID: in.Request.DatasetID, ConfigDecisions: decisions}, err
	}
	return EvaluatorOutput{
		Valid:           valid,
		DatasetID:       in.Request.DatasetID,
		PolicySetID:     in.Workload.PolicySetID,
		UUID:            in.Workload.UUID,
		ConfigDecisions: decisions,
	}, nil
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
func (r *RegoPolicyEvaluator) getOPADecisions(in *EvaluatorInput, rs rego.ResultSet) (DecisionPerCapabilityMap, bool, error) {
	log := r.Log.With().Str(utils.FybrikAppUUID, in.Workload.UUID).Logger()
	decisions := DecisionPerCapabilityMap{}
	if len(rs) == 0 {
		return decisions, false, errors.New("invalid opa evaluation - an empty result set has been received")
	}
	for _, result := range rs {
		for _, expr := range result.Expressions {
			bytes, err := yaml.Marshal(expr.Value)
			if err != nil {
				return nil, false, err
			}
			ruleDecisions := RuleDecisionList{}
			if err = yaml.Unmarshal(bytes, &ruleDecisions); err != nil {
				return nil, false, errors.Wrap(err, "Unexpected OPA response structure")
			}
			for _, rule := range ruleDecisions {
				for capability, newDecision := range rule {
					// filter by policySetID
					if newDecision.Policy.PolicySetID != "" && in.Workload.PolicySetID != "" && newDecision.Policy.PolicySetID != in.Workload.PolicySetID {
						continue
					}
					// apply defaults for undefined fields
					// string -> ConditionStatus conversion
					switch newDecision.Deploy {
					case "true":
						newDecision.Deploy = corev1.ConditionTrue
					case "false":
						newDecision.Deploy = corev1.ConditionFalse
					case "":
						newDecision.Deploy = corev1.ConditionUnknown
					default:
						return nil, false, errors.New("Illegal value for Deploy: " + string(newDecision.Deploy))
					}
					// a single decision should be made for a capability
					decision, exists := decisions[capability]
					if !exists {
						decisions[capability] = newDecision
					} else {
						valid, mergedDecision := r.merge(newDecision, decision)
						if !valid {
							joinedStr := strings.Join([]string{decision.Policy.Description, newDecision.Policy.Description}, ";")
							log.Error().Str("decisions", joinedStr).Msg("Conflict while merging OPA decisions")
							return decisions, false, nil
						}
						decisions[capability] = mergedDecision
					}
				}
			}
		}
	}
	return decisions, true, nil
}

// This function merges two decisions for the same capability using the following logic:
// deploy: true/false take precedence over undefined, true and false result in a conflict.
// restrictions: new pairs <key, value> are added, if both exist - compatibility is checked.
// policy: concatenation of IDs and descriptions.
func (r *RegoPolicyEvaluator) merge(newDecision Decision, oldDecision Decision) (bool, Decision) {
	mergedDecision := Decision{}
	// merge deployment decisions
	deploy := oldDecision.Deploy
	if deploy == corev1.ConditionUnknown {
		deploy = newDecision.Deploy
	} else if newDecision.Deploy != corev1.ConditionUnknown {
		if newDecision.Deploy != deploy {
			return false, mergedDecision
		}
	}
	mergedDecision.Deploy = deploy
	// merge restrictions
	mergedDecision.DeploymentRestrictions = oldDecision.DeploymentRestrictions
	if mergedDecision.DeploymentRestrictions == nil {
		mergedDecision.DeploymentRestrictions = make(Restrictions)
	}
	for entity, restrictions := range newDecision.DeploymentRestrictions {
		if mergedRestriction, found := mergedDecision.DeploymentRestrictions[entity]; !found {
			mergedDecision.DeploymentRestrictions[entity] = restrictions
		} else {
			for key, values := range restrictions {
				if len(mergedRestriction[key]) == 0 {
					mergedRestriction[key] = values
				} else {
					mergedRestriction[key] = utils.Intersection(mergedRestriction[key], values)
					if len(mergedRestriction[key]) == 0 {
						return false, mergedDecision
					}
				}
			}
			mergedDecision.DeploymentRestrictions[entity] = mergedRestriction
		}
	}
	// merge policies descriptions/ids
	mergedDecision.Policy = oldDecision.Policy
	if mergedDecision.Policy.ID != "" {
		mergedDecision.Policy.ID += ";"
	}
	mergedDecision.Policy.ID += newDecision.Policy.ID
	if mergedDecision.Policy.Description != "" {
		mergedDecision.Policy.Description += ";"
	}
	mergedDecision.Policy.Description += newDecision.Policy.Description
	return true, mergedDecision
}
