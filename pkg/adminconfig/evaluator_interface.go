// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

// EvaluatorInterface is an interface for config policies' evaluator
type EvaluatorInterface interface {
	Evaluate(in *EvaluatorInput) (EvaluatorOutput, error)
}
