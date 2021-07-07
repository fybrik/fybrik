// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	app "github.com/mesh-for-data/mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/mesh-for-data/mesh-for-data/pkg/storage"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/mesh-for-data/mesh-for-data/manager/controllers/utils"
	"k8s.io/apimachinery/pkg/types"
)

func includesGeography(array []string, element string) bool {
	for _, geo := range array {
		if geo == element {
			return true
		}
	}
	return false
}

// AllocateBucket allocates a bucket in the relevant geo
// The buckets are created as temporary, i.e. to be removed after the owner Dataset is deleted
// After a successful copy and registering a dataset, the bucket will become persistent
func AllocateBucket(c client.Client, log logr.Logger, owner types.NamespacedName, id string, geo string) (*storage.ProvisionedBucket, error) {
	ctx := context.Background()
	log.Info("Searching for a storage account matching the geography " + geo)
	var accountList app.M4DStorageAccountList
	if err := c.List(ctx, &accountList, client.InNamespace(utils.GetSystemNamespace())); err != nil {
		log.Info(err.Error())
		return nil, err
	}
	for _, account := range accountList.Items {
		utils.PrintStructure(account, log, "Account ")
		if !includesGeography(account.Spec.Regions, geo) {
			continue
		}
		genName := generateDatasetName(owner, id)
		return &storage.ProvisionedBucket{
			Name:      genName,
			Endpoint:  account.Spec.Endpoint,
			SecretRef: types.NamespacedName{Name: account.Spec.SecretRef, Namespace: utils.GetSystemNamespace()},
		}, nil
	}
	return nil, fmt.Errorf("could not allocate a bucket in %s", geo)
}

func generateDatasetName(owner types.NamespacedName, id string) string {
	name := owner.Name + "-" + owner.Namespace + utils.Hash(id, 10)
	name = strings.ReplaceAll(name, ".", "-")
	return utils.K8sConformName(name)
}
