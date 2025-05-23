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
	runtimefolderedgeorchestratorintelcomv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/runtimefolder.edge-orchestrator.intel.com/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeRuntimeFolders implements RuntimeFolderInterface
type FakeRuntimeFolders struct {
	Fake *FakeRuntimefolderEdgeV1
}

var runtimefoldersResource = schema.GroupVersionResource{Group: "runtimefolder.edge-orchestrator.intel.com", Version: "v1", Resource: "runtimefolders"}

var runtimefoldersKind = schema.GroupVersionKind{Group: "runtimefolder.edge-orchestrator.intel.com", Version: "v1", Kind: "RuntimeFolder"}

// Get takes name of the runtimeFolder, and returns the corresponding runtimeFolder object, and an error if there is any.
func (c *FakeRuntimeFolders) Get(ctx context.Context, name string, options v1.GetOptions) (result *runtimefolderedgeorchestratorintelcomv1.RuntimeFolder, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(runtimefoldersResource, name), &runtimefolderedgeorchestratorintelcomv1.RuntimeFolder{})
	if obj == nil {
		return nil, err
	}
	return obj.(*runtimefolderedgeorchestratorintelcomv1.RuntimeFolder), err
}

// List takes label and field selectors, and returns the list of RuntimeFolders that match those selectors.
func (c *FakeRuntimeFolders) List(ctx context.Context, opts v1.ListOptions) (result *runtimefolderedgeorchestratorintelcomv1.RuntimeFolderList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(runtimefoldersResource, runtimefoldersKind, opts), &runtimefolderedgeorchestratorintelcomv1.RuntimeFolderList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &runtimefolderedgeorchestratorintelcomv1.RuntimeFolderList{ListMeta: obj.(*runtimefolderedgeorchestratorintelcomv1.RuntimeFolderList).ListMeta}
	for _, item := range obj.(*runtimefolderedgeorchestratorintelcomv1.RuntimeFolderList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested runtimeFolders.
func (c *FakeRuntimeFolders) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(runtimefoldersResource, opts))
}

// Create takes the representation of a runtimeFolder and creates it.  Returns the server's representation of the runtimeFolder, and an error, if there is any.
func (c *FakeRuntimeFolders) Create(ctx context.Context, runtimeFolder *runtimefolderedgeorchestratorintelcomv1.RuntimeFolder, opts v1.CreateOptions) (result *runtimefolderedgeorchestratorintelcomv1.RuntimeFolder, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(runtimefoldersResource, runtimeFolder), &runtimefolderedgeorchestratorintelcomv1.RuntimeFolder{})
	if obj == nil {
		return nil, err
	}
	return obj.(*runtimefolderedgeorchestratorintelcomv1.RuntimeFolder), err
}

// Update takes the representation of a runtimeFolder and updates it. Returns the server's representation of the runtimeFolder, and an error, if there is any.
func (c *FakeRuntimeFolders) Update(ctx context.Context, runtimeFolder *runtimefolderedgeorchestratorintelcomv1.RuntimeFolder, opts v1.UpdateOptions) (result *runtimefolderedgeorchestratorintelcomv1.RuntimeFolder, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(runtimefoldersResource, runtimeFolder), &runtimefolderedgeorchestratorintelcomv1.RuntimeFolder{})
	if obj == nil {
		return nil, err
	}
	return obj.(*runtimefolderedgeorchestratorintelcomv1.RuntimeFolder), err
}

// Delete takes name of the runtimeFolder and deletes it. Returns an error if one occurs.
func (c *FakeRuntimeFolders) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(runtimefoldersResource, name, opts), &runtimefolderedgeorchestratorintelcomv1.RuntimeFolder{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeRuntimeFolders) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(runtimefoldersResource, listOpts)

	_, err := c.Fake.Invokes(action, &runtimefolderedgeorchestratorintelcomv1.RuntimeFolderList{})
	return err
}

// Patch applies the patch and returns the patched runtimeFolder.
func (c *FakeRuntimeFolders) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *runtimefolderedgeorchestratorintelcomv1.RuntimeFolder, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(runtimefoldersResource, name, pt, data, subresources...), &runtimefolderedgeorchestratorintelcomv1.RuntimeFolder{})
	if obj == nil {
		return nil, err
	}
	return obj.(*runtimefolderedgeorchestratorintelcomv1.RuntimeFolder), err
}
