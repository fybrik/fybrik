// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/rs/zerolog"

	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/optimizer"
)

// find a solution for a data path
// satisfying governance and admin policies
// with respect to the optimization strategy
func solve(env *optimizer.Environment, datasetInfo *optimizer.DataInfo, log *zerolog.Logger) (optimizer.Solution, error) {
	cspPath := utils.GetCSPPath()
	if cspPath != "" {
		cspOptimizer := optimizer.NewOptimizer(env, datasetInfo, cspPath)
		return cspOptimizer.Solve()
	}
	pathBuilder := PathBuilder{Log: log, Env: env, Asset: datasetInfo}
	return pathBuilder.solve()
}
