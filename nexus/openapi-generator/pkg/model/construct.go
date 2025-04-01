// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"sync"

	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var (
	RestURIChan = make(chan []nexus.RestURIs, 100)
	CrdTypeChan = make(chan string, 100)

	CrdTypeToRestUris      = make(map[string][]nexus.RestURIs)
	crdTypeToRestUrisMutex = &sync.Mutex{}

	// CRD name to CRD type (Gns.gns => gns.vmware.org)
	UriToCRDType      = make(map[string]string)
	uriToCRDTypeMutex = &sync.Mutex{}

	// URI to info about this URI
	UriToUriInfo      = make(map[string]RestURIInfo)
	UriToUriInfoMutex = &sync.Mutex{}

	// CRD Type to NodeInfo (gns.vmware.org => NodeInfo{})
	CrdTypeToNodeInfo      = make(map[string]NodeInfo)
	crdTypeToNodeInfoMutex = &sync.Mutex{}

	// CRD Type to k8s spec (gns.vmware.org => CustomResourceDefinitionSpec)
	CrdTypeToSpec      = make(map[string]apiextensionsv1.CustomResourceDefinitionSpec)
	crdTypeToSpecMutex = &sync.Mutex{}

	DatamodelsChan                = make(chan string, 100)
	DatamodelToDatamodelInfo      = make(map[string]DatamodelInfo)
	DatamodelToDatamodelInfoMutex = &sync.Mutex{}
)

func ConstructMapURIToCRDType(eventType EventType, crdType string, apiURIs []nexus.RestURIs) {
	uriToCRDTypeMutex.Lock()
	defer uriToCRDTypeMutex.Unlock()

	if eventType == Delete {
		for uri, cType := range UriToCRDType {
			if cType == crdType {
				delete(UriToCRDType, uri)
			}
		}
	}

	for _, u := range apiURIs {
		UriToCRDType[u.Uri] = crdType
	}
}

func ConstructMapCRDTypeToNode(eventType EventType, crdType, name string, parentHierarchy []string,
	children, links map[string]NodeHelperChild, isSingleton bool, description string, deferredDelete bool) {
	crdTypeToNodeInfoMutex.Lock()
	defer crdTypeToNodeInfoMutex.Unlock()

	if eventType == Delete {
		delete(CrdTypeToNodeInfo, crdType)
	}

	CrdTypeToNodeInfo[crdType] = NodeInfo{
		Name:            name,
		ParentHierarchy: parentHierarchy,
		Children:        children,
		Links:           links,
		IsSingleton:     isSingleton,
		Description:     description,
		DeferredDelete:  deferredDelete,
	}

	// Push new CRD Type to chan
	CrdTypeChan <- crdType
}

func ConstructMapCRDTypeToSpec(eventType EventType, crdType string, spec apiextensionsv1.CustomResourceDefinitionSpec) {
	crdTypeToSpecMutex.Lock()
	defer crdTypeToSpecMutex.Unlock()

	if eventType == Delete {
		delete(CrdTypeToSpec, crdType)
	}
	CrdTypeToSpec[crdType] = spec
}

func ConstructMapCRDTypeToRestUris(eventType EventType, crdType string, restSpec nexus.RestAPISpec) {
	crdTypeToRestUrisMutex.Lock()
	defer crdTypeToRestUrisMutex.Unlock()

	if eventType == Delete {
		delete(CrdTypeToRestUris, crdType)
		return
	}

	CrdTypeToRestUris[crdType] = restSpec.Uris

	// Push new uris to chan
	RestURIChan <- restSpec.Uris
}

func ConstructMapUriToUriInfo(eventType EventType, m map[string]RestURIInfo) {
	UriToUriInfoMutex.Lock()
	defer UriToUriInfoMutex.Unlock()

	if eventType == Delete {
		for k := range m {
			delete(UriToUriInfo, k)
		}
	}
	for k, v := range m {
		UriToUriInfo[k] = v
	}
}

func GetUriInfo(uriPath string) (RestURIInfo, bool) {
	UriToUriInfoMutex.Lock()
	defer UriToUriInfoMutex.Unlock()
	info, ok := UriToUriInfo[uriPath]
	return info, ok
}
