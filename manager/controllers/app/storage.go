// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fybrik.io/fybrik/pkg/model/taxonomy"
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
func AllocateBucket(c client.Client, log zerolog.Logger, bucketName string, account *taxonomy.StorageAccount) *storage.ProvisionedBucket {
	logging.LogStructure("Account", account, log, false, false)
	if account == nil {
		return nil
	}
	return &storage.ProvisionedBucket{
		Name:      bucketName,
		Endpoint:  account.Endpoint,
		SecretRef: types.NamespacedName{Name: account.SecretRef, Namespace: utils.GetSystemNamespace()},
	}
}
