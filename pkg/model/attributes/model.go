// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package infraattributes

import (
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

type Infrastructure struct {
	// a list of infrastructure metrics including scale and units
	// shared by various attributes
	Metrics []taxonomy.InfrastructureMetrics `json:"metrics,omitempty"`
	// a list of infrastructure arguments
	Attributes []taxonomy.InfrastructureElement `json:"infrastructure"`
}
