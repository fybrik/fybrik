// +build !ignore_autogenerated

// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

// Code generated by controller-gen. DO NOT EDIT.

package adminrules

import ()

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Decision) DeepCopyInto(out *Decision) {
	*out = *in
	in.DeploymentRestrictions.DeepCopyInto(&out.DeploymentRestrictions)
	out.Policy = in.Policy
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Decision.
func (in *Decision) DeepCopy() *Decision {
	if in == nil {
		return nil
	}
	out := new(Decision)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in DecisionPerCapabilityMap) DeepCopyInto(out *DecisionPerCapabilityMap) {
	{
		in := &in
		*out = make(DecisionPerCapabilityMap, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DecisionPerCapabilityMap.
func (in DecisionPerCapabilityMap) DeepCopy() DecisionPerCapabilityMap {
	if in == nil {
		return nil
	}
	out := new(DecisionPerCapabilityMap)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DecisionPolicy) DeepCopyInto(out *DecisionPolicy) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DecisionPolicy.
func (in *DecisionPolicy) DeepCopy() *DecisionPolicy {
	if in == nil {
		return nil
	}
	out := new(DecisionPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EvaluationOutputStructure) DeepCopyInto(out *EvaluationOutputStructure) {
	*out = *in
	if in.Config != nil {
		in, out := &in.Config, &out.Config
		*out = make(RuleDecisionList, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = make(DecisionPerCapabilityMap, len(*in))
				for key, val := range *in {
					(*out)[key] = *val.DeepCopy()
				}
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EvaluationOutputStructure.
func (in *EvaluationOutputStructure) DeepCopy() *EvaluationOutputStructure {
	if in == nil {
		return nil
	}
	out := new(EvaluationOutputStructure)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Restriction) DeepCopyInto(out *Restriction) {
	{
		in := &in
		*out = make(Restriction, len(*in))
		for key, val := range *in {
			var outVal []string
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = make([]string, len(*in))
				copy(*out, *in)
			}
			(*out)[key] = outVal
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Restriction.
func (in Restriction) DeepCopy() Restriction {
	if in == nil {
		return nil
	}
	out := new(Restriction)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Restrictions) DeepCopyInto(out *Restrictions) {
	*out = *in
	if in.Clusters != nil {
		in, out := &in.Clusters, &out.Clusters
		*out = make(Restriction, len(*in))
		for key, val := range *in {
			var outVal []string
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = make([]string, len(*in))
				copy(*out, *in)
			}
			(*out)[key] = outVal
		}
	}
	if in.Modules != nil {
		in, out := &in.Modules, &out.Modules
		*out = make(Restriction, len(*in))
		for key, val := range *in {
			var outVal []string
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = make([]string, len(*in))
				copy(*out, *in)
			}
			(*out)[key] = outVal
		}
	}
	if in.StorageAccounts != nil {
		in, out := &in.StorageAccounts, &out.StorageAccounts
		*out = make(Restriction, len(*in))
		for key, val := range *in {
			var outVal []string
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = make([]string, len(*in))
				copy(*out, *in)
			}
			(*out)[key] = outVal
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Restrictions.
func (in *Restrictions) DeepCopy() *Restrictions {
	if in == nil {
		return nil
	}
	out := new(Restrictions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in RuleDecisionList) DeepCopyInto(out *RuleDecisionList) {
	{
		in := &in
		*out = make(RuleDecisionList, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = make(DecisionPerCapabilityMap, len(*in))
				for key, val := range *in {
					(*out)[key] = *val.DeepCopy()
				}
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RuleDecisionList.
func (in RuleDecisionList) DeepCopy() RuleDecisionList {
	if in == nil {
		return nil
	}
	out := new(RuleDecisionList)
	in.DeepCopyInto(out)
	return *out
}
