// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/rs/zerolog"

	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/datapath"
	"fybrik.io/fybrik/pkg/optimizer"
)

// find a solution for a data path
// satisfying governance and admin policies
// with respect to the optimization strategy
func solveSingleDataset(env *datapath.Environment, dataset *datapath.DataInfo, log *zerolog.Logger) (datapath.Solution, error) {
	cspPath := utils.GetCSPPath()
	if utils.UseCSP() && cspPath != "" {
		cspOptimizer := optimizer.NewOptimizer(env, dataset, cspPath, log)
		return cspOptimizer.Solve()
	}
	pathBuilder := PathBuilder{Log: log, Env: env, Asset: dataset}
	return pathBuilder.solve()
}

// find a solution for all data paths at once
func solve(env *datapath.Environment, datasets []datapath.DataInfo, log *zerolog.Logger) ([]datapath.Solution, error) {
	solutions := []datapath.Solution{}
	for i := range datasets {
		solution, err := solveSingleDataset(env, &datasets[i], log)
		if err != nil {
			return solutions, err
		}
		solutions = append(solutions, solution)
	}
	return solutions, nil
}
