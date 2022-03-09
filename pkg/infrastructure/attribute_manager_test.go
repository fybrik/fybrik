// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package infrastructure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const Path string = "/tmp/taxonomy/infraattributes.json#/definitions/Infrastructure"

func TestValidInfrastructureAttributeSC(t *testing.T) {
	t.Parallel()

	data := `{
		"infrastructure":[{
			"attribute": "storage-cost",
			"description": "neverland object store",
			"value": "100",
			"type": "numeric",
			"units": "US Dollar",
			"instance": "account-neverland"
		}]
	}`
	validateErr := validateStructure([]byte(data), Path)
	assert.Nil(t, validateErr, "No error should be found")
}

func TestInvalidInfrastructureAttributeSC(t *testing.T) {
	t.Parallel()

	data := `{
		"infrastructure":[{
			"attribute": "storage-cost",
			"description": "neverland object store",
			"value": "100",
			"type": "numeric",
			"units": "USDollar",
			"instance": "account-neverland"
		}]
	}`
	validateErr := validateStructure([]byte(data), Path)
	assert.Error(t, validateErr, "An error is expected")
}

func TestValidInfrastructureAttributeBW(t *testing.T) {
	t.Parallel()

	data := `{
		"infrastructure":[{
			"attribute": "bandwidth",
			"description": "neverland object store",
			"value": "100",
			"type": "numeric",
			"units": "GB/s",
			"instance": "account-neverland"
		}]
	}`
	validateErr := validateStructure([]byte(data), Path)
	assert.Nil(t, validateErr, "No error should be found")
}

func TestInvalidInfrastructureAttributeBW(t *testing.T) {
	t.Parallel()

	data := `{
		"infrastructure":[{
			"attribute": "bandwidth",
			"description": "neverland object store",
			"value": "100",
			"type": "numeric",
			"units": "GB",
			"instance": "account-neverland"
		}]
	}`
	validateErr := validateStructure([]byte(data), Path)
	assert.Error(t, validateErr, "An error is expected")
}
