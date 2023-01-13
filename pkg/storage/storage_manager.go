// Copyright 2023 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"errors"

	fappv1 "fybrik.io/fybrik/manager/apis/app/v1beta1"
	fappv2 "fybrik.io/fybrik/manager/apis/app/v1beta2"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/storage/registrator"
	"fybrik.io/fybrik/pkg/storage/registrator/agent"

	// Registration of the implementation agents is done by adding blank imports which invoke init() method of each package
	_ "fybrik.io/fybrik/pkg/storage/impl/mysql"
	_ "fybrik.io/fybrik/pkg/storage/impl/s3"
)

// internal implementation of StorageManager APIs
// OpenAPI - TBD

// AllocateStorage allocates storage based on the selected storage account by invoking the specific implementation agent
// returns a Connection object in case of success, and an error - otherwise
func AllocateStorage(account *fappv2.FybrikStorageAccountSpec, secret *fappv1.SecretRef, opts *agent.Options) (taxonomy.Connection, error) {
	if account == nil {
		return taxonomy.Connection{}, errors.New("invalid storage account")
	}
	impl, err := registrator.GetAgent(account.Type)
	if err != nil {
		return taxonomy.Connection{}, err
	}
	return impl.AllocateStorage(account, secret, opts)
}

// DeleteStorage deletes the existing storage by invoking the specific implementation agent based on the connection type
func DeleteStorage(connection *taxonomy.Connection, secret *fappv1.SecretRef, opts *agent.Options) error {
	if connection == nil {
		return errors.New("invalid connection object")
	}
	impl, err := registrator.GetAgent(connection.Name)
	if err != nil {
		return err
	}
	return impl.DeleteStorage(connection, secret, opts)
}

// retun a list of supported connection types
func GetSupportedConnectionTypes() []taxonomy.ConnectionType {
	return registrator.GetRegisteredTypes()
}
