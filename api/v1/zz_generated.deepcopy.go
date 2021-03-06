// +build !ignore_autogenerated

/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EscalationPolicySecretSpec) DeepCopyInto(out *EscalationPolicySecretSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EscalationPolicySecretSpec.
func (in *EscalationPolicySecretSpec) DeepCopy() *EscalationPolicySecretSpec {
	if in == nil {
		return nil
	}
	out := new(EscalationPolicySecretSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LabelSpec) DeepCopyInto(out *LabelSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LabelSpec.
func (in *LabelSpec) DeepCopy() *LabelSpec {
	if in == nil {
		return nil
	}
	out := new(LabelSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PagerdutyRuleset) DeepCopyInto(out *PagerdutyRuleset) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PagerdutyRuleset.
func (in *PagerdutyRuleset) DeepCopy() *PagerdutyRuleset {
	if in == nil {
		return nil
	}
	out := new(PagerdutyRuleset)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PagerdutyRuleset) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PagerdutyRulesetList) DeepCopyInto(out *PagerdutyRulesetList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PagerdutyRuleset, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PagerdutyRulesetList.
func (in *PagerdutyRulesetList) DeepCopy() *PagerdutyRulesetList {
	if in == nil {
		return nil
	}
	out := new(PagerdutyRulesetList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PagerdutyRulesetList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PagerdutyRulesetSpec) DeepCopyInto(out *PagerdutyRulesetSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PagerdutyRulesetSpec.
func (in *PagerdutyRulesetSpec) DeepCopy() *PagerdutyRulesetSpec {
	if in == nil {
		return nil
	}
	out := new(PagerdutyRulesetSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PagerdutyRulesetStatus) DeepCopyInto(out *PagerdutyRulesetStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PagerdutyRulesetStatus.
func (in *PagerdutyRulesetStatus) DeepCopy() *PagerdutyRulesetStatus {
	if in == nil {
		return nil
	}
	out := new(PagerdutyRulesetStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PagerdutyService) DeepCopyInto(out *PagerdutyService) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PagerdutyService.
func (in *PagerdutyService) DeepCopy() *PagerdutyService {
	if in == nil {
		return nil
	}
	out := new(PagerdutyService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PagerdutyService) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PagerdutyServiceList) DeepCopyInto(out *PagerdutyServiceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PagerdutyService, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PagerdutyServiceList.
func (in *PagerdutyServiceList) DeepCopy() *PagerdutyServiceList {
	if in == nil {
		return nil
	}
	out := new(PagerdutyServiceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PagerdutyServiceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PagerdutyServiceSpec) DeepCopyInto(out *PagerdutyServiceSpec) {
	*out = *in
	out.EscalationPolicySecret = in.EscalationPolicySecret
	if in.MatchLabels != nil {
		in, out := &in.MatchLabels, &out.MatchLabels
		*out = make([]LabelSpec, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PagerdutyServiceSpec.
func (in *PagerdutyServiceSpec) DeepCopy() *PagerdutyServiceSpec {
	if in == nil {
		return nil
	}
	out := new(PagerdutyServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PagerdutyServiceStatus) DeepCopyInto(out *PagerdutyServiceStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PagerdutyServiceStatus.
func (in *PagerdutyServiceStatus) DeepCopy() *PagerdutyServiceStatus {
	if in == nil {
		return nil
	}
	out := new(PagerdutyServiceStatus)
	in.DeepCopyInto(out)
	return out
}
