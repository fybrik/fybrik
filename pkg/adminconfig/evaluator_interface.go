// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import "github.com/rs/zerolog"

// EvaluatorInterface is an interface for config policies' evaluator
type EvaluatorInterface interface {
	Evaluate(in *EvaluatorInput, log zerolog.Logger) (EvaluatorOutput, error)
}
