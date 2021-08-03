// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

func TestValidApplication(t *testing.T) {
	t.Parallel()

	filename := "../../../../samples/kubeflow/fybrikapplication.yaml"
	applicationYaml, err := ioutil.ReadFile(filename)
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
	taxonomyFile := "../../../../charts/fybrik/files/taxonomy/application.values.schema.json"
	validateErr := fybrikApp.ValidateFybrikApplication(taxonomyFile)
	assert.Nil(t, validateErr, "No error should be found")
}

func TestInvalidAppInfo(t *testing.T) {
	t.Parallel()

	filename := "../../../testdata/unittests/fybrikapplication-appInfoErrors.yaml"
	buf, err := ioutil.ReadFile(filename)
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

	taxonomyFile := "../../../../charts/fybrik/files/taxonomy/application.values.schema.json"
	validateErr := (*fybrikApp).ValidateFybrikApplication(taxonomyFile)
	assert.NotNil(t, validateErr, "Invalid appInfo error should be found")
}

func TestInvalidInterface(t *testing.T) {
	t.Parallel()

	filename := "../../../testdata/unittests/fybrikapplication-interfaceErrors.yaml"
	buf, err := ioutil.ReadFile(filename)
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

	taxonomyFile := "../../../../charts/fybrik/files/taxonomy/application.values.schema.json"
	validateErr := (*fybrikApp).ValidateFybrikApplication(taxonomyFile)
	assert.NotNil(t, validateErr, "Invalid interface error should be found")
}
