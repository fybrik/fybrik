// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package base

import (
	"github.com/mohae/deepcopy"
)

func (in *Connection) DeepCopyInto(out *Connection) {
	copy, _ := deepcopy.Copy(in).(*Connection)
	*out = *copy
}

func (in *Resource) DeepCopyInto(out *Resource) {
	copy, _ := deepcopy.Copy(in).(*Resource)
	*out = *copy
}

func (in *DataCatalogResponse) DeepCopyInto(out *DataCatalogResponse) {
	copy, _ := deepcopy.Copy(in).(*DataCatalogResponse)
	*out = *copy
}
