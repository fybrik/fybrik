package s3_agent

import (
	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	sa "fybrik.io/fybrik/pkg/storage/apis/app/v1beta2"
	"fybrik.io/fybrik/pkg/storage/registrator"
	"fybrik.io/fybrik/pkg/storage/registrator/agent"
)

// s3 storage manager implementaton
type S3Impl struct {
	Name taxonomy.ConnectionType
}

func NewS3Impl() *S3Impl {
	return &S3Impl{Name: "s3"}
}

// register the implementation for s3
func init() {
	registrator.Register(NewS3Impl())
}

// return the supported connection type
func (impl *S3Impl) GetConnectionType() taxonomy.ConnectionType {
	return impl.Name
}

// allocate storage for s3 - placeholder
func (impl *S3Impl) AllocateStorage(account *sa.FybrikStorageAccountSpec, secret *fapp.SecretRef, opts *agent.Options) (taxonomy.Connection, error) {
	return taxonomy.Connection{Name: impl.Name}, nil
}

// delete s3 storage - placeholder
func (impl *S3Impl) DeleteStorage(connection *taxonomy.Connection, secret *fapp.SecretRef, opts *agent.Options) error {
	return nil
}
