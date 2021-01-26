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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=m4dstorageaccounts,verbs=get;list;watch;update;
// +kubebuilder:rbac:groups=com.ie.ibm.hpsys,resources=datasets,verbs=get;list;watch;create;update;patch;delete

// ProvisionInterface is an interface for managing Dataset resources
type ProvisionInterface interface {
	CreateDataset(dataset *comv1alpha1.Dataset) error
	//DeleteDataset(namespaced types.NamespacedName) error
	//GetStatus(namespaced types.NamespacedName) (*comv1alpha1.DatasetStatus, error)
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
	return r.Client.Create(context.Background(), dataset)
}

type ProvisionFake struct {
	datasets []*comv1alpha1.Dataset
}

func NewProvisionFake() *ProvisionFake {
	return &ProvisionFake{
		datasets: []*comv1alpha1.Dataset{},
	}
}

func (r *ProvisionFake) CreateDataset(dataset *comv1alpha1.Dataset) error {
	for _, d := range r.datasets {
		if d.Name == dataset.Name {
			return errors.New("Dataset exists")
		}
	}
	r.datasets = append(r.datasets, dataset)
	return nil
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
	name := owner.Name + "-" + owner.Namespace + "-" + utils.Hash(id, 20)
	return utils.K8sConformName(name)
}
