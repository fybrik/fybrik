// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/pkg/storage"
	"github.com/rs/zerolog"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/logging"
	"k8s.io/apimachinery/pkg/types"
)

// AllocateBucket allocates a bucket in the relevant geo
// The buckets are created as temporary, i.e. to be removed after the owner Dataset is deleted
// After a successful copy and registering a dataset, the bucket will become persistent
func AllocateBucket(c client.Client, log zerolog.Logger, bucketName string, geo string) (*storage.ProvisionedBucket, error) {
	ctx := context.Background()
	log.Trace().Msg("Searching for a storage account matching the geography " + geo)
	var accountList app.FybrikStorageAccountList
	if err := c.List(ctx, &accountList, client.InNamespace(utils.GetSystemNamespace())); err != nil {
		log.Error().Err(err).Msg("Error listing storage accounts")
		return nil, err
	}

	for _, account := range accountList.Items {
		logging.LogStructure("Account", account, log, false, false)
		for key, value := range account.Spec.Endpoints {
			if geo != key {
				continue
			}
			return &storage.ProvisionedBucket{
				Name:      bucketName,
				Endpoint:  value,
				SecretRef: types.NamespacedName{Name: account.Spec.SecretRef, Namespace: utils.GetSystemNamespace()},
			}, nil
		}
	}
	return nil, fmt.Errorf("could not allocate a bucket in %s", geo)
}
