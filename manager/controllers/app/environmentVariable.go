package app

import (
	"os"
)

// BlueprintNamespace defines a namespace where blueprints and associated resources will be allocated
const DefaultBlueprintNamespace = "fybrik-blueprints"

// Controller namespace defines a namespace where
const DefaultControllerNamespace = "fybrik-system"

func getBlueprintNamespace() string {

	blueprintNamespace := os.Getenv("BLUEPRINT_NAMESPACE")
	if len(blueprintNamespace) <= 0 {
		blueprintNamespace = DefaultBlueprintNamespace
	}
	return blueprintNamespace
}

func getControllerNamespace() string {

	controllerNamespace := os.Getenv("CONTROLLER_NAMESPACE")
	if len(controllerNamespace) <= 0 {
		controllerNamespace = DefaultControllerNamespace
	}
	return controllerNamespace
}
