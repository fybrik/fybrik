// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"os"
)

// BlueprintNamespace defines a namespace where blueprints and associated resources will be allocated
const DefaultBlueprintNamespace = "fybrik-blueprints"

// Controller namespace defines a namespace where
const DefaultControllerNamespace = "fybrik-system"

func GetBlueprintNamespace() string {
	blueprintNamespace := os.Getenv("BLUEPRINT_NAMESPACE")
	if len(blueprintNamespace) == 0 {
		blueprintNamespace = DefaultBlueprintNamespace
	}
	return blueprintNamespace
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
	return GetBlueprintNamespace()
}

func GetStreamTransferNamespace() string {
	return GetBlueprintNamespace()
}
