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
	roottsmtanzuvmwarecomv1 "github.com/vmware-tanzu/graph-framework-for-microservices/compiler/example/output/generated/apis/root.tsm.tanzu.vmware.com/v1"
	versioned "github.com/vmware-tanzu/graph-framework-for-microservices/compiler/example/output/generated/client/clientset/versioned"
	internalinterfaces "github.com/vmware-tanzu/graph-framework-for-microservices/compiler/example/output/generated/client/informers/externalversions/internalinterfaces"
	v1 "github.com/vmware-tanzu/graph-framework-for-microservices/compiler/example/output/generated/client/listers/root.tsm.tanzu.vmware.com/v1"
	time "time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// RootInformer provides access to a shared informer and lister for
// Roots.
type RootInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.RootLister
}

type rootInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewRootInformer constructs a new informer for Root type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewRootInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredRootInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredRootInformer constructs a new informer for Root type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredRootInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.RootTsmV1().Roots().List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.RootTsmV1().Roots().Watch(context.TODO(), options)
			},
		},
		&roottsmtanzuvmwarecomv1.Root{},
		resyncPeriod,
		indexers,
	)
}

func (f *rootInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredRootInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *rootInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&roottsmtanzuvmwarecomv1.Root{}, f.defaultInformer)
}

func (f *rootInformer) Lister() v1.RootLister {
	return v1.NewRootLister(f.Informer().GetIndexer())
}
