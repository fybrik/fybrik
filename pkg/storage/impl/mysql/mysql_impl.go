// Copyright 2023 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"emperror.dev/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"

	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/storagemanager"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/random"
	"fybrik.io/fybrik/pkg/serde"
	"fybrik.io/fybrik/pkg/storage/registrator"
	"fybrik.io/fybrik/pkg/storage/registrator/agent"
)

const (
	hostKey            = "host"
	portKey            = "port"
	dbKey              = "database"
	tableKey           = "table"
	usernameKey        = "username"
	passwordKey        = "password"
	timeout            = time.Second * 5
	mysqlAgent         = "mysql"
	randomSuffixLength = 5
	endStatement       = ";"
)

// Storage manager implementation for MySQL
type MySQLImpl struct {
	Name taxonomy.ConnectionType
	Log  zerolog.Logger
}

// implementation of AgentInterface for MySQL
func NewMySQLImpl() *MySQLImpl {
	return &MySQLImpl{Name: mysqlAgent, Log: logging.LogInit(logging.CONNECTOR, "StorageManager")}
}

// register the implementation for MySQL
func init() {
	mysqlImpl := NewMySQLImpl()
	if err := registrator.Register(mysqlImpl); err != nil {
		mysqlImpl.Log.Error().Err(err)
	}
}

// return the supported connection type
func (impl *MySQLImpl) GetConnectionType() taxonomy.ConnectionType {
	return impl.Name
}

func dsn(username, password, hostname, port, dbName string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, hostname, port, dbName)
}

// returns a connection to MySQL
func NewClient(host, port, dbName string, secretRef taxonomy.SecretRef, client kclient.Client) (*sql.DB, error) {
	// Get credentials
	secret := v1.Secret{}
	if err := client.Get(context.Background(), types.NamespacedName{Name: secretRef.Name,
		Namespace: secretRef.Namespace}, &secret); err != nil {
		return nil, errors.Wrapf(err, "could not get a secret %s", secretRef.Name)
	}

	username, password := string(secret.Data[usernameKey]), string(secret.Data[passwordKey])
	if username == "" || password == "" {
		return nil, errors.Errorf("could not retrieve credentials from the secret %s", secretRef.Name)
	}

	// connect to the server
	return sql.Open(mysqlAgent, dsn(username, password, host, port, dbName))
}

// storage allocation
// host, port are taken from the storage account
// database name is generated, table is taken from the DatasetProperties.Name
// TODO: allow database and table specification inside FybrikApplication
func (impl *MySQLImpl) AllocateStorage(request *storagemanager.AllocateStorageRequest, client kclient.Client) (taxonomy.Connection, error) {
	var host, port string
	var err error
	if host, err = agent.GetProperty(request.AccountProperties.Items, impl.Name, hostKey); err != nil {
		return taxonomy.Connection{}, err
	}
	if port, err = agent.GetProperty(request.AccountProperties.Items, impl.Name, portKey); err != nil {
		return taxonomy.Connection{}, err
	}
	portVal, err := strconv.Atoi(port)
	if err != nil {
		return taxonomy.Connection{}, err
	}

	// generate database name
	suffix, _ := random.Hex(randomSuffixLength)
	database := request.Opts.AppDetails.Name + "-" + request.Opts.AppDetails.Namespace + suffix

	table := request.Opts.DatasetProperties.Name
	// connect to the server and create the database if empty
	db, err := NewClient(host, port, "", request.Secret, client)
	if err != nil {
		return taxonomy.Connection{}, err
	}
	defer db.Close()
	ctx, cancelfunc := context.WithTimeout(context.Background(), timeout)
	defer cancelfunc()
	query := "CREATE DATABASE IF NOT EXISTS " + database + endStatement
	impl.Log.Info().Msgf("Sending query %s\n", query)
	if _, err = db.ExecContext(ctx, query); err != nil {
		return taxonomy.Connection{}, err
	}
	connection := taxonomy.Connection{
		Name: impl.Name,
		AdditionalProperties: serde.Properties{
			Items: map[string]interface{}{
				string(impl.Name): map[string]interface{}{
					hostKey:  host,
					portKey:  portVal,
					dbKey:    database,
					tableKey: table,
				},
			},
		},
	}
	return connection, nil
}

// storage deletion
func (impl *MySQLImpl) DeleteStorage(request *storagemanager.DeleteStorageRequest, client kclient.Client) error {
	var host, port, database string
	var err error
	if host, err = agent.GetProperty(request.Connection.AdditionalProperties.Items, impl.Name, hostKey); err != nil {
		return err
	}
	if port, err = agent.GetProperty(request.Connection.AdditionalProperties.Items, impl.Name, portKey); err != nil {
		return err
	}
	if database, err = agent.GetProperty(request.Connection.AdditionalProperties.Items, impl.Name, dbKey); err != nil {
		return err
	}
	// connect to the server
	db, err := NewClient(host, port, database, request.Secret, client)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancelfunc := context.WithTimeout(context.Background(), timeout)
	defer cancelfunc()
	if _, err = db.ExecContext(ctx, "DROP DATABASE IF EXISTS "+database+endStatement); err != nil {
		return err
	}
	return nil
}
