// Copyright The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "/build/apis/config.tsm.tanzu.vmware.com/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// FooTypeABCLister helps list FooTypeABCs.
// All objects returned here must be treated as read-only.
type FooTypeABCLister interface {
	// List lists all FooTypeABCs in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.FooTypeABC, err error)
	// Get retrieves the FooTypeABC from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.FooTypeABC, error)
	FooTypeABCListerExpansion
}

// fooTypeABCLister implements the FooTypeABCLister interface.
type fooTypeABCLister struct {
	indexer cache.Indexer
}

// NewFooTypeABCLister returns a new FooTypeABCLister.
func NewFooTypeABCLister(indexer cache.Indexer) FooTypeABCLister {
	return &fooTypeABCLister{indexer: indexer}
}

// List lists all FooTypeABCs in the indexer.
func (s *fooTypeABCLister) List(selector labels.Selector) (ret []*v1.FooTypeABC, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.FooTypeABC))
	})
	return ret, err
}

// Get retrieves the FooTypeABC from the index for a given name.
func (s *fooTypeABCLister) Get(name string) (*v1.FooTypeABC, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("footypeabc"), name)
	}
	return obj.(*v1.FooTypeABC), nil
}
