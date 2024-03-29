//go:build !ignore_autogenerated

// Copyright 2023 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

// Code generated by controller-gen. DO NOT EDIT.

package v1beta2

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FybrikStorageAccount) DeepCopyInto(out *FybrikStorageAccount) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FybrikStorageAccount.
func (in *FybrikStorageAccount) DeepCopy() *FybrikStorageAccount {
	if in == nil {
		return nil
	}
	out := new(FybrikStorageAccount)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FybrikStorageAccount) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FybrikStorageAccountList) DeepCopyInto(out *FybrikStorageAccountList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]FybrikStorageAccount, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FybrikStorageAccountList.
func (in *FybrikStorageAccountList) DeepCopy() *FybrikStorageAccountList {
	if in == nil {
		return nil
	}
	out := new(FybrikStorageAccountList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FybrikStorageAccountList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FybrikStorageAccountSpec) DeepCopyInto(out *FybrikStorageAccountSpec) {
	*out = *in
	in.AdditionalProperties.DeepCopyInto(&out.AdditionalProperties)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FybrikStorageAccountSpec.
func (in *FybrikStorageAccountSpec) DeepCopy() *FybrikStorageAccountSpec {
	if in == nil {
		return nil
	}
	out := new(FybrikStorageAccountSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FybrikStorageAccountStatus) DeepCopyInto(out *FybrikStorageAccountStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FybrikStorageAccountStatus.
func (in *FybrikStorageAccountStatus) DeepCopy() *FybrikStorageAccountStatus {
	if in == nil {
		return nil
	}
	out := new(FybrikStorageAccountStatus)
	in.DeepCopyInto(out)
	return out
}
