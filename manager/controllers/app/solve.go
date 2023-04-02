// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"emperror.dev/errors"
	"github.com/rs/zerolog"

	"fybrik.io/fybrik/pkg/datapath"
	"fybrik.io/fybrik/pkg/environment"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/optimizer"
)

// find a solution for a data path
// satisfying governance and admin policies
// with respect to the optimization strategy
func solveSingleDataset(env *datapath.Environment, dataset *datapath.DataInfo, log *zerolog.Logger) (datapath.Solution, error) {
	cspPath := environment.GetCSPPath()
	if environment.UseCSP() && cspPath != "" {
		cspOptimizer := optimizer.NewOptimizer(env, dataset, cspPath, log)
		solution, err := cspOptimizer.Solve()
		if err == nil {
			if len(solution.DataPath) > 0 { // solver found a solution
				return solution, nil
			}
			if len(solution.DataPath) == 0 { // solver returned UNSAT
				msg := "Data path cannot be constructed given the deployed modules and the active restrictions"
				log.Error().Str(logging.DATASETID, dataset.Context.DataSetID).Msg(msg)
				logging.LogStructure("Data Item Context", dataset, log, zerolog.TraceLevel, true, true)
				logging.LogStructure("Module Map", env.Modules, log, zerolog.TraceLevel, true, true)
				return datapath.Solution{}, errors.New(msg + " for " + dataset.Context.DataSetID)
			}
		} else {
			msg := "Error solving CSP. Fybrik will now search for a solution without considering optimization goals."
			log.Error().Err(err).Str(logging.DATASETID, dataset.Context.DataSetID).Msg(msg)
			// now fallback to finding a non-optimized solution
		}
	}
	pathBuilder := PathBuilder{Log: log, Env: env, Asset: dataset}
	return pathBuilder.solve()
}

// find a solution for all data paths at once
func solve(env *datapath.Environment, datasets []datapath.DataInfo, log *zerolog.Logger) ([]datapath.Solution, error) {
	solutions := []datapath.Solution{}
	if len(env.Modules) == 0 {
		msg := "There are no deployed modules in the environment"
		log.Error().Msg(msg)
		return solutions, errors.New(msg)
	}
	for i := range datasets {
		solution, err := solveSingleDataset(env, &datasets[i], log)
		if err != nil {
			return solutions, err
		}
		solutions = append(solutions, solution)
	}
	return solutions, nil
}
