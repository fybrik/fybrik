// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import "fybrik.io/fybrik/pkg/serde"

// Application specific properties, e.g., intent for using the data, user role and workload characteristics
// +kubebuilder:pruning:PreserveUnknownFields
type AppInfo struct {
	serde.Properties `json:"-"`
}
