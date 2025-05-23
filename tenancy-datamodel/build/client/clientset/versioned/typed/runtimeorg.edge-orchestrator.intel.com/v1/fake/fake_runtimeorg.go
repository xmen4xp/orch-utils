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
	runtimeorgedgeorchestratorintelcomv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/runtimeorg.edge-orchestrator.intel.com/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeRuntimeOrgs implements RuntimeOrgInterface
type FakeRuntimeOrgs struct {
	Fake *FakeRuntimeorgEdgeV1
}

var runtimeorgsResource = schema.GroupVersionResource{Group: "runtimeorg.edge-orchestrator.intel.com", Version: "v1", Resource: "runtimeorgs"}

var runtimeorgsKind = schema.GroupVersionKind{Group: "runtimeorg.edge-orchestrator.intel.com", Version: "v1", Kind: "RuntimeOrg"}

// Get takes name of the runtimeOrg, and returns the corresponding runtimeOrg object, and an error if there is any.
func (c *FakeRuntimeOrgs) Get(ctx context.Context, name string, options v1.GetOptions) (result *runtimeorgedgeorchestratorintelcomv1.RuntimeOrg, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(runtimeorgsResource, name), &runtimeorgedgeorchestratorintelcomv1.RuntimeOrg{})
	if obj == nil {
		return nil, err
	}
	return obj.(*runtimeorgedgeorchestratorintelcomv1.RuntimeOrg), err
}

// List takes label and field selectors, and returns the list of RuntimeOrgs that match those selectors.
func (c *FakeRuntimeOrgs) List(ctx context.Context, opts v1.ListOptions) (result *runtimeorgedgeorchestratorintelcomv1.RuntimeOrgList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(runtimeorgsResource, runtimeorgsKind, opts), &runtimeorgedgeorchestratorintelcomv1.RuntimeOrgList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &runtimeorgedgeorchestratorintelcomv1.RuntimeOrgList{ListMeta: obj.(*runtimeorgedgeorchestratorintelcomv1.RuntimeOrgList).ListMeta}
	for _, item := range obj.(*runtimeorgedgeorchestratorintelcomv1.RuntimeOrgList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested runtimeOrgs.
func (c *FakeRuntimeOrgs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(runtimeorgsResource, opts))
}

// Create takes the representation of a runtimeOrg and creates it.  Returns the server's representation of the runtimeOrg, and an error, if there is any.
func (c *FakeRuntimeOrgs) Create(ctx context.Context, runtimeOrg *runtimeorgedgeorchestratorintelcomv1.RuntimeOrg, opts v1.CreateOptions) (result *runtimeorgedgeorchestratorintelcomv1.RuntimeOrg, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(runtimeorgsResource, runtimeOrg), &runtimeorgedgeorchestratorintelcomv1.RuntimeOrg{})
	if obj == nil {
		return nil, err
	}
	return obj.(*runtimeorgedgeorchestratorintelcomv1.RuntimeOrg), err
}

// Update takes the representation of a runtimeOrg and updates it. Returns the server's representation of the runtimeOrg, and an error, if there is any.
func (c *FakeRuntimeOrgs) Update(ctx context.Context, runtimeOrg *runtimeorgedgeorchestratorintelcomv1.RuntimeOrg, opts v1.UpdateOptions) (result *runtimeorgedgeorchestratorintelcomv1.RuntimeOrg, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(runtimeorgsResource, runtimeOrg), &runtimeorgedgeorchestratorintelcomv1.RuntimeOrg{})
	if obj == nil {
		return nil, err
	}
	return obj.(*runtimeorgedgeorchestratorintelcomv1.RuntimeOrg), err
}

// Delete takes name of the runtimeOrg and deletes it. Returns an error if one occurs.
func (c *FakeRuntimeOrgs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(runtimeorgsResource, name, opts), &runtimeorgedgeorchestratorintelcomv1.RuntimeOrg{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeRuntimeOrgs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(runtimeorgsResource, listOpts)

	_, err := c.Fake.Invokes(action, &runtimeorgedgeorchestratorintelcomv1.RuntimeOrgList{})
	return err
}

// Patch applies the patch and returns the patched runtimeOrg.
func (c *FakeRuntimeOrgs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *runtimeorgedgeorchestratorintelcomv1.RuntimeOrg, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(runtimeorgsResource, name, pt, data, subresources...), &runtimeorgedgeorchestratorintelcomv1.RuntimeOrg{})
	if obj == nil {
		return nil, err
	}
	return obj.(*runtimeorgedgeorchestratorintelcomv1.RuntimeOrg), err
}
