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

	filename := "../../../../samples/kubeflow/m4dapplication.yaml"
	applicationYaml, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	m4dApp := &M4DApplication{}
	err = yaml.Unmarshal(applicationYaml, m4dApp)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	taxonomyFile := "../../../../charts/m4d/files/taxonomy/application.values.schema.json"
	validateErr := m4dApp.validateM4DApplication(taxonomyFile)
	assert.Nil(t, validateErr, "No error should be found")
}

func TestInvalidAppInfo(t *testing.T) {
	t.Parallel()

	filename := "../../../testdata/unittests/m4dapplication-appInfoErrors.yaml"
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	m4dApp := &M4DApplication{}
	err = yaml.Unmarshal(buf, m4dApp)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	taxonomyFile := "../../../../charts/m4d/files/taxonomy/application.values.schema.json"
	validateErr := (*m4dApp).validateM4DApplication(taxonomyFile)
	assert.NotNil(t, validateErr, "Invalid appInfo error should be found")
}

func TestInvalidInterface(t *testing.T) {
	t.Parallel()

	filename := "../../../testdata/unittests/m4dapplication-interfaceErrors.yaml"
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	m4dApp := &M4DApplication{}
	err = yaml.Unmarshal(buf, m4dApp)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	taxonomyFile := "../../../../charts/m4d/files/taxonomy/application.values.schema.json"
	validateErr := (*m4dApp).validateM4DApplication(taxonomyFile)
	assert.NotNil(t, validateErr, "Invalid interface error should be found")
}
