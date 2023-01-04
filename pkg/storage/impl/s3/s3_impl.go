// Copyright 2023 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package s3

import (
	"github.com/rs/zerolog"

	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	sa "fybrik.io/fybrik/pkg/storage/apis/app/v1beta2"
	registrator "fybrik.io/fybrik/pkg/storage/registrator"
	agent "fybrik.io/fybrik/pkg/storage/registrator/agent"
)

// s3 storage manager implementaton
type S3Impl struct {
	Name taxonomy.ConnectionType
	Log  zerolog.Logger
}

func NewS3Impl() *S3Impl {
	return &S3Impl{Name: "s3", Log: logging.LogInit(logging.CONNECTOR, "StorageManager")}
}

// register the implementation for s3
func init() {
	s3Impl := NewS3Impl()
	if err := registrator.Register(s3Impl); err != nil {
		s3Impl.Log.Error().Err(err)
	}
}

// return the supported connection type
func (impl *S3Impl) GetConnectionType() taxonomy.ConnectionType {
	return impl.Name
}

// allocate storage for s3 - placeholder
func (impl *S3Impl) AllocateStorage(account *sa.FybrikStorageAccountSpec, secret *fapp.SecretRef,
	opts *agent.Options) (taxonomy.Connection, error) {
	return taxonomy.Connection{Name: impl.Name}, nil
}

// delete s3 storage - placeholder
func (impl *S3Impl) DeleteStorage(connection *taxonomy.Connection, secret *fapp.SecretRef, opts *agent.Options) error {
	return nil
}
