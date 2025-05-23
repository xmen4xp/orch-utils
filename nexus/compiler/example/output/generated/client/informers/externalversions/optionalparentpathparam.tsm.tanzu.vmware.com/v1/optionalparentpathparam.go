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
	optionalparentpathparamtsmtanzuvmwarecomv1 "github.com/vmware-tanzu/graph-framework-for-microservices/compiler/example/output/generated/apis/optionalparentpathparam.tsm.tanzu.vmware.com/v1"
	versioned "github.com/vmware-tanzu/graph-framework-for-microservices/compiler/example/output/generated/client/clientset/versioned"
	internalinterfaces "github.com/vmware-tanzu/graph-framework-for-microservices/compiler/example/output/generated/client/informers/externalversions/internalinterfaces"
	v1 "github.com/vmware-tanzu/graph-framework-for-microservices/compiler/example/output/generated/client/listers/optionalparentpathparam.tsm.tanzu.vmware.com/v1"
	time "time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// OptionalParentPathParamInformer provides access to a shared informer and lister for
// OptionalParentPathParams.
type OptionalParentPathParamInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.OptionalParentPathParamLister
}

type optionalParentPathParamInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewOptionalParentPathParamInformer constructs a new informer for OptionalParentPathParam type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewOptionalParentPathParamInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredOptionalParentPathParamInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredOptionalParentPathParamInformer constructs a new informer for OptionalParentPathParam type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredOptionalParentPathParamInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.OptionalparentpathparamTsmV1().OptionalParentPathParams().List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.OptionalparentpathparamTsmV1().OptionalParentPathParams().Watch(context.TODO(), options)
			},
		},
		&optionalparentpathparamtsmtanzuvmwarecomv1.OptionalParentPathParam{},
		resyncPeriod,
		indexers,
	)
}

func (f *optionalParentPathParamInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredOptionalParentPathParamInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *optionalParentPathParamInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&optionalparentpathparamtsmtanzuvmwarecomv1.OptionalParentPathParam{}, f.defaultInformer)
}

func (f *optionalParentPathParamInformer) Lister() v1.OptionalParentPathParamLister {
	return v1.NewOptionalParentPathParamLister(f.Informer().GetIndexer())
}
