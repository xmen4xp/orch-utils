// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"fmt"
	"sync"

	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"

	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const constDefaultChanSize = 100

// NexusCRDRestURIs bundles CRD type with its REST URIs for route registration.
type NexusCRDRestURIs struct {
	CRDType  string           // e.g., "datacenterses.datacenters.hd.cisco.com"
	RestURIs []nexus.RestURIs // REST URI specs from the CRD annotation
}

var (
	appName = "nexus-api-gw-model"
	log     = logging.GetLogger(appName)

	RestURIChan = make(chan NexusCRDRestURIs, constDefaultChanSize)
	CrdTypeChan = make(chan string, constDefaultChanSize)

	CrdTypeToRestUris      = make(map[string][]nexus.RestURIs)
	crdTypeToRestUrisMutex = &sync.Mutex{}

	// CRD name to CRD type (Gns.gns => gns.vmware.org).
	URIToCRDType      = make(map[string]string)
	uriToCRDTypeMutex = &sync.Mutex{}

	// URI to info about this URI.
	URIToURIInfo      = make(map[string]RestURIInfo)
	URIToURIInfoMutex = &sync.Mutex{}

	// CRD Type to NodeInfo (gns.vmware.org => NodeInfo{}).
	CrdTypeToNodeInfo      = make(map[string]NodeInfo)
	crdTypeToNodeInfoMutex = &sync.RWMutex{}

	// CRD Type to k8s spec (gns.vmware.org => CustomResourceDefinitionSpec).
	CrdTypeToSpec      = make(map[string]apiextensionsv1.CustomResourceDefinitionSpec)
	crdTypeToSpecMutex = &sync.Mutex{}

	DatamodelsChan                = make(chan string, constDefaultChanSize)
	DatamodelToDatamodelInfo      = make(map[string]DatamodelInfo)
	DatamodelToDatamodelInfoMutex = &sync.Mutex{}
)

func ConstructDatamodel(eventType EventType, name string, unstructuredObj *unstructured.Unstructured) {
	DatamodelToDatamodelInfoMutex.Lock()
	if eventType == Delete {
		delete(DatamodelToDatamodelInfo, name)
		DatamodelToDatamodelInfoMutex.Unlock()
		return
	}
	obj := unstructuredObj.Object

	spec, ok := obj["spec"].(map[string]interface{})
	if !ok {
		DatamodelToDatamodelInfoMutex.Unlock()
		fmt.Println("obj[spec] is not of type (map[string]interface{})")
		return
	}

	log.Debug().Msgf("ConstructDatamodel: Spec: %#v", spec)
	title, ok := spec["title"]
	if !ok {
		DatamodelToDatamodelInfoMutex.Unlock()
		return
	}
	titleInString, ok := title.(string)
	if !ok {
		DatamodelToDatamodelInfoMutex.Unlock()
		fmt.Println("title is not of type string")
		return
	}
	datamodelName := name
	DatamodelToDatamodelInfo[datamodelName] = DatamodelInfo{
		Title: titleInString,
	}
	log.Debug().Msgf("ConstructDatamodel: Datamodel info: %v", DatamodelToDatamodelInfo)
	DatamodelToDatamodelInfoMutex.Unlock()

	// Push to chan - MUST be outside mutex to avoid deadlock
	DatamodelsChan <- datamodelName
}

func ConstructMapURIToCRDType(eventType EventType, crdType string, apiURIs []nexus.RestURIs) {
	uriToCRDTypeMutex.Lock()
	defer uriToCRDTypeMutex.Unlock()

	if eventType == Delete {
		for uri, cType := range URIToCRDType {
			if cType == crdType {
				delete(URIToCRDType, uri)
			}
		}
	}

	for _, u := range apiURIs {
		URIToCRDType[u.Uri] = crdType
	}
}

func ConstructMapCRDTypeToNode(eventType EventType, crdType, name string, parentHierarchy []string,
	children, links map[string]NodeHelperChild, isSingleton bool, description string, deferredDelete bool,
) {
	crdTypeToNodeInfoMutex.Lock()
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
	crdTypeToNodeInfoMutex.Unlock()

	// Push new CRD Type to chan - MUST be outside mutex to avoid deadlock
	CrdTypeChan <- crdType
}

func GetCRDTypeToNodeInfo(crdType string) (NodeInfo, bool) {
	crdTypeToNodeInfoMutex.RLock()
	defer crdTypeToNodeInfoMutex.RUnlock()

	info, ok := CrdTypeToNodeInfo[crdType]
	return info, ok
}

func GetDatamodel(name string) (DatamodelInfo, bool) {
	DatamodelToDatamodelInfoMutex.Lock()
	defer DatamodelToDatamodelInfoMutex.Unlock()

	info, ok := DatamodelToDatamodelInfo[name]
	return info, ok
}

func ConstructMapCRDTypeToSpec(eventType EventType, crdType string, spec apiextensionsv1.CustomResourceDefinitionSpec) {
	crdTypeToSpecMutex.Lock()
	defer crdTypeToSpecMutex.Unlock()

	if eventType == Delete {
		delete(CrdTypeToSpec, crdType)
	}
	log.Debug().Msgf("Constructing map CRD type to spec for %s %#v", crdType, spec)
	CrdTypeToSpec[crdType] = spec
}

func GetRestUris(crdType string) ([]nexus.RestURIs, bool) {
	crdTypeToRestUrisMutex.Lock()
	defer crdTypeToRestUrisMutex.Unlock()

	uris, ok := CrdTypeToRestUris[crdType]
	return uris, ok
}

func ConstructMapCRDTypeToRestUris(eventType EventType, crdType string, restSpec nexus.RestAPISpec) {
	crdTypeToRestUrisMutex.Lock()
	if eventType == Delete {
		delete(CrdTypeToRestUris, crdType)
		crdTypeToRestUrisMutex.Unlock()
		return
	}

	log.Debug().Msgf("Constructing map CRD type to rest uris for %s %v", crdType, restSpec.Uris)
	CrdTypeToRestUris[crdType] = restSpec.Uris
	crdTypeToRestUrisMutex.Unlock()

	// Push new uris to chan - MUST be outside mutex to avoid deadlock
	RestURIChan <- NexusCRDRestURIs{
		CRDType:  crdType,
		RestURIs: restSpec.Uris,
	}
}

func ConstructMapURIToURIInfo(eventType EventType, m map[string]RestURIInfo) {
	URIToURIInfoMutex.Lock()
	defer URIToURIInfoMutex.Unlock()

	if eventType == Delete {
		for k := range m {
			delete(URIToURIInfo, k)
		}
	}
	for k, v := range m {
		URIToURIInfo[k] = v
	}
}

func GetURIInfo(uriPath string) (RestURIInfo, bool) {
	URIToURIInfoMutex.Lock()
	defer URIToURIInfoMutex.Unlock()
	info, ok := URIToURIInfo[uriPath]
	return info, ok
}

func SetURIToCRDType(uri, crdType string) {
	uriToCRDTypeMutex.Lock()
	defer uriToCRDTypeMutex.Unlock()
	URIToCRDType[uri] = crdType
}

func GetURIToCRDType(uri string) (string, bool) {
	uriToCRDTypeMutex.Lock()
	defer uriToCRDTypeMutex.Unlock()
	crdType, ok := URIToCRDType[uri]
	return crdType, ok
}

func GetCrdTypeToSpec(crdType string) (apiextensionsv1.CustomResourceDefinitionSpec, bool) {
	crdTypeToSpecMutex.Lock()
	defer crdTypeToSpecMutex.Unlock()
	spec, ok := CrdTypeToSpec[crdType]
	return spec, ok
}

func GetAllCrdTypeToNodeInfo() map[string]NodeInfo {
	crdTypeToNodeInfoMutex.RLock()
	defer crdTypeToNodeInfoMutex.RUnlock()
	result := make(map[string]NodeInfo, len(CrdTypeToNodeInfo))
	for k, v := range CrdTypeToNodeInfo {
		result[k] = v
	}
	return result
}

func GetAllCrdTypeToRestUris() map[string][]nexus.RestURIs {
	crdTypeToRestUrisMutex.Lock()
	defer crdTypeToRestUrisMutex.Unlock()
	result := make(map[string][]nexus.RestURIs, len(CrdTypeToRestUris))
	for k, v := range CrdTypeToRestUris {
		result[k] = v
	}
	return result
}
