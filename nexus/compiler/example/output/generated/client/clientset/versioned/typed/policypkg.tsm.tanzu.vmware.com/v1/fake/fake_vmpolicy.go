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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"
	policypkgtsmtanzuvmwarecomv1 "github.com/vmware-tanzu/graph-framework-for-microservices/compiler/example/output/generated/apis/policypkg.tsm.tanzu.vmware.com/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeVMpolicies implements VMpolicyInterface
type FakeVMpolicies struct {
	Fake *FakePolicypkgTsmV1
}

var vmpoliciesResource = schema.GroupVersionResource{Group: "policypkg.tsm.tanzu.vmware.com", Version: "v1", Resource: "vmpolicies"}

var vmpoliciesKind = schema.GroupVersionKind{Group: "policypkg.tsm.tanzu.vmware.com", Version: "v1", Kind: "VMpolicy"}

// Get takes name of the vMpolicy, and returns the corresponding vMpolicy object, and an error if there is any.
func (c *FakeVMpolicies) Get(ctx context.Context, name string, options v1.GetOptions) (result *policypkgtsmtanzuvmwarecomv1.VMpolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(vmpoliciesResource, name), &policypkgtsmtanzuvmwarecomv1.VMpolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*policypkgtsmtanzuvmwarecomv1.VMpolicy), err
}

// List takes label and field selectors, and returns the list of VMpolicies that match those selectors.
func (c *FakeVMpolicies) List(ctx context.Context, opts v1.ListOptions) (result *policypkgtsmtanzuvmwarecomv1.VMpolicyList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(vmpoliciesResource, vmpoliciesKind, opts), &policypkgtsmtanzuvmwarecomv1.VMpolicyList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &policypkgtsmtanzuvmwarecomv1.VMpolicyList{ListMeta: obj.(*policypkgtsmtanzuvmwarecomv1.VMpolicyList).ListMeta}
	for _, item := range obj.(*policypkgtsmtanzuvmwarecomv1.VMpolicyList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested vMpolicies.
func (c *FakeVMpolicies) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(vmpoliciesResource, opts))
}

// Create takes the representation of a vMpolicy and creates it.  Returns the server's representation of the vMpolicy, and an error, if there is any.
func (c *FakeVMpolicies) Create(ctx context.Context, vMpolicy *policypkgtsmtanzuvmwarecomv1.VMpolicy, opts v1.CreateOptions) (result *policypkgtsmtanzuvmwarecomv1.VMpolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(vmpoliciesResource, vMpolicy), &policypkgtsmtanzuvmwarecomv1.VMpolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*policypkgtsmtanzuvmwarecomv1.VMpolicy), err
}

// Update takes the representation of a vMpolicy and updates it. Returns the server's representation of the vMpolicy, and an error, if there is any.
func (c *FakeVMpolicies) Update(ctx context.Context, vMpolicy *policypkgtsmtanzuvmwarecomv1.VMpolicy, opts v1.UpdateOptions) (result *policypkgtsmtanzuvmwarecomv1.VMpolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(vmpoliciesResource, vMpolicy), &policypkgtsmtanzuvmwarecomv1.VMpolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*policypkgtsmtanzuvmwarecomv1.VMpolicy), err
}

// Delete takes name of the vMpolicy and deletes it. Returns an error if one occurs.
func (c *FakeVMpolicies) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(vmpoliciesResource, name, opts), &policypkgtsmtanzuvmwarecomv1.VMpolicy{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVMpolicies) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(vmpoliciesResource, listOpts)

	_, err := c.Fake.Invokes(action, &policypkgtsmtanzuvmwarecomv1.VMpolicyList{})
	return err
}

// Patch applies the patch and returns the patched vMpolicy.
func (c *FakeVMpolicies) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *policypkgtsmtanzuvmwarecomv1.VMpolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(vmpoliciesResource, name, pt, data, subresources...), &policypkgtsmtanzuvmwarecomv1.VMpolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*policypkgtsmtanzuvmwarecomv1.VMpolicy), err
}
