// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

func TestValidModuleWithTaxonomy(t *testing.T) {
	t.Parallel()

	filename := "../../../testdata/unittests/fybrikmodule-validActions.yaml"
	moduleYaml, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	fybrikModule := &FybrikModule{}
	err = yaml.Unmarshal(moduleYaml, fybrikModule)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	taxonomyFile := "../../../testdata/unittests/sampletaxonomy/fybrik_module.json"
	validateErr := fybrikModule.ValidateFybrikModule(taxonomyFile)
	assert.Nil(t, validateErr, "No error should be found")
}

func TestInvalidModuleWithTaxonomy(t *testing.T) {
	t.Parallel()

	filename := "../../../testdata/unittests/fybrikmodule-actionsErrors.yaml"
	buf, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	fybrikModule := &FybrikModule{}
	err = yaml.Unmarshal(buf, fybrikModule)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	taxonomyFile := "../../../testdata/unittests/sampletaxonomy/fybrik_module.json"
	validateErr := fybrikModule.ValidateFybrikModule(taxonomyFile)
	assert.NotNil(t, validateErr, "Invalid actions error should be found")
}
