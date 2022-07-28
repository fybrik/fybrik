// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v12

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

func TestValidApplicationWithBaseTaxonomy(t *testing.T) {
	t.Parallel()

	filename := "../../../testdata/unittests/fybrikapplication-validForBase.yaml"
	applicationYaml, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	fybrikApp := &FybrikApplication{}
	err = yaml.Unmarshal(applicationYaml, fybrikApp)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	taxonomyFile := "../../../testdata/unittests/basetaxonomy/fybrik_application.json"
	validateErr := fybrikApp.ValidateFybrikApplication(taxonomyFile)
	assert.Nil(t, validateErr, "No error should be found")
}

func TestValidApplicationWithEnhancedTaxonomy(t *testing.T) {
	t.Parallel()

	filename := "../../../../samples/kubeflow/fybrikapplication.yaml"
	applicationYaml, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	fybrikApp := &FybrikApplication{}
	err = yaml.Unmarshal(applicationYaml, fybrikApp)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	taxonomyFile := "../../../testdata/unittests/sampletaxonomy/fybrik_application.json"
	validateErr := fybrikApp.ValidateFybrikApplication(taxonomyFile)
	assert.Nil(t, validateErr, "No error should be found")
}

func TestInvalidAppInfoWithEnhancedTaxonomy(t *testing.T) {
	t.Parallel()

	filename := "../../../testdata/unittests/fybrikapplication-appInfoErrors.yaml"
	buf, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	fybrikApp := &FybrikApplication{}
	err = yaml.Unmarshal(buf, fybrikApp)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	taxonomyFile := "../../../testdata/unittests/sampletaxonomy/fybrik_application.json"
	validateErr := (*fybrikApp).ValidateFybrikApplication(taxonomyFile)
	assert.NotNil(t, validateErr, "Invalid appInfo error should be found")
}

func TestInvalidInterfaceWithEnhancedTaxonomy(t *testing.T) {
	t.Parallel()

	filename := "../../../testdata/unittests/fybrikapplication-interfaceErrors.yaml"
	buf, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	fybrikApp := &FybrikApplication{}
	err = yaml.Unmarshal(buf, fybrikApp)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	taxonomyFile := "../../../testdata/unittests/sampletaxonomy/fybrik_application.json"
	validateErr := (*fybrikApp).ValidateFybrikApplication(taxonomyFile)
	assert.NotNil(t, validateErr, "Invalid interface error should be found")
}
