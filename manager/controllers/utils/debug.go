// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
)

// PrintStructure prints the structure in a textual format
func PrintStructure(argStruct interface{}, log logr.Logger, argName string) {
	log.Info(argName + ":")
	yaml, err := yaml.Marshal(argStruct)
	if err != nil {
		log.Info("\t Error printing " + argName + "\n")
		return
	}
	log.Info("\t" + string(yaml))
}
