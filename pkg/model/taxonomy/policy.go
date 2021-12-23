// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import (
	"fybrik.io/fybrik/pkg/serde"
)

// +kubebuilder:pruning:PreserveUnknownFields
type PolicyManagerRequestContext struct {
	serde.Properties `json:"-"`
}
