//go:build !ignore_autogenerated

// © Broadcom. All Rights Reserved.
// The term “Broadcom” refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AvailabilityZone) DeepCopyInto(out *AvailabilityZone) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AvailabilityZone.
func (in *AvailabilityZone) DeepCopy() *AvailabilityZone {
	if in == nil {
		return nil
	}
	out := new(AvailabilityZone)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AvailabilityZone) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AvailabilityZoneList) DeepCopyInto(out *AvailabilityZoneList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AvailabilityZone, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AvailabilityZoneList.
func (in *AvailabilityZoneList) DeepCopy() *AvailabilityZoneList {
	if in == nil {
		return nil
	}
	out := new(AvailabilityZoneList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AvailabilityZoneList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AvailabilityZoneReference) DeepCopyInto(out *AvailabilityZoneReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AvailabilityZoneReference.
func (in *AvailabilityZoneReference) DeepCopy() *AvailabilityZoneReference {
	if in == nil {
		return nil
	}
	out := new(AvailabilityZoneReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AvailabilityZoneSpec) DeepCopyInto(out *AvailabilityZoneSpec) {
	*out = *in
	if in.ClusterComputeResourceMoIDs != nil {
		in, out := &in.ClusterComputeResourceMoIDs, &out.ClusterComputeResourceMoIDs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Namespaces != nil {
		in, out := &in.Namespaces, &out.Namespaces
		*out = make(map[string]NamespaceInfo, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AvailabilityZoneSpec.
func (in *AvailabilityZoneSpec) DeepCopy() *AvailabilityZoneSpec {
	if in == nil {
		return nil
	}
	out := new(AvailabilityZoneSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AvailabilityZoneStatus) DeepCopyInto(out *AvailabilityZoneStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AvailabilityZoneStatus.
func (in *AvailabilityZoneStatus) DeepCopy() *AvailabilityZoneStatus {
	if in == nil {
		return nil
	}
	out := new(AvailabilityZoneStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NamespaceInfo) DeepCopyInto(out *NamespaceInfo) {
	*out = *in
	if in.PoolMoIDs != nil {
		in, out := &in.PoolMoIDs, &out.PoolMoIDs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NamespaceInfo.
func (in *NamespaceInfo) DeepCopy() *NamespaceInfo {
	if in == nil {
		return nil
	}
	out := new(NamespaceInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereEntityInfo) DeepCopyInto(out *VSphereEntityInfo) {
	*out = *in
	if in.PoolMoIDs != nil {
		in, out := &in.PoolMoIDs, &out.PoolMoIDs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VSphereEntityInfo.
func (in *VSphereEntityInfo) DeepCopy() *VSphereEntityInfo {
	if in == nil {
		return nil
	}
	out := new(VSphereEntityInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereZone) DeepCopyInto(out *VSphereZone) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VSphereZone.
func (in *VSphereZone) DeepCopy() *VSphereZone {
	if in == nil {
		return nil
	}
	out := new(VSphereZone)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VSphereZone) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereZoneList) DeepCopyInto(out *VSphereZoneList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VSphereZone, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VSphereZoneList.
func (in *VSphereZoneList) DeepCopy() *VSphereZoneList {
	if in == nil {
		return nil
	}
	out := new(VSphereZoneList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VSphereZoneList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereZoneSpec) DeepCopyInto(out *VSphereZoneSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VSphereZoneSpec.
func (in *VSphereZoneSpec) DeepCopy() *VSphereZoneSpec {
	if in == nil {
		return nil
	}
	out := new(VSphereZoneSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VSphereZoneStatus) DeepCopyInto(out *VSphereZoneStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VSphereZoneStatus.
func (in *VSphereZoneStatus) DeepCopy() *VSphereZoneStatus {
	if in == nil {
		return nil
	}
	out := new(VSphereZoneStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Zone) DeepCopyInto(out *Zone) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Zone.
func (in *Zone) DeepCopy() *Zone {
	if in == nil {
		return nil
	}
	out := new(Zone)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Zone) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZoneList) DeepCopyInto(out *ZoneList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Zone, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZoneList.
func (in *ZoneList) DeepCopy() *ZoneList {
	if in == nil {
		return nil
	}
	out := new(ZoneList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ZoneList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZoneSpec) DeepCopyInto(out *ZoneSpec) {
	*out = *in
	in.Namespace.DeepCopyInto(&out.Namespace)
	in.VSpherePods.DeepCopyInto(&out.VSpherePods)
	in.ManagedVMs.DeepCopyInto(&out.ManagedVMs)
	out.Zone = in.Zone
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZoneSpec.
func (in *ZoneSpec) DeepCopy() *ZoneSpec {
	if in == nil {
		return nil
	}
	out := new(ZoneSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZoneStatus) DeepCopyInto(out *ZoneStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZoneStatus.
func (in *ZoneStatus) DeepCopy() *ZoneStatus {
	if in == nil {
		return nil
	}
	out := new(ZoneStatus)
	in.DeepCopyInto(out)
	return out
}
