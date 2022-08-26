//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2022.

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
func (in *DefaultResources) DeepCopyInto(out *DefaultResources) {
	*out = *in
	if in.Request != nil {
		in, out := &in.Request, &out.Request
		*out = new(Resources)
		(*in).DeepCopyInto(*out)
	}
	if in.Limit != nil {
		in, out := &in.Limit, &out.Limit
		*out = new(Resources)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DefaultResources.
func (in *DefaultResources) DeepCopy() *DefaultResources {
	if in == nil {
		return nil
	}
	out := new(DefaultResources)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OrgNamespace) DeepCopyInto(out *OrgNamespace) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OrgNamespace.
func (in *OrgNamespace) DeepCopy() *OrgNamespace {
	if in == nil {
		return nil
	}
	out := new(OrgNamespace)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OrgNamespace) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OrgNamespaceList) DeepCopyInto(out *OrgNamespaceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OrgNamespace, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OrgNamespaceList.
func (in *OrgNamespaceList) DeepCopy() *OrgNamespaceList {
	if in == nil {
		return nil
	}
	out := new(OrgNamespaceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OrgNamespaceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OrgNamespaceSpec) DeepCopyInto(out *OrgNamespaceSpec) {
	*out = *in
	if in.ImportSecrets != nil {
		in, out := &in.ImportSecrets, &out.ImportSecrets
		*out = make([]SecretRef, len(*in))
		copy(*out, *in)
	}
	if in.DefaultResources != nil {
		in, out := &in.DefaultResources, &out.DefaultResources
		*out = new(DefaultResources)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OrgNamespaceSpec.
func (in *OrgNamespaceSpec) DeepCopy() *OrgNamespaceSpec {
	if in == nil {
		return nil
	}
	out := new(OrgNamespaceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OrgNamespaceStatus) DeepCopyInto(out *OrgNamespaceStatus) {
	*out = *in
	in.NamespaceStatus.DeepCopyInto(&out.NamespaceStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OrgNamespaceStatus.
func (in *OrgNamespaceStatus) DeepCopy() *OrgNamespaceStatus {
	if in == nil {
		return nil
	}
	out := new(OrgNamespaceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Resources) DeepCopyInto(out *Resources) {
	*out = *in
	if in.CPU != nil {
		in, out := &in.CPU, &out.CPU
		x := (*in).DeepCopy()
		*out = &x
	}
	if in.Memory != nil {
		in, out := &in.Memory, &out.Memory
		x := (*in).DeepCopy()
		*out = &x
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Resources.
func (in *Resources) DeepCopy() *Resources {
	if in == nil {
		return nil
	}
	out := new(Resources)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretRef) DeepCopyInto(out *SecretRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretRef.
func (in *SecretRef) DeepCopy() *SecretRef {
	if in == nil {
		return nil
	}
	out := new(SecretRef)
	in.DeepCopyInto(out)
	return out
}
