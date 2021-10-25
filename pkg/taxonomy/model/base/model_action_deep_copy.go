// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package base

import "github.com/mohae/deepcopy"

func (in *Action) DeepCopyInto(out *Action) {
	copy, _ := deepcopy.Copy(in).(*Action)
	*out = *copy
}
