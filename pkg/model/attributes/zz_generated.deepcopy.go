//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

// Code generated by controller-gen. DO NOT EDIT.

package attributes

import (
	"fybrik.io/fybrik/pkg/model/adminrules"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Infrastructure) DeepCopyInto(out *Infrastructure) {
	*out = *in
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]InfrastructureElement, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Infrastructure.
func (in *Infrastructure) DeepCopy() *Infrastructure {
	if in == nil {
		return nil
	}
	out := new(Infrastructure)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InfrastructureElement) DeepCopyInto(out *InfrastructureElement) {
	*out = *in
	if in.Scale != nil {
		in, out := &in.Scale, &out.Scale
		*out = new(adminrules.RangeType)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InfrastructureElement.
func (in *InfrastructureElement) DeepCopy() *InfrastructureElement {
	if in == nil {
		return nil
	}
	out := new(InfrastructureElement)
	in.DeepCopyInto(out)
	return out
}
