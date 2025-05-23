// Copyright The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"
	v1 "/build/apis/gns.tsm.tanzu.vmware.com/v1"
	scheme "/build/client/clientset/versioned/scheme"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// BarChildsGetter has a method to return a BarChildInterface.
// A group's client should implement this interface.
type BarChildsGetter interface {
	BarChilds() BarChildInterface
}

// BarChildInterface has methods to work with BarChild resources.
type BarChildInterface interface {
	Create(ctx context.Context, barChild *v1.BarChild, opts metav1.CreateOptions) (*v1.BarChild, error)
	Update(ctx context.Context, barChild *v1.BarChild, opts metav1.UpdateOptions) (*v1.BarChild, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.BarChild, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.BarChildList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.BarChild, err error)
	BarChildExpansion
}

// barChilds implements BarChildInterface
type barChilds struct {
	client rest.Interface
}

// newBarChilds returns a BarChilds
func newBarChilds(c *GnsTsmV1Client) *barChilds {
	return &barChilds{
		client: c.RESTClient(),
	}
}

// Get takes name of the barChild, and returns the corresponding barChild object, and an error if there is any.
func (c *barChilds) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.BarChild, err error) {
	result = &v1.BarChild{}
	err = c.client.Get().
		Resource("barchilds").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of BarChilds that match those selectors.
func (c *barChilds) List(ctx context.Context, opts metav1.ListOptions) (result *v1.BarChildList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.BarChildList{}
	err = c.client.Get().
		Resource("barchilds").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested barChilds.
func (c *barChilds) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("barchilds").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a barChild and creates it.  Returns the server's representation of the barChild, and an error, if there is any.
func (c *barChilds) Create(ctx context.Context, barChild *v1.BarChild, opts metav1.CreateOptions) (result *v1.BarChild, err error) {
	result = &v1.BarChild{}
	err = c.client.Post().
		Resource("barchilds").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(barChild).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a barChild and updates it. Returns the server's representation of the barChild, and an error, if there is any.
func (c *barChilds) Update(ctx context.Context, barChild *v1.BarChild, opts metav1.UpdateOptions) (result *v1.BarChild, err error) {
	result = &v1.BarChild{}
	err = c.client.Put().
		Resource("barchilds").
		Name(barChild.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(barChild).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the barChild and deletes it. Returns an error if one occurs.
func (c *barChilds) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Resource("barchilds").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *barChilds) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("barchilds").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched barChild.
func (c *barChilds) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.BarChild, err error) {
	result = &v1.BarChild{}
	err = c.client.Patch(pt).
		Resource("barchilds").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
