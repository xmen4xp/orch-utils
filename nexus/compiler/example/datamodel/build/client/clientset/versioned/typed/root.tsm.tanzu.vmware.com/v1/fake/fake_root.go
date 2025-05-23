// Copyright The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"
	roottsmtanzuvmwarecomv1 "/build/apis/root.tsm.tanzu.vmware.com/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeRoots implements RootInterface
type FakeRoots struct {
	Fake *FakeRootTsmV1
}

var rootsResource = schema.GroupVersionResource{Group: "root.tsm.tanzu.vmware.com", Version: "v1", Resource: "roots"}

var rootsKind = schema.GroupVersionKind{Group: "root.tsm.tanzu.vmware.com", Version: "v1", Kind: "Root"}

// Get takes name of the root, and returns the corresponding root object, and an error if there is any.
func (c *FakeRoots) Get(ctx context.Context, name string, options v1.GetOptions) (result *roottsmtanzuvmwarecomv1.Root, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(rootsResource, name), &roottsmtanzuvmwarecomv1.Root{})
	if obj == nil {
		return nil, err
	}
	return obj.(*roottsmtanzuvmwarecomv1.Root), err
}

// List takes label and field selectors, and returns the list of Roots that match those selectors.
func (c *FakeRoots) List(ctx context.Context, opts v1.ListOptions) (result *roottsmtanzuvmwarecomv1.RootList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(rootsResource, rootsKind, opts), &roottsmtanzuvmwarecomv1.RootList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &roottsmtanzuvmwarecomv1.RootList{ListMeta: obj.(*roottsmtanzuvmwarecomv1.RootList).ListMeta}
	for _, item := range obj.(*roottsmtanzuvmwarecomv1.RootList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested roots.
func (c *FakeRoots) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(rootsResource, opts))
}

// Create takes the representation of a root and creates it.  Returns the server's representation of the root, and an error, if there is any.
func (c *FakeRoots) Create(ctx context.Context, root *roottsmtanzuvmwarecomv1.Root, opts v1.CreateOptions) (result *roottsmtanzuvmwarecomv1.Root, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(rootsResource, root), &roottsmtanzuvmwarecomv1.Root{})
	if obj == nil {
		return nil, err
	}
	return obj.(*roottsmtanzuvmwarecomv1.Root), err
}

// Update takes the representation of a root and updates it. Returns the server's representation of the root, and an error, if there is any.
func (c *FakeRoots) Update(ctx context.Context, root *roottsmtanzuvmwarecomv1.Root, opts v1.UpdateOptions) (result *roottsmtanzuvmwarecomv1.Root, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(rootsResource, root), &roottsmtanzuvmwarecomv1.Root{})
	if obj == nil {
		return nil, err
	}
	return obj.(*roottsmtanzuvmwarecomv1.Root), err
}

// Delete takes name of the root and deletes it. Returns an error if one occurs.
func (c *FakeRoots) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(rootsResource, name, opts), &roottsmtanzuvmwarecomv1.Root{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeRoots) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(rootsResource, listOpts)

	_, err := c.Fake.Invokes(action, &roottsmtanzuvmwarecomv1.RootList{})
	return err
}

// Patch applies the patch and returns the patched root.
func (c *FakeRoots) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *roottsmtanzuvmwarecomv1.Root, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(rootsResource, name, pt, data, subresources...), &roottsmtanzuvmwarecomv1.Root{})
	if obj == nil {
		return nil, err
	}
	return obj.(*roottsmtanzuvmwarecomv1.Root), err
}
