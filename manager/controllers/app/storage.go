// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"strings"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/pkg/storage"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"fybrik.io/fybrik/manager/controllers/utils"
	"k8s.io/apimachinery/pkg/types"
)

// AllocateBucket allocates a bucket in the relevant geo
// The buckets are created as temporary, i.e. to be removed after the owner Dataset is deleted
// After a successful copy and registering a dataset, the bucket will become persistent
func AllocateBucket(c client.Client, log logr.Logger, owner types.NamespacedName, id string, geo string) (*storage.ProvisionedBucket, error) {
	ctx := context.Background()
	log.Info("Searching for a storage account matching the geography " + geo)
	var accountList app.FybrikStorageAccountList
	if err := c.List(ctx, &accountList, client.InNamespace(utils.GetSystemNamespace())); err != nil {
		log.Info(err.Error())
		return nil, err
	}

	for _, account := range accountList.Items {
		utils.PrintStructure(account, log, "Account ")
		for key, value := range account.Spec.Endpoints {
			if geo != key {
				continue
			}
			genName := generateDatasetName(owner, id)
			return &storage.ProvisionedBucket{
				Name:      genName,
				Endpoint:  value,
				SecretRef: types.NamespacedName{Name: account.Spec.SecretRef, Namespace: utils.GetSystemNamespace()},
			}, nil
		}
	}
	return nil, fmt.Errorf("could not allocate a bucket in %s", geo)
}

func generateDatasetName(owner types.NamespacedName, id string) string {
	name := owner.Name + "-" + owner.Namespace + utils.Hash(id, 10)
	name = strings.ReplaceAll(name, ".", "-")
	return utils.K8sConformName(name)
}
