// Copyright The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	"context"
	optionalparentpathparamtsmtanzuvmwarecomv1 "/build/apis/optionalparentpathparam.tsm.tanzu.vmware.com/v1"
	versioned "/build/client/clientset/versioned"
	internalinterfaces "/build/client/informers/externalversions/internalinterfaces"
	v1 "/build/client/listers/optionalparentpathparam.tsm.tanzu.vmware.com/v1"
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
