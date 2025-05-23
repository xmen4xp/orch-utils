//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

/*
Copyright The Kubernetes Authors.

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1

import (
	gnstsmtanzuvmwarecomv1 "github.com/vmware-tanzu/graph-framework-for-microservices/compiler/example/output/generated/apis/gns.tsm.tanzu.vmware.com/v1"

	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in AMap) DeepCopyInto(out *AMap) {
	{
		in := &in
		*out = make(AMap, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AMap.
func (in AMap) DeepCopy() AMap {
	if in == nil {
		return nil
	}
	out := new(AMap)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in BArray) DeepCopyInto(out *BArray) {
	{
		in := &in
		*out = make(BArray, len(*in))
		copy(*out, *in)
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BArray.
func (in BArray) DeepCopy() BArray {
	if in == nil {
		return nil
	}
	out := new(BArray)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Child) DeepCopyInto(out *Child) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Child.
func (in *Child) DeepCopy() *Child {
	if in == nil {
		return nil
	}
	out := new(Child)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Cluster) DeepCopyInto(out *Cluster) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Cluster.
func (in *Cluster) DeepCopy() *Cluster {
	if in == nil {
		return nil
	}
	out := new(Cluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterNamespace) DeepCopyInto(out *ClusterNamespace) {
	*out = *in
	out.Cluster = in.Cluster
	out.Namespace = in.Namespace
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterNamespace.
func (in *ClusterNamespace) DeepCopy() *ClusterNamespace {
	if in == nil {
		return nil
	}
	out := new(ClusterNamespace)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Config) DeepCopyInto(out *Config) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Config.
func (in *Config) DeepCopy() *Config {
	if in == nil {
		return nil
	}
	out := new(Config)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Config) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConfigList) DeepCopyInto(out *ConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Config, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConfigList.
func (in *ConfigList) DeepCopy() *ConfigList {
	if in == nil {
		return nil
	}
	out := new(ConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConfigNexusStatus) DeepCopyInto(out *ConfigNexusStatus) {
	*out = *in
	out.Nexus = in.Nexus
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConfigNexusStatus.
func (in *ConfigNexusStatus) DeepCopy() *ConfigNexusStatus {
	if in == nil {
		return nil
	}
	out := new(ConfigNexusStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConfigSpec) DeepCopyInto(out *ConfigSpec) {
	*out = *in
	if in.MyStr0 != nil {
		in, out := &in.MyStr0, &out.MyStr0
		*out = new(gnstsmtanzuvmwarecomv1.MyStr)
		**out = **in
	}
	if in.MyStr1 != nil {
		in, out := &in.MyStr1, &out.MyStr1
		*out = make([]gnstsmtanzuvmwarecomv1.MyStr, len(*in))
		copy(*out, *in)
	}
	if in.MyStr2 != nil {
		in, out := &in.MyStr2, &out.MyStr2
		*out = make(map[string]gnstsmtanzuvmwarecomv1.MyStr, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.ABCHost != nil {
		in, out := &in.ABCHost, &out.ABCHost
		*out = make([]gnstsmtanzuvmwarecomv1.Host, len(*in))
		copy(*out, *in)
	}
	if in.ClusterNamespaces != nil {
		in, out := &in.ClusterNamespaces, &out.ClusterNamespaces
		*out = make([]ClusterNamespace, len(*in))
		copy(*out, *in)
	}
	in.TestValMarkers.DeepCopyInto(&out.TestValMarkers)
	if in.GNSGvk != nil {
		in, out := &in.GNSGvk, &out.GNSGvk
		*out = new(Child)
		**out = **in
	}
	if in.DNSGvk != nil {
		in, out := &in.DNSGvk, &out.DNSGvk
		*out = new(Child)
		**out = **in
	}
	if in.VMPPoliciesGvk != nil {
		in, out := &in.VMPPoliciesGvk, &out.VMPPoliciesGvk
		*out = new(Child)
		**out = **in
	}
	if in.DomainGvk != nil {
		in, out := &in.DomainGvk, &out.DomainGvk
		*out = new(Child)
		**out = **in
	}
	if in.FooExampleGvk != nil {
		in, out := &in.FooExampleGvk, &out.FooExampleGvk
		*out = make(map[string]Child, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.SvcGrpInfoGvk != nil {
		in, out := &in.SvcGrpInfoGvk, &out.SvcGrpInfoGvk
		*out = new(Child)
		**out = **in
	}
	if in.ACPPoliciesGvk != nil {
		in, out := &in.ACPPoliciesGvk, &out.ACPPoliciesGvk
		*out = make(map[string]Link, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConfigSpec.
func (in *ConfigSpec) DeepCopy() *ConfigSpec {
	if in == nil {
		return nil
	}
	out := new(ConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CrossPackageTester) DeepCopyInto(out *CrossPackageTester) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CrossPackageTester.
func (in *CrossPackageTester) DeepCopy() *CrossPackageTester {
	if in == nil {
		return nil
	}
	out := new(CrossPackageTester)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Domain) DeepCopyInto(out *Domain) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Domain.
func (in *Domain) DeepCopy() *Domain {
	if in == nil {
		return nil
	}
	out := new(Domain)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Domain) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DomainList) DeepCopyInto(out *DomainList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Domain, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DomainList.
func (in *DomainList) DeepCopy() *DomainList {
	if in == nil {
		return nil
	}
	out := new(DomainList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DomainList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DomainNexusStatus) DeepCopyInto(out *DomainNexusStatus) {
	*out = *in
	out.Nexus = in.Nexus
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DomainNexusStatus.
func (in *DomainNexusStatus) DeepCopy() *DomainNexusStatus {
	if in == nil {
		return nil
	}
	out := new(DomainNexusStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DomainSpec) DeepCopyInto(out *DomainSpec) {
	*out = *in
	if in.PointPort != nil {
		in, out := &in.PointPort, &out.PointPort
		*out = new(gnstsmtanzuvmwarecomv1.Port)
		**out = **in
	}
	if in.PointString != nil {
		in, out := &in.PointString, &out.PointString
		*out = new(string)
		**out = **in
	}
	if in.PointInt != nil {
		in, out := &in.PointInt, &out.PointInt
		*out = new(int)
		**out = **in
	}
	if in.PointMap != nil {
		in, out := &in.PointMap, &out.PointMap
		*out = new(map[string]string)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]string, len(*in))
			for key, val := range *in {
				(*out)[key] = val
			}
		}
	}
	if in.PointSlice != nil {
		in, out := &in.PointSlice, &out.PointSlice
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.SliceOfPoints != nil {
		in, out := &in.SliceOfPoints, &out.SliceOfPoints
		*out = make([]*string, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(string)
				**out = **in
			}
		}
	}
	if in.SliceOfArrPoints != nil {
		in, out := &in.SliceOfArrPoints, &out.SliceOfArrPoints
		*out = make([]*BArray, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(BArray)
				if **in != nil {
					in, out := *in, *out
					*out = make([]string, len(*in))
					copy(*out, *in)
				}
			}
		}
	}
	if in.MapOfArrsPoints != nil {
		in, out := &in.MapOfArrsPoints, &out.MapOfArrsPoints
		*out = make(map[string]*BArray, len(*in))
		for key, val := range *in {
			var outVal *BArray
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(BArray)
				if **in != nil {
					in, out := *in, *out
					*out = make([]string, len(*in))
					copy(*out, *in)
				}
			}
			(*out)[key] = outVal
		}
	}
	if in.PointStruct != nil {
		in, out := &in.PointStruct, &out.PointStruct
		*out = new(Cluster)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DomainSpec.
func (in *DomainSpec) DeepCopy() *DomainSpec {
	if in == nil {
		return nil
	}
	out := new(DomainSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EmptyStructTest) DeepCopyInto(out *EmptyStructTest) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EmptyStructTest.
func (in *EmptyStructTest) DeepCopy() *EmptyStructTest {
	if in == nil {
		return nil
	}
	out := new(EmptyStructTest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FooTypeABC) DeepCopyInto(out *FooTypeABC) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FooTypeABC.
func (in *FooTypeABC) DeepCopy() *FooTypeABC {
	if in == nil {
		return nil
	}
	out := new(FooTypeABC)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FooTypeABC) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FooTypeABCList) DeepCopyInto(out *FooTypeABCList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]FooTypeABC, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FooTypeABCList.
func (in *FooTypeABCList) DeepCopy() *FooTypeABCList {
	if in == nil {
		return nil
	}
	out := new(FooTypeABCList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FooTypeABCList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FooTypeABCNexusStatus) DeepCopyInto(out *FooTypeABCNexusStatus) {
	*out = *in
	out.Nexus = in.Nexus
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FooTypeABCNexusStatus.
func (in *FooTypeABCNexusStatus) DeepCopy() *FooTypeABCNexusStatus {
	if in == nil {
		return nil
	}
	out := new(FooTypeABCNexusStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FooTypeABCSpec) DeepCopyInto(out *FooTypeABCSpec) {
	*out = *in
	if in.FooA != nil {
		in, out := &in.FooA, &out.FooA
		*out = make(AMap, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.FooB != nil {
		in, out := &in.FooB, &out.FooB
		*out = make(BArray, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FooTypeABCSpec.
func (in *FooTypeABCSpec) DeepCopy() *FooTypeABCSpec {
	if in == nil {
		return nil
	}
	out := new(FooTypeABCSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Link) DeepCopyInto(out *Link) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Link.
func (in *Link) DeepCopy() *Link {
	if in == nil {
		return nil
	}
	out := new(Link)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MatchCondition) DeepCopyInto(out *MatchCondition) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MatchCondition.
func (in *MatchCondition) DeepCopy() *MatchCondition {
	if in == nil {
		return nil
	}
	out := new(MatchCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NexusStatus) DeepCopyInto(out *NexusStatus) {
	*out = *in
	out.SyncerStatus = in.SyncerStatus
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NexusStatus.
func (in *NexusStatus) DeepCopy() *NexusStatus {
	if in == nil {
		return nil
	}
	out := new(NexusStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SomeStruct) DeepCopyInto(out *SomeStruct) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SomeStruct.
func (in *SomeStruct) DeepCopy() *SomeStruct {
	if in == nil {
		return nil
	}
	out := new(SomeStruct)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StructWithEmbeddedField) DeepCopyInto(out *StructWithEmbeddedField) {
	*out = *in
	out.SomeStruct = in.SomeStruct
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StructWithEmbeddedField.
func (in *StructWithEmbeddedField) DeepCopy() *StructWithEmbeddedField {
	if in == nil {
		return nil
	}
	out := new(StructWithEmbeddedField)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SyncerStatus) DeepCopyInto(out *SyncerStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SyncerStatus.
func (in *SyncerStatus) DeepCopy() *SyncerStatus {
	if in == nil {
		return nil
	}
	out := new(SyncerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TestValMarkers) DeepCopyInto(out *TestValMarkers) {
	*out = *in
	if in.MySlice != nil {
		in, out := &in.MySlice, &out.MySlice
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TestValMarkers.
func (in *TestValMarkers) DeepCopy() *TestValMarkers {
	if in == nil {
		return nil
	}
	out := new(TestValMarkers)
	in.DeepCopyInto(out)
	return out
}
