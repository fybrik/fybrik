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

	"emperror.dev/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	comv1alpha1 "github.com/IBM/dataset-lifecycle-framework/src/dataset-operator/pkg/apis/com/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// ProvisionedBucket holds information about the bucket to be provisioned.
// In the future releases this structure may be extented to include other data store types.
type ProvisionedBucket struct {
	// Bucket name
	Name string
	// Endpoint
	Endpoint string
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

func equal(required *ProvisionedBucket, existing *comv1alpha1.DatasetSpec) bool {
	if required.Name != existing.Local["bucket"] {
		return false
	}
	if required.Endpoint != existing.Local["endpoint"] {
		return false
	}
	if required.SecretRef.Name != existing.Local["secret-name"] || required.SecretRef.Namespace != existing.Local["secret-namespace"] {
		return false
	}
	return true
}

// CreateDataset generates a Dataset resource
func (r *ProvisionImpl) CreateDataset(ref *types.NamespacedName, bucket *ProvisionedBucket, owner *types.NamespacedName) error {
	existing := &comv1alpha1.Dataset{}
	if err := r.Client.Get(context.Background(), *ref, existing); err == nil {
		if equal(bucket, &existing.Spec) {
			// update is not required
			return nil
		}
		// re-create the dataset
		if err = r.DeleteDataset(ref); err != nil {
			return err
		}
	}
	values := map[string]string{
		"type":             "COS",
		"secret-name":      bucket.SecretRef.Name,
		"secret-namespace": bucket.SecretRef.Namespace,
		"endpoint":         bucket.Endpoint,
		"bucket":           bucket.Name,
		"provision":        "true"}
	dataset := &comv1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ref.Name,
			Namespace: ref.Namespace,
			Labels: map[string]string{
				"m4d.ibm.com/owner": owner.Namespace + "." + owner.Name,
				"remove-on-delete":  "true"},
		},
		Spec: comv1alpha1.DatasetSpec{Local: values},
	}

	return r.Client.Create(context.Background(), dataset)
}

// SetPersistent updates a "remove-on-delete" label of the existing Dataset resource
func (r *ProvisionImpl) SetPersistent(ref *types.NamespacedName, persistent bool) error {
	existing := &comv1alpha1.Dataset{}
	if err := r.Client.Get(context.Background(), *ref, existing); err != nil {
		return err
	}
	if existing.Labels == nil {
		existing.Labels = make(map[string]string)
	}
	var removeOnDelete string
	if persistent {
		removeOnDelete = "false"
	} else {
		removeOnDelete = "true"
	}
	existing.Labels["remove-on-delete"] = removeOnDelete
	return r.Client.Update(context.Background(), existing)
}

// GetDatasetStatus returns status of an existing Dataset resource.
func (r *ProvisionImpl) GetDatasetStatus(ref *types.NamespacedName) (*ProvisionedStorageStatus, error) {
	dataset := &comv1alpha1.Dataset{}
	if err := r.Client.Get(context.Background(), types.NamespacedName{Name: ref.Name, Namespace: ref.Namespace}, dataset); err != nil {
		return nil, err
	}
	return &ProvisionedStorageStatus{Provisioned: dataset.Status.Provision.Status == "OK", ErrorMsg: dataset.Status.Provision.Info}, nil
}

// DeleteDataset deletes the existing Dataset resource
func (r *ProvisionImpl) DeleteDataset(ref *types.NamespacedName) error {
	dataset := &comv1alpha1.Dataset{}
	if err := r.Client.Get(context.Background(), types.NamespacedName{Name: ref.Name, Namespace: ref.Namespace}, dataset); err != nil {
		return err
	}
	return r.Client.Delete(context.Background(), dataset)
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
	return errors.New("Could not find a dataset: " + ref.Name)
}

// GetDatasetStatus returns status of an existing Dataset resource.
func (r *ProvisionTest) GetDatasetStatus(ref *types.NamespacedName) (*ProvisionedStorageStatus, error) {
	for _, d := range r.datasets {
		if d.Name == ref.Name {
			return &ProvisionedStorageStatus{Provisioned: true}, nil
		}
	}
	return nil, errors.New("Could not find a dataset: " + ref.Name)
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
	return errors.New("Could not delete a dataset " + ref.Name + "\n" + errMessage)
}
