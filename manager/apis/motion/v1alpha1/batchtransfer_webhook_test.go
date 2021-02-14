// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestValidBatchTransfer(t *testing.T) {
	t.Parallel()
	batchTransfer := BatchTransfer{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: BatchTransferSpec{
			Source: DataStore{
				Database: &Database{
					Db2URL:    "jdbc:db2://host:1234/DB",
					Table:     "MY.TABLE",
					User:      "user",
					Password:  "password",
					VaultPath: nil,
				},
				S3:    nil,
				Kafka: nil,
			},
			Destination: DataStore{
				Database: nil,
				S3: &S3{
					Endpoint:   "my.endpoint",
					Region:     "eu-gb",
					Bucket:     "myBucket",
					AccessKey:  "ab",
					SecretKey:  "cd",
					ObjectKey:  "obj.parq",
					DataFormat: "parquet",
					VaultPath:  nil,
				},
				Kafka: nil,
			},
			Transformation:            nil,
			Schedule:                  "",
			Image:                     "",
			ImagePullPolicy:           "",
			SecretProviderURL:         "",
			SecretProviderRole:        "",
			Suspend:                   false,
			MaxFailedRetries:          0,
			SuccessfulJobHistoryLimit: 0,
			FailedJobHistoryLimit:     0,
		},
		Status: BatchTransferStatus{},
	}

	err := batchTransfer.validateBatchTransfer()
	assert.Nil(t, err, "No error should be found")
}

func TestValidBatchTransferKafka(t *testing.T) {
	t.Parallel()
	batchTransfer := BatchTransfer{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: BatchTransferSpec{
			Source: DataStore{
				Database: &Database{
					Db2URL:    "jdbc:db2://host:1234/DB",
					Table:     "MY.TABLE",
					User:      "user",
					Password:  "password",
					VaultPath: nil,
				},
				S3:    nil,
				Kafka: nil,
			},
			Destination: DataStore{
				Database: nil,
				Kafka: &Kafka{
					KafkaBrokers:      "localhost:9092",
					SchemaRegistryURL: "http://localhost:8080/v1",
					User:              "user",
					Password:          "pwd",
					KafkaTopic:        "topic",
					CreateSnapshot:    false,
					VaultPath:         nil,
				},
			},
			Transformation:            nil,
			Schedule:                  "",
			Image:                     "",
			ImagePullPolicy:           "",
			SecretProviderURL:         "",
			SecretProviderRole:        "",
			Suspend:                   false,
			MaxFailedRetries:          0,
			SuccessfulJobHistoryLimit: 0,
			FailedJobHistoryLimit:     0,
		},
		Status: BatchTransferStatus{},
	}

	err := batchTransfer.validateBatchTransfer()
	assert.Nil(t, err, "No error should be found")
}

func TestInValidKafkaConfiguration(t *testing.T) {
	t.Parallel()
	kafka := DataStore{
		Database: nil,
		Kafka: &Kafka{
			KafkaBrokers:      "localhost:99092",
			SchemaRegistryURL: "http://loca!lhost:8080/v1",
			User:              "user",
			Password:          "pwd",
			KafkaTopic:        "topic",
			CreateSnapshot:    false,
			VaultPath:         nil,
		},
	}

	path := field.NewPath("spec", "source")

	errors := validateDataStore(path, &kafka)

	assert.NotNil(t, errors)
	assert.Len(t, errors, 2)
	assert.Equal(t, "spec.source.kafka.kafkaBrokers", errors[0].Field)
	assert.Equal(t, "spec.source.kafka.schemaRegistryUrl", errors[1].Field)
}

func TestInvalidS3Bucket(t *testing.T) {
	t.Parallel()
	datastore := DataStore{
		Database: nil,
		S3: &S3{
			Endpoint:   "my.endpoint",
			Region:     "eu-gb",
			Bucket:     "",
			AccessKey:  "ab",
			SecretKey:  "cd",
			ObjectKey:  "obj.parq",
			DataFormat: "parquet",
			VaultPath:  nil,
		},
		Kafka: nil,
	}

	path := field.NewPath("spec", "source")

	// test missing bucket name
	err := validateDataStore(path, &datastore)

	assert.NotNil(t, err)
	assert.Len(t, err, 1)
	assert.Equal(t, "spec.source.s3.bucket", err[0].Field)

	// Test missing object key
	datastore.S3.Bucket = "mybucket"
	datastore.S3.ObjectKey = ""

	err = validateDataStore(path, &datastore)
	assert.NotNil(t, err)
	assert.Len(t, err, 1)
	assert.Equal(t, "spec.source.s3.objectKey", err[0].Field)

	// Test multiple errors
	datastore.S3.Endpoint = "test@wrong.com"
	datastore.S3.Bucket = ""
	datastore.S3.ObjectKey = ""
	err = validateDataStore(path, &datastore)
	assert.NotNil(t, err)
	assert.Len(t, err, 3)
}

func TestDefaultingS3Bucket(t *testing.T) {
	t.Parallel()
	_ = os.Setenv("IMAGE_PULL_POLICY", "Always")
	_ = os.Setenv("MOVER_IMAGE", "mover-test:latest")
	_ = os.Setenv("SECRET_PROVIDER_URL", "mysecrets:123")
	_ = os.Setenv("SECRET_PROVIDER_ROLE", "demo")

	batchTransfer := BatchTransfer{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: BatchTransferSpec{
			Source: DataStore{
				Database: &Database{
					Db2URL:    "jdbc:db2://host:1234/DB",
					Table:     "MY.TABLE",
					User:      "user",
					Password:  "password",
					VaultPath: nil,
				},
				S3:    nil,
				Kafka: nil,
			},
			Destination: DataStore{
				Database: nil,
				S3: &S3{
					Endpoint:   "my.endpoint",
					Region:     "eu-gb",
					Bucket:     "myBucket",
					AccessKey:  "ab",
					SecretKey:  "cd",
					ObjectKey:  "obj.parq",
					DataFormat: "parquet",
					VaultPath:  nil,
				},
				Kafka: nil,
			},
			Transformation:            nil,
			Schedule:                  "",
			Image:                     "",
			ImagePullPolicy:           "",
			SecretProviderURL:         "",
			SecretProviderRole:        "",
			Suspend:                   false,
			MaxFailedRetries:          0,
			SuccessfulJobHistoryLimit: 0,
			FailedJobHistoryLimit:     0,
		},
		Status: BatchTransferStatus{},
	}

	batchTransfer.Default()

	assert.Equal(t, corev1.PullAlways, batchTransfer.Spec.ImagePullPolicy)
	assert.Equal(t, "mover-test:latest", batchTransfer.Spec.Image)
	assert.Equal(t, "mysecrets:123", batchTransfer.Spec.SecretProviderURL)
	assert.Equal(t, "demo", batchTransfer.Spec.SecretProviderRole)
	assert.Equal(t, false, batchTransfer.Spec.Suspend)
	assert.Equal(t, DefaultFailedJobHistoryLimit, batchTransfer.Spec.FailedJobHistoryLimit)
	assert.Equal(t, DefaultSuccessfulJobHistoryLimit, batchTransfer.Spec.SuccessfulJobHistoryLimit)
	assert.Equal(t, "jdbc:db2://host:1234/DB/MY.TABLE", batchTransfer.Spec.Source.Description)
	assert.Equal(t, "s3://myBucket/obj.parq", batchTransfer.Spec.Destination.Description)
	assert.Equal(t, Batch, batchTransfer.Spec.DataFlowType)
	assert.Equal(t, LogData, batchTransfer.Spec.ReadDataType)
	assert.Equal(t, LogData, batchTransfer.Spec.WriteDataType)
	assert.Equal(t, Overwrite, batchTransfer.Spec.WriteOperation)
	_ = os.Unsetenv("IMAGE_PULL_POLICY")
	_ = os.Unsetenv("MOVER_IMAGE")
	_ = os.Unsetenv("SECRET_PROVIDER_URL")
	_ = os.Unsetenv("SECRET_PROVIDER_ROLE")
}
