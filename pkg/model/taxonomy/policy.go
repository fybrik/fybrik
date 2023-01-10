// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import (
	"fybrik.io/fybrik/pkg/serde"
)

// Context in which a policy is evaluated, e.g., details of the data user such as role and intent
// +kubebuilder:pruning:PreserveUnknownFields
type PolicyManagerRequestContext struct {
	serde.Properties `json:"-"`
}
