// Copyright 2023 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"k8s.io/apimachinery/pkg/types"

	fappv1 "fybrik.io/fybrik/manager/apis/app/v1beta1"
	fappv2 "fybrik.io/fybrik/manager/apis/app/v1beta2"
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

// Details of the owner application
type ApplicationDetails struct {
	// Name and namespace
	Owner *types.NamespacedName
	// uuid
	UUID string
}

// Details of the new asset
// The current implementation includes only a name provided in the write flow for a new asset
type DatasetDetails struct {
	Name string
}

// Configuration options
// TODO: extend IT config policies to return options for storage management
type ConfigOptions struct {
	// Delete an empty folder/bucket when the allocated storage is deleted
	DeleteEmptyFolder bool
}

// Additional options provided for storage allocation/deletion
type Options struct {
	AppDetails        ApplicationDetails
	DatasetProperties DatasetDetails
	ConfigurationOpts ConfigOptions
}

// agent interface for managing storage for a supported connection type
type AgentInterface interface {
	// allocate storage
	AllocateStorage(account *fappv2.FybrikStorageAccountSpec, secret *fappv1.SecretRef, opts *Options) (taxonomy.Connection, error)
	// delete storage
	DeleteStorage(connection *taxonomy.Connection, secret *fappv1.SecretRef, opts *Options) error
	// return the supported connection type
	GetConnectionType() taxonomy.ConnectionType
}
