// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"os"
)

// DefaultDataAccessModuleNamespace defines a namespace where module resources will be allocated
const DefaultDataAccessModuleNamespace = "fybrik-blueprints"

// Controller namespace defines a namespace where
const DefaultControllerNamespace = "fybrik-system"

func GetDefaultModulesNamespace() string {
	ns := os.Getenv("MODULES_NAMESPACE")
	if len(ns) == 0 {
		ns = DefaultDataAccessModuleNamespace
	}
	return ns
}

func GetControllerNamespace() string {
	controllerNamespace := os.Getenv("CONTROLLER_NAMESPACE")
	if len(controllerNamespace) == 0 {
		controllerNamespace = DefaultControllerNamespace
	}
	return controllerNamespace
}

func GetApplicationNamespace() string {
	return os.Getenv("APPLICATION_NAMESPACE")
}

func GetBatchTransferNamespace() string {
	return GetDefaultModulesNamespace()
}

func GetStreamTransferNamespace() string {
	return GetDefaultModulesNamespace()
}
