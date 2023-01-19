// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	fappv1 "fybrik.io/fybrik/manager/apis/app/v1beta1"
	fappv2 "fybrik.io/fybrik/manager/apis/app/v1beta2"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/storage/registrator/agent"
)

// StorageManagerInterface is an interface for storage management
type StorageManagerInterface interface {
	// AllocateStorage allocates storage based on the selected storage account by invoking the specific implementation agent
	// returns a Connection object in case of success, and an error - otherwise
	AllocateStorage(account *fappv2.FybrikStorageAccountSpec, secret *fappv1.SecretRef, opts *agent.Options) (taxonomy.Connection, error)
	// DeleteStorage deletes the allocated storage
	DeleteStorage(connection *taxonomy.Connection, secret *fappv1.SecretRef, opts *agent.Options) error
	// GetSupportedConnectionTypes returns a list of supported connection types
	GetSupportedConnectionTypes() []taxonomy.ConnectionType
}

type StorageManager struct{}

func NewStorageManager() *StorageManager {
	if err := InitK8sClient(); err != nil {
		return nil
	}
	return &StorageManager{}
}

func (r *StorageManager) AllocateStorage(account *fappv2.FybrikStorageAccountSpec, secret *fappv1.SecretRef,
	opts *agent.Options) (taxonomy.Connection, error) {
	return AllocateStorage(account, secret, opts)
}

func (r *StorageManager) DeleteStorage(connection *taxonomy.Connection, secret *fappv1.SecretRef, opts *agent.Options) error {
	return DeleteStorage(connection, secret, opts)
}

func (r *StorageManager) GetSupportedConnectionTypes() []taxonomy.ConnectionType {
	return GetSupportedConnectionTypes()
}

// For testing: support s3 and db2
type FakeStorageManager struct{}

func NewFakeStorageManager() *FakeStorageManager {
	return &FakeStorageManager{}
}

func (r *FakeStorageManager) AllocateStorage(account *fappv2.FybrikStorageAccountSpec, secret *fappv1.SecretRef,
	opts *agent.Options) (taxonomy.Connection, error) {
	return taxonomy.Connection{Name: account.Type, AdditionalProperties: account.AdditionalProperties}, nil
}

func (r *FakeStorageManager) DeleteStorage(connection *taxonomy.Connection, secret *fappv1.SecretRef, opts *agent.Options) error {
	return nil
}

func (r *FakeStorageManager) GetSupportedConnectionTypes() []taxonomy.ConnectionType {
	return []taxonomy.ConnectionType{"s3", "db2", "mysql"}
}
