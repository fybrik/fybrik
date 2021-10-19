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

func GetDataAccessModuleNamespace() string {
	ns := os.Getenv("DATA_ACCESS_MODULE_NAMESPACE")
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
	return GetDataAccessModuleNamespace()
}

func GetStreamTransferNamespace() string {
	return GetDataAccessModuleNamespace()
}
