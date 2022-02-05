// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import (
	"context"
	"encoding/json"

	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/adminrules"
	"fybrik.io/fybrik/pkg/taxonomy/validate"
	"github.com/open-policy-agent/opa/rego"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

const ValidationPath string = "/tmp/taxonomy/adminrules.json#/definitions/EvaluationOutputStructure"

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
		Policies:        []adminrules.DecisionPolicy{},
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
	log := r.Log.With().Str(utils.FybrikAppUUID, in.Workload.UUID).Logger()
	out.Valid = false
	if len(rs) == 0 {
		return errors.New("invalid opa evaluation - an empty result set has been received")
	}
	for _, result := range rs {
		for _, expr := range result.Expressions {
			if err := validateStructure(expr.Value, ValidationPath, in.Workload.UUID); err != nil {
				return err
			}
			bytes, err := yaml.Marshal(expr.Value)
			if err != nil {
				return err
			}
			evalStruct := adminrules.EvaluationOutputStructure{}
			if err = yaml.Unmarshal(bytes, &evalStruct); err != nil {
				return errors.Wrap(err, "Unexpected OPA response structure")
			}
			for _, rule := range evalStruct.Config {
				capability := rule.Capability
				newDecision := rule.Decision
				// filter by policySetID
				if newDecision.Policy.PolicySetID != "" && in.Workload.PolicySetID != "" && newDecision.Policy.PolicySetID != in.Workload.PolicySetID {
					continue
				}
				// apply defaults for undefined fields
				if newDecision.Deploy == "" {
					newDecision.Deploy = adminrules.StatusUnknown
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
						return nil
					}
					out.ConfigDecisions[capability] = mergedDecision
				}
			}
		}
	}
	out.Valid = true
	return nil
}

// This function merges two decisions for the same capability using the following logic:
// deploy: true/false take precedence over undefined, true and false result in a conflict.
// restrictions: new pairs <key, value> are added, if both exist - compatibility is checked.
// policy: concatenation of IDs and descriptions.
func (r *RegoPolicyEvaluator) merge(newDecision adminrules.Decision, oldDecision adminrules.Decision) (bool, adminrules.Decision) {
	mergedDecision := adminrules.Decision{}
	// merge deployment decisions
	deploy := oldDecision.Deploy
	if deploy == adminrules.StatusUnknown {
		deploy = newDecision.Deploy
	} else if newDecision.Deploy != adminrules.StatusUnknown {
		if newDecision.Deploy != deploy {
			return false, mergedDecision
		}
	}
	mergedDecision.Deploy = deploy
	// merge restrictions
	mergedDecision.DeploymentRestrictions = oldDecision.DeploymentRestrictions
	if mergedDecision.DeploymentRestrictions.Clusters == nil {
		mergedDecision.DeploymentRestrictions.Clusters = make(adminrules.Restriction)
	}
	if err := mergeRestrictions(&mergedDecision.DeploymentRestrictions.Clusters, &newDecision.DeploymentRestrictions.Clusters); err != nil {
		return false, adminrules.Decision{}
	}
	if mergedDecision.DeploymentRestrictions.Modules == nil {
		mergedDecision.DeploymentRestrictions.Modules = make(adminrules.Restriction)
	}
	if err := mergeRestrictions(&mergedDecision.DeploymentRestrictions.Modules, &newDecision.DeploymentRestrictions.Modules); err != nil {
		return false, adminrules.Decision{}
	}
	if mergedDecision.DeploymentRestrictions.StorageAccounts == nil {
		mergedDecision.DeploymentRestrictions.StorageAccounts = make(adminrules.Restriction)
	}

	if err := mergeRestrictions(&mergedDecision.DeploymentRestrictions.StorageAccounts, &newDecision.DeploymentRestrictions.StorageAccounts); err != nil {
		return false, adminrules.Decision{}
	}
	// policies are appended to the output, no need to merge
	mergedDecision.Policy = adminrules.DecisionPolicy{}
	return true, mergedDecision
}

func mergeRestrictions(r1 *adminrules.Restriction, r2 *adminrules.Restriction) error {
	if r2 == nil {
		return nil
	}
	for key, values := range *r2 {
		if len((*r1)[key]) == 0 {
			(*r1)[key] = values
		} else {
			(*r1)[key] = utils.Intersection((*r1)[key], values)
			if len((*r1)[key]) == 0 {
				return errors.New("unable to merge restrictions")
			}
		}
	}
	return nil
}

func validateStructure(obj interface{}, taxonomySchema string, uuid string) error {
	// validate against taxonomy
	bytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	allErrs, err := validate.TaxonomyCheck(bytes, ValidationPath)
	if err != nil {
		return err
	}
	if len(allErrs) != 0 {
		return apierrors.NewInvalid(
			schema.GroupKind{Group: "app.fybrik.io", Kind: "ConfigurationPolicies-ExpressionValue"},
			uuid, allErrs)
	}
	return nil
}
