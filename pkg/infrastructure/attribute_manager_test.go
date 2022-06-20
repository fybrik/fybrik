// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package infrastructure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidInfrastructureAttributeSC(t *testing.T) {
	t.Parallel()

	data := `{
		"metrics": [{
			"name": "cost",
			"type": "numeric",
			"units": "US Dollar per TB per month",
			"scale": {"min": 0, "max": 100}
		}],
		"infrastructure":[{
			"attribute": "storage-cost",
			"description": "neverland object store",
			"metricsName": "cost",
			"value": "100",
			"object": "fybrikstorageaccount",
			"instance": "account-neverland"
		}]
	}`
	validateErr := validateStructure([]byte(data))
	assert.Nil(t, validateErr, "No error should be found")
}

func TestInvalidInfrastructureAttributeSC(t *testing.T) {
	t.Parallel()

	data := `{
		"metrics": [{
			"name": "cost",
			"type": "numeric",
			"units": "km"
		}],
		"infrastructure":[{
			"attribute": "storage-cost",
			"metricsName": "cost",
			"description": "neverland object store",
			"value": "100",
			"instance": "account-neverland"
		}]
	}`
	validateErr := validateStructure([]byte(data))
	assert.Error(t, validateErr, "An error is expected")
}

func TestValidInfrastructureAttributeDist(t *testing.T) {
	t.Parallel()

	data := `{
		"metrics": [{
			"name": "distance",
			"type": "numeric",
			"units": "km"
		}],
		"infrastructure":[{
			"attribute": "distance",
			"metricsName": "distance",
			"value": "1000",
			"arguments": ["neverland","theshire"]
		}]
	}`
	validateErr := validateStructure([]byte(data))
	assert.Nil(t, validateErr, "No error should be found")
}

func TestInvalidInfrastructureAttributeDist(t *testing.T) {
	t.Parallel()

	data := `{
		"metrics": [{
			"name": "distance",
			"type": "object",
			"units": "m"
		}],
		"infrastructure":[{
			"attribute": "distance",
			"metricsName": "distance",
			"value": "100",
			"objects": ["fybrikstorageaccount","fybrikstorageaccount"]
		}]
	}`
	validateErr := validateStructure([]byte(data))
	assert.Error(t, validateErr, "An error is expected")
}
