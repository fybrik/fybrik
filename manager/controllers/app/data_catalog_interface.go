// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"errors"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
)

// RegisterAsset registers a new asset in the specified catalog
// Input arguments:
// - catalogID: the destination catalog identifier
// - info: connection and credential details
// Returns:
// - an error if happened
// - the new asset identifier
func (r *FybrikApplicationReconciler) RegisterAsset(catalogID string, info *app.DatasetDetails,
	input *app.FybrikApplication) (string, error) {
	return "", errors.New("unsupported feature")
}
