// Copyright 2023 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"errors"
	"fmt"

	kclient "sigs.k8s.io/controller-runtime/pkg/client"

	"fybrik.io/fybrik/pkg/model/storagemanager"
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

// agent interface for managing storage for a supported connection type
type AgentInterface interface {
	// allocate storage
	AllocateStorage(request *storagemanager.AllocateStorageRequest, client kclient.Client) (taxonomy.Connection, error)
	// delete storage
	DeleteStorage(request *storagemanager.DeleteStorageRequest, client kclient.Client) error
	// return the supported connection type
	GetConnectionType() taxonomy.ConnectionType
}

// get property
func GetProperty(props map[string]interface{}, t taxonomy.ConnectionType, key string) (string, error) {
	propertyMap := props[string(t)]
	if propertyMap == nil {
		return "", errors.New("undefined properties for " + string(t))
	}
	switch propertyMap := propertyMap.(type) {
	case map[string]interface{}:
		property := propertyMap[key]
		if property != nil {
			return fmt.Sprintf("%v", property), nil
		}
	default:
		break
	}
	return "", errors.New("undefined or missing property " + key)
}
