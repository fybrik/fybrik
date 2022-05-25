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
func solve(env *datapath.Environment, datasetInfo *datapath.DataInfo, log *zerolog.Logger) (datapath.Solution, error) {
	cspPath := utils.GetCSPPath()
	if utils.UseCSP() && cspPath != "" {
		cspOptimizer := optimizer.NewOptimizer(env, datasetInfo, cspPath, log)
		return cspOptimizer.Solve()
	}
	pathBuilder := PathBuilder{Log: log, Env: env, Asset: datasetInfo}
	return pathBuilder.solve()
}
