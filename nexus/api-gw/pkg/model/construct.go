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

var (
	appName = "nexus-api-gw-model"
	log     = logging.GetLogger(appName)

	RestURIChan = make(chan []nexus.RestURIs, constDefaultChanSize)
	CrdTypeChan = make(chan string, constDefaultChanSize)

	CrdTypeToRestUris      = make(map[string][]nexus.RestURIs)
	crdTypeToRestUrisMutex = &sync.RWMutex{}

	// CRD name to CRD type (Gns.gns => gns.vmware.org).
	URIToCRDType      = make(map[string]string)
	uriToCRDTypeMutex = &sync.RWMutex{}

	// URI to info about this URI.
	URIToURIInfo      = make(map[string]RestURIInfo)
	URIToURIInfoMutex = &sync.RWMutex{}

	// CRD Type to NodeInfo (gns.vmware.org => NodeInfo{}).
	CrdTypeToNodeInfo      = make(map[string]NodeInfo)
	crdTypeToNodeInfoMutex = &sync.RWMutex{}

	// CRD Type to k8s spec (gns.vmware.org => CustomResourceDefinitionSpec).
	CrdTypeToSpec      = make(map[string]apiextensionsv1.CustomResourceDefinitionSpec)
	crdTypeToSpecMutex = &sync.RWMutex{}

	DatamodelsChan                = make(chan string, constDefaultChanSize)
	DatamodelToDatamodelInfo      = make(map[string]DatamodelInfo)
	DatamodelToDatamodelInfoMutex = &sync.RWMutex{}

	// Cache for fast parameter name lookups: URI -> nodeType -> parameterName
	// These are built from PathParams, QueryParams, and HeaderParams maps
	URIToNodeTypeHeaderMap      = make(map[string]map[string]string)
	URIToNodeTypeHeaderMapMutex = &sync.RWMutex{}

	URIToNodeTypeQueryMap      = make(map[string]map[string]string)
	URIToNodeTypeQueryMapMutex = &sync.RWMutex{}

	URIToNodeTypePathMap      = make(map[string]map[string]string)
	URIToNodeTypePathMapMutex = &sync.RWMutex{}
)

func ConstructDatamodel(eventType EventType, name string, unstructuredObj *unstructured.Unstructured) {
	DatamodelToDatamodelInfoMutex.Lock()
	defer DatamodelToDatamodelInfoMutex.Unlock()

	if eventType == Delete {
		delete(DatamodelToDatamodelInfo, name)
		return
	}
	obj := unstructuredObj.Object

	spec, ok := obj["spec"].(map[string]interface{})
	if !ok {
		fmt.Println("obj[spec] is not of type (map[string]interface{})")
		return
	}

	log.Debug().Msgf("ConstructDatamodel: Spec: %#v", spec)
	if title, ok := spec["title"]; ok {
		titleInString, ok := title.(string)
		if !ok {
			fmt.Println("title is not of type string")
			return
		}
		datamodelName := name
		DatamodelToDatamodelInfo[datamodelName] = DatamodelInfo{
			Title: titleInString,
		}

		log.Debug().Msgf("ConstructDatamodel: Datamodel info: %v", DatamodelToDatamodelInfo)

		DatamodelsChan <- datamodelName
	}
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

	// Push new CRD Type to chan.
	CrdTypeChan <- crdType
}

func GetCRDTypeToNodeInfo(crdType string) (NodeInfo, bool) {
	crdTypeToNodeInfoMutex.RLock()
	defer crdTypeToNodeInfoMutex.RUnlock()

	info, ok := CrdTypeToNodeInfo[crdType]
	return info, ok
}

func GetDatamodel(name string) (DatamodelInfo, bool) {
	DatamodelToDatamodelInfoMutex.RLock()
	defer DatamodelToDatamodelInfoMutex.RUnlock()

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
	crdTypeToRestUrisMutex.RLock()
	defer crdTypeToRestUrisMutex.RUnlock()

	uris, ok := CrdTypeToRestUris[crdType]
	return uris, ok
}

func ConstructMapCRDTypeToRestUris(eventType EventType, crdType string, restSpec nexus.RestAPISpec) {
	crdTypeToRestUrisMutex.Lock()
	defer crdTypeToRestUrisMutex.Unlock()

	if eventType == Delete {
		delete(CrdTypeToRestUris, crdType)
		// Also clear parameter name caches for deleted URIs
		ClearParameterCachesForCRD(crdType)
		return
	}

	log.Debug().Msgf("Constructing map CRD type to rest uris for %s %v", crdType, restSpec.Uris)
	CrdTypeToRestUris[crdType] = restSpec.Uris

	// Build parameter name caches for fast lookups
	BuildParameterCaches(restSpec.Uris)

	// Push new uris to chan.
	RestURIChan <- restSpec.Uris
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
	URIToURIInfoMutex.RLock()
	defer URIToURIInfoMutex.RUnlock()
	info, ok := URIToURIInfo[uriPath]
	return info, ok
}

// BuildParameterCaches builds reverse lookup caches from RestURIs for O(1) parameter name lookups.
func BuildParameterCaches(uris []nexus.RestURIs) {
	for _, uri := range uris {
		// Build header cache: nodeType -> headerName
		if len(uri.HeaderParams) > 0 {
			URIToNodeTypeHeaderMapMutex.Lock()
			if URIToNodeTypeHeaderMap[uri.Uri] == nil {
				URIToNodeTypeHeaderMap[uri.Uri] = make(map[string]string)
			}
			for headerName, nodeType := range uri.HeaderParams {
				URIToNodeTypeHeaderMap[uri.Uri][nodeType] = headerName
			}
			URIToNodeTypeHeaderMapMutex.Unlock()
		}

		// Build query cache: nodeType -> queryParamName
		if len(uri.QueryParams) > 0 {
			URIToNodeTypeQueryMapMutex.Lock()
			if URIToNodeTypeQueryMap[uri.Uri] == nil {
				URIToNodeTypeQueryMap[uri.Uri] = make(map[string]string)
			}
			for queryName, nodeType := range uri.QueryParams {
				URIToNodeTypeQueryMap[uri.Uri][nodeType] = queryName
			}
			URIToNodeTypeQueryMapMutex.Unlock()
		}

		// Build path cache: nodeType -> pathParamName
		if len(uri.PathParams) > 0 {
			URIToNodeTypePathMapMutex.Lock()
			if URIToNodeTypePathMap[uri.Uri] == nil {
				URIToNodeTypePathMap[uri.Uri] = make(map[string]string)
			}
			for pathName, nodeType := range uri.PathParams {
				URIToNodeTypePathMap[uri.Uri][nodeType] = pathName
			}
			URIToNodeTypePathMapMutex.Unlock()
		}
	}
}

// ClearParameterCachesForCRD clears all parameter caches for URIs belonging to a CRD.
func ClearParameterCachesForCRD(crdType string) {
	uris, ok := CrdTypeToRestUris[crdType]
	if !ok {
		return
	}

	URIToNodeTypeHeaderMapMutex.Lock()
	for _, uri := range uris {
		delete(URIToNodeTypeHeaderMap, uri.Uri)
	}
	URIToNodeTypeHeaderMapMutex.Unlock()

	URIToNodeTypeQueryMapMutex.Lock()
	for _, uri := range uris {
		delete(URIToNodeTypeQueryMap, uri.Uri)
	}
	URIToNodeTypeQueryMapMutex.Unlock()

	URIToNodeTypePathMapMutex.Lock()
	for _, uri := range uris {
		delete(URIToNodeTypePathMap, uri.Uri)
	}
	URIToNodeTypePathMapMutex.Unlock()
}

// GetHeaderNameForNodeType returns the header parameter name for a given node type and URI.
func GetHeaderNameForNodeType(uri, nodeType string) (string, bool) {
	URIToNodeTypeHeaderMapMutex.RLock()
	defer URIToNodeTypeHeaderMapMutex.RUnlock()
	if m, ok := URIToNodeTypeHeaderMap[uri]; ok {
		headerName, found := m[nodeType]
		return headerName, found
	}
	return "", false
}

// GetQueryNameForNodeType returns the query parameter name for a given node type and URI.
func GetQueryNameForNodeType(uri, nodeType string) (string, bool) {
	URIToNodeTypeQueryMapMutex.RLock()
	defer URIToNodeTypeQueryMapMutex.RUnlock()
	if m, ok := URIToNodeTypeQueryMap[uri]; ok {
		queryName, found := m[nodeType]
		return queryName, found
	}
	return "", false
}

// GetPathNameForNodeType returns the path parameter name for a given node type and URI.
func GetPathNameForNodeType(uri, nodeType string) (string, bool) {
	URIToNodeTypePathMapMutex.RLock()
	defer URIToNodeTypePathMapMutex.RUnlock()
	if m, ok := URIToNodeTypePathMap[uri]; ok {
		pathName, found := m[nodeType]
		return pathName, found
	}
	return "", false
}
