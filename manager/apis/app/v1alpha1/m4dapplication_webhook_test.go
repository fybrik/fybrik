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
	errors := m4dApp.validateM4DApplication(taxonomyFile)
	assert.Nil(t, errors, "No error should be found")
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
	errors := (*m4dApp).validateM4DApplication(taxonomyFile)
	assert.NotNil(t, errors)
	assert.Len(t, errors, 2)
	// assert.Equal(t, "spec.appInfo", errors[0].Field.errorList)
	// assert.Equal(t, "spec.appInfo.role", errors[1].Field)
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
	errors := m4dApp.validateM4DApplication(taxonomyFile)
	assert.Len(t, errors, 2)
	// assert.Equal(t, "spec.data.0.requirements.interface", errors[0].Field)
	// assert.Equal(t, "spec.data.0.requirements.interface.dataformat", errors[1].Field)
}
