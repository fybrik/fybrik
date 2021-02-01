// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/ibm/the-mesh-for-data/pkg/storage"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=m4dstorageaccounts,verbs=get;list;watch;update;
// +kubebuilder:rbac:groups=com.ie.ibm.hpsys,resources=datasets,verbs=get;list;watch;create;update;patch;delete

// AllocateBucket allocates a bucket in the relevant geo
// The buckets are created as temporary, i.e. to be removed after the owner Dataset is deleted
// After a successful copy and registering a dataset, the bucket will become persistent
func AllocateBucket(c client.Client, log logr.Logger, owner types.NamespacedName, id string, geo string) (*storage.ProvisionedBucket, error) {
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
		return &storage.ProvisionedBucket{
			Name:      genName,
			Endpoint:  account.Spec.Endpoint,
			SecretRef: types.NamespacedName{Name: account.Spec.SecretRef, Namespace: utils.GetSystemNamespace()},
		}, nil
	}
	return nil, errors.New("Could not allocate a bucket in " + geo)
}

func generateDatasetName(owner types.NamespacedName, id string) string {
	name := id + "-" + owner.Name + "-" + owner.Namespace
	return utils.K8sConformName(name)
}
