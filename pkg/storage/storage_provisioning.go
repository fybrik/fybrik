// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

/*
	This package defines an interface for managing dynamically allocated S3 buckets.
	The current implementation manages buckets using Dataset resources.
	Convention: Dataset resources have the same name as the name of the provisioned bucket.
	The following functionality is supported:
	- allocating a bucket
	- checking allocation status
	- deleting a temporary bucket
	- marking a bucket as persistent (will not be removed upon Dataset deletion)
*/

package storage

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	OwnerLabel          = "fybrik.io/owner"
	RemoveOnDeleteLabel = "remove-on-delete"
	RequestTrue         = "true"
	RequestFalse        = "false"
	SpecKey             = "spec"
	LocalKey            = "local"
	StatusKey           = "status"
	BucketKey           = "bucket"
	SecretNameKey       = "secret-name"
	SecretNamespaceKey  = "secret-namespace"
	EndpointKey         = "endpoint"
	ProvisionKey        = "provision"
	ProvisionInfoKey    = "info"
	ProvisionStatusKey  = "status"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "com.ie.ibm.hpsys", Version: "v1alpha1"}
)

// ProvisionedBucket holds information about the bucket to be provisioned.
// In the future releases this structure may be extented to include other data store types.
type ProvisionedBucket struct {
	// Bucket name
	Name string
	// Endpoint
	Endpoint string
	// Storage Region
	Region string
	// Secret containing credentials
	SecretRef types.NamespacedName
}

// ProvisionedStorageStatus includes the status of the provisioning and an error message if the provisioning has failed
type ProvisionedStorageStatus struct {
	Provisioned bool
	ErrorMsg    string
}

// ProvisionInterface is an interface for managing dynamically allocated Dataset resources
type ProvisionInterface interface {
	CreateDataset(ref *types.NamespacedName, dataset *ProvisionedBucket, owner *types.NamespacedName) error
	DeleteDataset(ref *types.NamespacedName) error
	GetDatasetStatus(ref *types.NamespacedName) (*ProvisionedStorageStatus, error)
	SetPersistent(ref *types.NamespacedName, persistent bool) error
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

func newDatasetAsUnstructured(name, namespace string) *unstructured.Unstructured {
	object := &unstructured.Unstructured{}
	object.SetGroupVersionKind(schema.GroupVersionKind{Group: GroupVersion.Group, Version: GroupVersion.Version, Kind: "Dataset"})
	object.SetNamespace(namespace)
	object.SetName(name)
	return object
}

func (r *ProvisionImpl) getDatasetAsUnstructured(name, namespace string) (*unstructured.Unstructured, error) {
	object := newDatasetAsUnstructured(name, namespace)
	objectKey := client.ObjectKeyFromObject(object)

	if err := r.Client.Get(context.Background(), objectKey, object); err != nil {
		return nil, err
	}
	return object, nil
}

func getValue(obj map[string]interface{}, path ...string) string {
	if valStr, exists, err := unstructured.NestedString(obj, path...); err == nil && exists {
		return valStr
	}
	return ""
}

func equal(required *ProvisionedBucket, existing *unstructured.Unstructured) bool {
	obj := existing.UnstructuredContent()
	if required.Name != getValue(obj, SpecKey, LocalKey, BucketKey) {
		return false
	}
	if required.Endpoint != getValue(obj, SpecKey, LocalKey, EndpointKey) {
		return false
	}
	if required.SecretRef.Name != getValue(obj, SpecKey, LocalKey, SecretNameKey) {
		return false
	}
	if required.SecretRef.Namespace != getValue(obj, SpecKey, LocalKey, SecretNamespaceKey) {
		return false
	}
	return true
}

// CreateDataset generates a Dataset resource
func (r *ProvisionImpl) CreateDataset(ref *types.NamespacedName, bucket *ProvisionedBucket, owner *types.NamespacedName) error {
	existing, err := r.getDatasetAsUnstructured(ref.Name, ref.Namespace)
	if err == nil {
		if equal(bucket, existing) {
			// update is not required
			return nil
		}
		// re-create the dataset
		if err = r.DeleteDataset(ref); err != nil { //nolint:gocritic // Two lints conflicting on err assginment
			return err
		}
	}
	values := map[string]string{
		"type":             "COS",
		SecretNameKey:      bucket.SecretRef.Name,
		SecretNamespaceKey: bucket.SecretRef.Namespace,
		EndpointKey:        bucket.Endpoint,
		BucketKey:          bucket.Name,
		ProvisionKey:       RequestTrue}

	dataset := newDatasetAsUnstructured(ref.Name, ref.Namespace)
	dataset.SetLabels(map[string]string{
		OwnerLabel:          owner.Namespace + "." + owner.Name,
		RemoveOnDeleteLabel: removeOnDeleteValue(false)})

	if err := unstructured.SetNestedStringMap(dataset.Object, values, SpecKey, LocalKey); err != nil {
		return err
	}
	return r.Client.Create(context.Background(), dataset)
}

// SetPersistent updates a "remove-on-delete" label of the existing Dataset resource
func (r *ProvisionImpl) SetPersistent(ref *types.NamespacedName, persistent bool) error {
	existing, err := r.getDatasetAsUnstructured(ref.Name, ref.Namespace)
	if err != nil {
		return err
	}
	labels := existing.GetLabels()

	if labels == nil {
		labels = make(map[string]string)
	}
	labels[RemoveOnDeleteLabel] = removeOnDeleteValue(persistent)
	existing.SetLabels(labels)
	return r.Client.Update(context.Background(), existing)
}

// GetDatasetStatus returns status of an existing Dataset resource.
func (r *ProvisionImpl) GetDatasetStatus(ref *types.NamespacedName) (*ProvisionedStorageStatus, error) {
	dataset, err := r.getDatasetAsUnstructured(ref.Name, ref.Namespace)
	if err != nil {
		return nil, err
	}
	status := getValue(dataset.Object, StatusKey, ProvisionKey, ProvisionStatusKey)
	info := getValue(dataset.Object, StatusKey, ProvisionKey, ProvisionInfoKey)
	return &ProvisionedStorageStatus{Provisioned: status == "OK", ErrorMsg: info}, nil
}

// DeleteDataset deletes the existing Dataset resource
func (r *ProvisionImpl) DeleteDataset(ref *types.NamespacedName) error {
	dataset, err := r.getDatasetAsUnstructured(ref.Name, ref.Namespace)
	if err == nil {
		return r.Client.Delete(context.Background(), dataset)
	}
	return err
}

// ProvisionTest is an implementation of ProvisionInterface used for testing
type ProvisionTest struct {
	datasets []*ProvisionedBucket
}

// NewProvisionTest constructs a new ProvisionTest object
func NewProvisionTest() *ProvisionTest {
	return &ProvisionTest{
		datasets: []*ProvisionedBucket{},
	}
}

// CreateDataset generates a new dataset
func (r *ProvisionTest) CreateDataset(ref *types.NamespacedName, dataset *ProvisionedBucket, owner *types.NamespacedName) error {
	for i, d := range r.datasets {
		if d.Name == dataset.Name {
			r.datasets[i] = dataset
			return nil
		}
	}
	r.datasets = append(r.datasets, dataset)
	return nil
}

// SetPersistent does nothing for the testing implementation except for verifying that the dataset exists
func (r *ProvisionTest) SetPersistent(ref *types.NamespacedName, persistent bool) error {
	for _, d := range r.datasets {
		if d.Name == ref.Name {
			return nil
		}
	}
	return fmt.Errorf("could not find a dataset: %s", ref.Name)
}

// GetDatasetStatus returns status of an existing Dataset resource.
func (r *ProvisionTest) GetDatasetStatus(ref *types.NamespacedName) (*ProvisionedStorageStatus, error) {
	for _, d := range r.datasets {
		if d.Name == ref.Name {
			return &ProvisionedStorageStatus{Provisioned: true}, nil
		}
	}
	return nil, fmt.Errorf("could not get status of a dataset: %s", ref.Name)
}

// DeleteDataset removes an existing dataset
func (r *ProvisionTest) DeleteDataset(ref *types.NamespacedName) error {
	newDatasets := []*ProvisionedBucket{}
	found := false
	errMessage := "The following datasets have been found:\n"
	for _, d := range r.datasets {
		errMessage += " " + d.Name + "\n"
		if d.Name == ref.Name {
			found = true
		} else {
			newDatasets = append(newDatasets, d)
		}
	}
	if found {
		r.datasets = newDatasets
		return nil
	}
	return fmt.Errorf("could not delete a dataset %s\n%s", ref.Name, errMessage)
}

// label value based on persistency
func removeOnDeleteValue(persistent bool) string {
	if persistent {
		return RequestFalse
	}
	return RequestTrue
}
