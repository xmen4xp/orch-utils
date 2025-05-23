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

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	"context"
	runtimeprojectedgeorchestratorintelcomv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/runtimeproject.edge-orchestrator.intel.com/v1"
	versioned "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/client/clientset/versioned"
	internalinterfaces "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/client/informers/externalversions/internalinterfaces"
	v1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/client/listers/runtimeproject.edge-orchestrator.intel.com/v1"
	time "time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// RuntimeProjectInformer provides access to a shared informer and lister for
// RuntimeProjects.
type RuntimeProjectInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.RuntimeProjectLister
}

type runtimeProjectInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewRuntimeProjectInformer constructs a new informer for RuntimeProject type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewRuntimeProjectInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredRuntimeProjectInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredRuntimeProjectInformer constructs a new informer for RuntimeProject type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredRuntimeProjectInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.RuntimeprojectEdgeV1().RuntimeProjects().List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.RuntimeprojectEdgeV1().RuntimeProjects().Watch(context.TODO(), options)
			},
		},
		&runtimeprojectedgeorchestratorintelcomv1.RuntimeProject{},
		resyncPeriod,
		indexers,
	)
}

func (f *runtimeProjectInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredRuntimeProjectInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *runtimeProjectInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&runtimeprojectedgeorchestratorintelcomv1.RuntimeProject{}, f.defaultInformer)
}

func (f *runtimeProjectInformer) Lister() v1.RuntimeProjectLister {
	return v1.NewRuntimeProjectLister(f.Informer().GetIndexer())
}
