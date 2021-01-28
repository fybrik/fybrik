// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	comv1alpha1 "github.com/IBM/dataset-lifecycle-framework/src/dataset-operator/pkg/apis/com/v1alpha1"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=m4dstorageaccounts,verbs=get;list;watch;update;
// +kubebuilder:rbac:groups=com.ie.ibm.hpsys,resources=datasets,verbs=get;list;watch;create;update;patch;delete

// ProvisionInterface is an interface for managing dynamically allocated Dataset resources
type ProvisionInterface interface {
	CreateDataset(dataset *comv1alpha1.Dataset) error
	DeleteDataset(ref *app.ResourceReference, force bool) error
	GetDataset(ref *app.ResourceReference) (*comv1alpha1.Dataset, error)
}

type ProvisionImpl struct {
	Client client.Client
}

func NewProvisionImpl(c client.Client) *ProvisionImpl {
	return &ProvisionImpl{
		Client: c,
	}
}

func (r *ProvisionImpl) CreateDataset(dataset *comv1alpha1.Dataset) error {
	ref := &app.ResourceReference{Name: dataset.Name, Namespace: dataset.Namespace}
	existing, err := r.GetDataset(ref)
	if err == nil {
		if equality.Semantic.DeepEqual(&existing.Spec, &dataset.Spec) {
			// nothing is needed to be done
			return nil
		}
		// re-create the dataset
		if err = r.DeleteDataset(ref, true); err != nil {
			return err
		}
	}
	return r.Client.Create(context.Background(), dataset)
}

func (r *ProvisionImpl) GetDataset(ref *app.ResourceReference) (*comv1alpha1.Dataset, error) {
	existing := &comv1alpha1.Dataset{}
	if err := r.Client.Get(context.Background(), types.NamespacedName{Name: ref.Name, Namespace: ref.Namespace}, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (r *ProvisionImpl) DeleteDataset(ref *app.ResourceReference, force bool) error {
	existing, err := r.GetDataset(ref)
	if err != nil {
		return err
	}
	// TODO(shlomitk1): update datasets with the appropriate annotation to delete a bucket upon removal
	return r.Client.Delete(context.Background(), existing)
}

type ProvisionTest struct {
	datasets []*comv1alpha1.Dataset
}

func NewProvisionTest() *ProvisionTest {
	return &ProvisionTest{
		datasets: []*comv1alpha1.Dataset{},
	}
}

func (r *ProvisionTest) CreateDataset(dataset *comv1alpha1.Dataset) error {
	for i, d := range r.datasets {
		if d.Name == dataset.Name {
			r.datasets[i] = dataset
			return nil
		}
	}
	r.datasets = append(r.datasets, dataset)
	return nil
}

func (r *ProvisionTest) GetDataset(ref *app.ResourceReference) (*comv1alpha1.Dataset, error) {
	for _, d := range r.datasets {
		if d.Name == ref.Name {
			d.Status.Provision.Status = "OK"
			return d, nil
		}
	}
	return nil, errors.New("Could not get a dataset: " + ref.Name)
}

func (r *ProvisionTest) DeleteDataset(ref *app.ResourceReference, force bool) error {
	newDatasets := []*comv1alpha1.Dataset{}
	found := false
	message := "Datasets:\n"
	for _, d := range r.datasets {
		message += " " + ref.Name + "\n"
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
	return errors.New("Could not delete a dataset: " + ref.Name + "\n" + message)
}

// AllocateBucket allocates a bucket in the relevant geo
func AllocateBucket(c client.Client, log logr.Logger, owner types.NamespacedName, id string, geo string) (*comv1alpha1.Dataset, error) {
	ctx := context.Background()
	log.Info("Searching for a storage account matching the geography " + geo)
	var accountList app.M4DStorageAccountList
	if err := c.List(ctx, &accountList); err != nil {
		log.Info(err.Error())
		return nil, err
	}
	for _, account := range accountList.Items {
		utils.PrintStructure(account, log, "Account ")
		if account.Spec.Region != geo {
			continue
		}
		genName := generateDatasetName(owner, id)
		values := map[string]string{
			"type":        "COS",
			"secret-name": account.Spec.SecretRef,
			"endpoint":    account.Spec.Endpoint,
			"bucket":      genName,
			"provision":   "true"}
		dataset := &comv1alpha1.Dataset{ObjectMeta: metav1.ObjectMeta{
			Name:      genName,
			Namespace: utils.GetSystemNamespace(),
			Labels:    map[string]string{"m4d.ibm.com/owner": owner.Namespace + "." + owner.Name},
		},
			Spec: comv1alpha1.DatasetSpec{Local: values},
		}
		return dataset, nil
	}
	return nil, errors.New("Could not allocate a bucket in " + geo)
}

func generateDatasetName(owner types.NamespacedName, id string) string {
	name := id + "-" + owner.Name + "-" + owner.Namespace
	return utils.K8sConformName(name)
}
