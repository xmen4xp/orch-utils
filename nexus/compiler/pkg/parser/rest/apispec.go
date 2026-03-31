// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"go/ast"
	"go/types"
	"regexp"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/pkg/config"
	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/pkg/parser"
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
)

var uris = map[string]string{}

func GetRestApiSpecs(p parser.Package, httpMethods map[string]nexus.HTTPMethodsResponses,
	httpCodes map[string]nexus.HTTPCodesResponse, parentsMap map[string]parser.NodeHelper) map[string]nexus.RestAPISpec {

	apiSpecs := make(map[string]nexus.RestAPISpec)
	for _, spec := range parser.GetNexusSpecs(p, "nexus.RestAPISpec") {
		apiSpec := nexus.RestAPISpec{}
		for _, elt := range spec.Value.Elts {
			uris := elt.(*ast.KeyValueExpr)

			for _, uri := range uris.Value.(*ast.CompositeLit).Elts {
				restUri := extractApiSpecRestURI(uri.(*ast.CompositeLit), httpMethods, httpCodes)
				apiSpec.Uris = append(apiSpec.Uris, restUri)
			}
		}

		apiSpecs[spec.Name] = apiSpec
	}

	return apiSpecs
}

func extractApiSpecRestURI(uri *ast.CompositeLit, httpMethods map[string]nexus.HTTPMethodsResponses, httpCodes map[string]nexus.HTTPCodesResponse) nexus.RestURIs {
	restUri := nexus.RestURIs{}
	for _, elt := range uri.Elts {
		kv := elt.(*ast.KeyValueExpr)

		switch types.ExprString(kv.Key) {
		case "Uri":
			key, err := strconv.Unquote(types.ExprString(kv.Value))
			if err != nil {
				log.Errorf("Error %v", err)
			}
			restUri.Uri = key
		case "PathParams":
			restUri.PathParams = extractApiSpecMapParams(kv)
		case "QueryParams":
			restUri.QueryParams = extractApiSpecMapParams(kv)
		case "HeaderParams":
			restUri.HeaderParams = extractApiSpecMapParams(kv)
		case "Methods":
			restUri.Methods = extractApiSpecMethods(kv, httpMethods, httpCodes)
		}
	}

	return restUri
}

func extractApiSpecMethods(methods *ast.KeyValueExpr, httpMethods map[string]nexus.HTTPMethodsResponses, httpCodes map[string]nexus.HTTPCodesResponse) nexus.HTTPMethodsResponses {
	switch val := methods.Value.(type) {
	case *ast.Ident:
		return httpMethods[val.Name]
	case *ast.SelectorExpr:
		return httpMethods[val.Sel.String()]
	case *ast.CompositeLit:
		met := make(nexus.HTTPMethodsResponses)
		for _, elt := range val.Elts {
			kv := elt.(*ast.KeyValueExpr)
			httpKey := extractHttpMethodsKey(kv.Key)
			httpValue := extractHttpMethodsValue(kv.Value, httpCodes)
			met[httpKey] = httpValue
		}
		return met
	}
	return nil
}

func extractApiSpecMapParams(kv *ast.KeyValueExpr) map[string]string {
	params := make(map[string]string)
	switch val := kv.Value.(type) {
	case *ast.CompositeLit:
		for _, v := range val.Elts {
			kvExpr := v.(*ast.KeyValueExpr)
			key, err := strconv.Unquote(types.ExprString(kvExpr.Key))
			if err != nil {
				log.Errorf("Error parsing key: %v", err)
				continue
			}
			value, err := strconv.Unquote(types.ExprString(kvExpr.Value))
			if err != nil {
				log.Errorf("Error parsing value: %v", err)
				continue
			}
			params[key] = value
		}
	}
	return params
}

func ValidateRestApiSpec(apiSpec nexus.RestAPISpec, parentsMap map[string]parser.NodeHelper, crdName string) {
	crdHelper := parentsMap[crdName]

	ignoredParentPathParams := make(map[string]struct{})
	for _, val := range config.ConfigInstance.IgnoredParentPathParams {
		ignoredParentPathParams[val] = struct{}{}
	}

	for _, uri := range apiSpec.Uris {
		uriRegex, _ := regexp.Compile("{.*?}")
		redactedUri := uriRegex.ReplaceAllString(uri.Uri, "{param}")

		if u, ok := uris[redactedUri]; ok {
			log.Fatalf("RestApiSpec: Duplicate found: %s and %s", u, uri.Uri)
		}

		if _, ok := uri.Methods["LIST"]; ok {
			if pathParamExist(crdHelper.RestName, uri.PathParams) || headerExist(crdHelper.RestName, uri.HeaderParams) || queryParamExist(crdHelper.RestName, uri.QueryParams) {
				log.Fatalf("RestApiSpec: Provided node name (%s) cannot be applied as a param because endpoint is a list. URI: %s", crdHelper.RestName, uri.Uri)
			}
		}

		// Check if node name is in multiple locations (URI, Header, Query param)
		// Parent info must exist in exactly ONE location
		inURI := pathParamExist(crdHelper.RestName, uri.PathParams)
		inHeader := headerExist(crdHelper.RestName, uri.HeaderParams)
		inQuery := queryParamExist(crdHelper.RestName, uri.QueryParams)
		locations := 0
		if inURI {
			locations++
		}
		if inHeader {
			locations++
		}
		if inQuery {
			locations++
		}
		if locations > 1 {
			log.Fatalf("RestApiSpec: Provided node name (%s) cannot be applied to multiple locations (URI/Header/Query). Found in %d locations. URI: %s", crdHelper.RestName, locations, uri.Uri)
		}

		for _, parentCrd := range crdHelper.Parents {
			parentCrdHelper := parentsMap[parentCrd]
			parentName := parentCrdHelper.RestName

			if parentCrdHelper.IsSingleton {
				continue
			}

			// Check if parent is in multiple locations
			parentInURI := pathParamExist(parentName, uri.PathParams)
			parentInHeader := headerExist(parentName, uri.HeaderParams)
			parentInQuery := queryParamExist(parentName, uri.QueryParams)
			parentLocations := 0
			if parentInURI {
				parentLocations++
			}
			if parentInHeader {
				parentLocations++
			}
			if parentInQuery {
				parentLocations++
			}

			if parentLocations > 1 {
				log.Fatalf("RestApiSpec: Provided parent name (%s) cannot be applied to multiple locations (URI/Header/Query). Found in %d locations. URI: %s", parentName, parentLocations, uri.Uri)
			}

			if parentLocations == 0 {
				if _, exists := ignoredParentPathParams[parentName]; exists {
					log.Warnf("RestApiSpec: Provided parent name (%s) not found for uri: %s. Ignoring and proceeding, as it is configured as ignored parent path param", parentName, uri.Uri)
				} else {
					log.Fatalf("RestApiSpec: Provided parent name (%s) not found for uri: %s", parentName, uri.Uri)
				}
			}
		}

		uris[redactedUri] = uri.Uri
	}
}

func nodeExist(name string, params [][]string) bool {
	for _, p := range params {
		if p[1] == name {
			return true
		}
	}

	return false
}

func queryParamExist(name string, params map[string]string) bool {
	for _, nodeType := range params {
		if nodeType == name {
			return true
		}
	}

	return false
}

func headerExist(name string, params map[string]string) bool {
	for _, nodeType := range params {
		if nodeType == name {
			return true
		}
	}

	return false
}

func pathParamExist(name string, params map[string]string) bool {
	for _, nodeType := range params {
		if nodeType == name {
			return true
		}
	}

	return false
}
