// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

/*
	This package defines an interface for creating/deleting connections.
	Only S3 connections are currently supported.
	The following functionality is supported:
	- create a connection
	- delete a connection
*/

package storage

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	"fybrik.io/fybrik/pkg/environment"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/random"
	"fybrik.io/fybrik/pkg/utils"
)

const (
	bucketNameHashLength = 10
	endpointKey          = "%endpoint%"
	bucketKey            = "%bucket%"
	objectKey            = "%object_key%"
	defaultTaxonomy      = "{\"s3\": {\"name\": \"s3\", \"s3\": {\"bucket\": \"%bucket%\",\"endpoint\": \"%endpoint%\", \"object_key\": \"%object_key%\"}}}"
)

func generateBucketName(owner *types.NamespacedName) string {
	suffix, _ := random.Hex(bucketNameHashLength)
	name := owner.Name + "-" + owner.Namespace + suffix
	return utils.K8sConformName(name)
}

// ProvisionInterface is an interface for managing connections
type ProvisionInterface interface {
	CreateConnection(sa *fapp.FybrikStorageAccountSpec, datasetName string, owner *types.NamespacedName) (taxonomy.Connection, error)
	DeleteConnection(conn taxonomy.Connection, secret *fapp.SecretRef) error
}

// ProvisionImpl is an implementation of ProvisionInterface using Dataset CRDs
type ProvisionImpl struct {
	Client client.Client
}

// NewProvisionImpl returns a new ProvisionImpl object
func NewProvisionImpl(c client.Client) *ProvisionImpl {
	return &ProvisionImpl{
		Client: c,
	}
}

func (r *ProvisionImpl) NewClient(endpointArg string, secretKey types.NamespacedName) (*minio.Client, error) {
	prefix := "https://"
	useSSL := strings.HasPrefix(endpointArg, prefix)
	var endpoint string
	if !useSSL {
		prefix = "http://"
	}
	endpoint = strings.TrimPrefix(endpointArg, prefix)
	// Get credentials
	secret := v1.Secret{}
	if err := r.Client.Get(context.Background(), secretKey, &secret); err != nil {
		return nil, errors.Wrapf(err, "could not get a secret %s", secretKey.Name)
	}

	accessKey, secretAccessKey := string(secret.Data["access_key"]), string(secret.Data["secret_key"])
	if accessKey == "" || secretAccessKey == "" {
		return nil, errors.Errorf("could not retrieve credentials from the secret %s", secretKey.Name)
	}

	// Initialize minio client object.
	return minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretAccessKey, ""),
		Secure: useSSL,
	})
}

// CreateConnection creates a new connection
func (r *ProvisionImpl) CreateConnection(sa *fapp.FybrikStorageAccountSpec,
	genName string, owner *types.NamespacedName) (taxonomy.Connection, error) {
	endpoint := sa.Endpoint
	key := types.NamespacedName{Name: sa.SecretRef, Namespace: environment.GetSystemNamespace()}
	// Initialize minio client object.
	minioClient, err := r.NewClient(endpoint, key)
	if err != nil {
		return taxonomy.Connection{}, err
	}
	genBucketName := generateBucketName(owner)

	if err = minioClient.MakeBucket(context.Background(), genBucketName, minio.MakeBucketOptions{}); err != nil {
		return taxonomy.Connection{}, errors.Wrapf(err, "could not create a bucket %s", genBucketName)
	}
	return createS3Connection(sa.Endpoint, genBucketName, genName)
}

func (r *ProvisionImpl) DeleteConnection(conn taxonomy.Connection, secretRef *fapp.SecretRef) error {
	bucket, err := getProperty(conn, utils.S3, bucketKey)
	if err != nil {
		return err
	}
	endpoint, err := getProperty(conn, utils.S3, endpointKey)
	if err != nil {
		return err
	}
	key := types.NamespacedName{Name: secretRef.Name, Namespace: secretRef.Namespace}
	// Initialize minio client object.
	minioClient, err := r.NewClient(endpoint, key)
	if err != nil {
		return err
	}
	exists, err := minioClient.BucketExists(context.Background(), bucket)
	if !exists {
		return client.IgnoreNotFound(err)
	}
	for object := range minioClient.ListObjects(context.Background(), bucket, minio.ListObjectsOptions{Recursive: true}) {
		if err := minioClient.RemoveObject(context.Background(), bucket, object.Key, minio.RemoveObjectOptions{}); err != nil {
			return err
		}
	}
	return minioClient.RemoveBucket(context.Background(), bucket)
}

// ProvisionTest is an implementation of ProvisionInterface used for testing
type ProvisionTest struct {
	buckets []string
}

// NewProvisionTest constructs a new ProvisionTest object
func NewProvisionTest() *ProvisionTest {
	return &ProvisionTest{
		buckets: []string{},
	}
}

func (r *ProvisionTest) CreateConnection(sa *fapp.FybrikStorageAccountSpec, datasetName string,
	owner *types.NamespacedName) (taxonomy.Connection, error) {
	genBucketName := generateBucketName(owner)
	for _, b := range r.buckets {
		if b == genBucketName {
			return taxonomy.Connection{}, errors.New("Bucket already exists")
		}
	}
	r.buckets = append(r.buckets, genBucketName)
	return createS3Connection(sa.Endpoint, genBucketName, datasetName)
}

func (r *ProvisionTest) DeleteConnection(conn taxonomy.Connection, secretRef *fapp.SecretRef) error {
	buckets := []string{}
	bucket, err := getProperty(conn, utils.S3, bucketKey)
	if err != nil {
		return err
	}
	found := false
	for _, b := range r.buckets {
		if b == bucket {
			found = true
		} else {
			buckets = append(buckets, bucket)
		}
	}
	if found {
		r.buckets = buckets
		return nil
	}
	return errors.Errorf("could not find %s to delete", bucket)
}

func getProperty(conn taxonomy.Connection, connType taxonomy.ConnectionType, key string) (string, error) {
	templated, err := getTemplatedConnection(connType)
	if err != nil {
		return "", err
	}
	value := conn.AdditionalProperties.MatchPattern(templated.AdditionalProperties, key)
	if value == nil {
		return "", errors.Errorf("unknown property %s", key)
	}
	strValue, ok := value.(string)
	if !ok {
		return "", errors.Errorf("%s is not a string property", key)
	}
	return strValue, nil
}

func getTemplatedConnection(connType taxonomy.ConnectionType) (taxonomy.Connection, error) {
	taxonomyDef := os.Getenv("CONNECTION_TAXONOMY")
	if taxonomyDef == "" {
		taxonomyDef = defaultTaxonomy
	}
	connections := make(map[string]taxonomy.Connection, 0)
	if err := json.Unmarshal([]byte(taxonomyDef), &connections); err != nil {
		return taxonomy.Connection{}, errors.Wrapf(err, "could not parse taxonomy template definition %s", taxonomyDef)
	}
	connection, found := connections[string(connType)]
	if !found {
		return taxonomy.Connection{}, errors.New("Missing taxonomy definition for S3")
	}
	return connection, nil
}

func createS3Connection(endpoint, bucket, object string) (taxonomy.Connection, error) {
	connection, err := getTemplatedConnection(utils.S3)
	if err != nil {
		return connection, err
	}
	connection.AdditionalProperties.ReplaceTemplateWithValue(endpointKey, endpoint)
	connection.AdditionalProperties.ReplaceTemplateWithValue(bucketKey, bucket)
	connection.AdditionalProperties.ReplaceTemplateWithValue(objectKey, object)
	return connection, nil
}
