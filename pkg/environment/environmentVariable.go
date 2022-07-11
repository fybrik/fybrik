// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package environment

import (
	"os"
)

// DefaultModulesNamespace defines a default namespace where module resources will be allocated
const DefaultModulesNamespace = "fybrik-blueprints"

// DefaultControllerNamespace defines a default namespace where fybrik control plane is running
const DefaultControllerNamespace = "fybrik-system"

func GetDefaultModulesNamespace() string {
	ns := os.Getenv("MODULES_NAMESPACE")
	if ns == "" {
		ns = DefaultModulesNamespace
	}
	return ns
}

func GetControllerNamespace() string {
	controllerNamespace := os.Getenv("CONTROLLER_NAMESPACE")
	if controllerNamespace == "" {
		controllerNamespace = DefaultControllerNamespace
	}
	return controllerNamespace
}

func GetApplicationNamespace() string {
	return os.Getenv("APPLICATION_NAMESPACE")
}
