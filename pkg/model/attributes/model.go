// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package infraattributes

import (
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

type Infrastructure struct {
	Definitions []taxonomy.AttributeDefinition   `json:"definitions"`
	Items       []taxonomy.InfrastructureElement `json:"infrastructure"`
}
