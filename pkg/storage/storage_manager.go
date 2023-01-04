package storage

import (
	"errors"

	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	sa "fybrik.io/fybrik/pkg/storage/apis/app/v1beta2"
	"fybrik.io/fybrik/pkg/storage/registrator"
	"fybrik.io/fybrik/pkg/storage/registrator/agent"

	// Registration of the implementation agents is done by adding blank imports which invoke init() method of each package
	_ "fybrik.io/fybrik/pkg/storage/mysql_agent"
	_ "fybrik.io/fybrik/pkg/storage/s3_agent"
)

// internal implementation of StorageManager APIs
// OpenAPI - TBD

// AllocateStorage allocates storage based on the selected storage account by invoking the specific implementation agent
// returns a Connection object in case of success, and an error - otherwise
func AllocateStorage(account *sa.FybrikStorageAccountSpec, secret *fapp.SecretRef, opts *agent.Options) (taxonomy.Connection, error) {
	if account == nil {
		return taxonomy.Connection{}, errors.New("invalid storage account")
	}
	agent := registrator.GetAgent(account.Type)
	if agent == nil {
		return taxonomy.Connection{}, errors.New("unsupported connection type")
	}
	return agent.AllocateStorage(account, secret, opts)
}

// DeleteStorage deletes the existing storage by invoking the specific implementation agent based on the connection type
func DeleteStorage(connection *taxonomy.Connection, secret *fapp.SecretRef, opts *agent.Options) error {
	if connection == nil {
		return errors.New("invalid connection object")
	}
	agent := registrator.GetAgent(connection.Name)
	if agent == nil {
		return errors.New("unsupported connection type")
	}
	return agent.DeleteStorage(connection, secret, opts)
}

// retun a list of supported connection types
func GetSupportedConnectionTypes() []taxonomy.ConnectionType {
	return registrator.GetRegisteredTypes()
}
