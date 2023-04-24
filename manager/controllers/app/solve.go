// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	"emperror.dev/errors"
	"github.com/rs/zerolog"

	"fybrik.io/fybrik/pkg/datapath"
	"fybrik.io/fybrik/pkg/environment"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/taxonomy"
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
	if err := validateBasicConditions(env, datasets, log); err != nil {
		return solutions, err
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

// perform basic checks before searching for a solution for a dataset
func validateBasicConditions(env *datapath.Environment, datasets []datapath.DataInfo, log *zerolog.Logger) error {
	if len(env.Modules) == 0 {
		log.Error().Msg(NoDeployedModules)
		return errors.New(NoDeployedModules)
	}
	for i := range datasets {
		dataset := &datasets[i]
		if dataset.Context.Flow == "" || dataset.Context.Flow == taxonomy.ReadFlow {
			if err := validateApplicationProtocol(env, dataset); err != nil {
				log.Error().Err(err)
				return err
			}
			if err := validateAssetProtocol(env, dataset); err != nil {
				log.Error().Err(err)
				return err
			}
		}
	}
	return nil
}

// create interface string to print in error messages
func createInterfaceString(interfacePtr *taxonomy.Interface) string {
	interfaceStr := string(interfacePtr.Protocol)
	if interfacePtr.DataFormat != "" {
		interfaceStr = interfaceStr + ", " + string(interfacePtr.DataFormat)
	}
	return interfaceStr
}

// check if any deployed module provides the requested read api by the application
// return nil if such module exists, and an error if not
func validateApplicationProtocol(env *datapath.Environment, dataset *datapath.DataInfo) error {
	applicationInterfacePtr := dataset.Context.Requirements.Interface
	for _, module := range env.Modules {
		for _, capability := range module.Spec.Capabilities {
			// check if the module capability matches the application protocol requirement
			if capability.API == nil {
				continue
			}
			capabilityInterfacePtr := &taxonomy.Interface{Protocol: capability.API.Connection.Name, DataFormat: capability.API.DataFormat}
			if match(capabilityInterfacePtr, applicationInterfacePtr) {
				return nil
			}
		}
	}
	message := fmt.Sprintf("The requested interface (%s) is not supported by the deployed modules for dataset '%s'",
		createInterfaceString(applicationInterfacePtr), dataset.Context.DataSetID)
	return errors.New(message)
}

// check if any deployed module provides the connection to read the asset
// return nil if such module exists, and an error if not
func validateAssetProtocol(env *datapath.Environment, dataset *datapath.DataInfo) error {
	assetConnection := dataset.DataDetails.Details.Connection.Name
	assetDataformat := dataset.DataDetails.Details.DataFormat
	assetInterfacePtr := &taxonomy.Interface{Protocol: assetConnection, DataFormat: assetDataformat}
	for _, module := range env.Modules {
		for _, capability := range module.Spec.Capabilities {
			// check if the module capability matches the asset connection requirement
			for _, capabilityInterface := range capability.SupportedInterfaces {
				if match(capabilityInterface.Source, assetInterfacePtr) {
					return nil
				}
			}
		}
	}
	message := fmt.Sprintf("The asset '%s' (%s) can't be read by the deployed modules",
		dataset.Context.DataSetID, createInterfaceString(assetInterfacePtr))
	return errors.New(message)
}
