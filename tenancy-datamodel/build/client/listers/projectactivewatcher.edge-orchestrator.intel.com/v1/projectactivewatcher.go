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

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/projectactivewatcher.edge-orchestrator.intel.com/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ProjectActiveWatcherLister helps list ProjectActiveWatchers.
// All objects returned here must be treated as read-only.
type ProjectActiveWatcherLister interface {
	// List lists all ProjectActiveWatchers in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.ProjectActiveWatcher, err error)
	// Get retrieves the ProjectActiveWatcher from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.ProjectActiveWatcher, error)
	ProjectActiveWatcherListerExpansion
}

// projectActiveWatcherLister implements the ProjectActiveWatcherLister interface.
type projectActiveWatcherLister struct {
	indexer cache.Indexer
}

// NewProjectActiveWatcherLister returns a new ProjectActiveWatcherLister.
func NewProjectActiveWatcherLister(indexer cache.Indexer) ProjectActiveWatcherLister {
	return &projectActiveWatcherLister{indexer: indexer}
}

// List lists all ProjectActiveWatchers in the indexer.
func (s *projectActiveWatcherLister) List(selector labels.Selector) (ret []*v1.ProjectActiveWatcher, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.ProjectActiveWatcher))
	})
	return ret, err
}

// Get retrieves the ProjectActiveWatcher from the index for a given name.
func (s *projectActiveWatcherLister) Get(name string) (*v1.ProjectActiveWatcher, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("projectactivewatcher"), name)
	}
	return obj.(*v1.ProjectActiveWatcher), nil
}
