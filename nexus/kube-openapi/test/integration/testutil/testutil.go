// Copyright 2021 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"fmt"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/common"
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/util"
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/validation/spec"
)

// CreateOpenAPIBuilderConfig hard-codes some values in the API builder
// config for testing.
func CreateOpenAPIBuilderConfig() *common.Config {
	return &common.Config{
		ProtocolList:   []string{"https"},
		IgnorePrefixes: []string{"/swaggerapi"},
		Info: &spec.Info{
			InfoProps: spec.InfoProps{
				Title:   "Integration Test",
				Version: "1.0",
			},
		},
		ResponseDefinitions: map[string]spec.Response{
			"NotFound": spec.Response{
				ResponseProps: spec.ResponseProps{
					Description: "Entity not found.",
				},
			},
		},
		CommonResponses: map[int]spec.Response{
			404: *spec.ResponseRef("#/responses/NotFound"),
		},
	}
}

// CreateWebServices hard-codes a simple WebService which only defines a GET and POST paths
// for testing.
func CreateWebServices(includeV2SchemaAnnotation bool) []*restful.WebService {
	w := new(restful.WebService)
	addRoutes(w, buildRouteForType(w, "dummytype", "Foo")...)
	addRoutes(w, buildRouteForType(w, "dummytype", "Bar")...)
	addRoutes(w, buildRouteForType(w, "dummytype", "Baz")...)
	addRoutes(w, buildRouteForType(w, "dummytype", "Waldo")...)
	addRoutes(w, buildRouteForType(w, "listtype", "AtomicList")...)
	addRoutes(w, buildRouteForType(w, "listtype", "MapList")...)
	addRoutes(w, buildRouteForType(w, "listtype", "SetList")...)
	addRoutes(w, buildRouteForType(w, "uniontype", "TopLevelUnion")...)
	addRoutes(w, buildRouteForType(w, "uniontype", "InlinedUnion")...)
	addRoutes(w, buildRouteForType(w, "custom", "Bal")...)
	addRoutes(w, buildRouteForType(w, "custom", "Bak")...)
	if includeV2SchemaAnnotation {
		addRoutes(w, buildRouteForType(w, "custom", "Bac")...)
		addRoutes(w, buildRouteForType(w, "custom", "Bah")...)
	}
	addRoutes(w, buildRouteForType(w, "maptype", "GranularMap")...)
	addRoutes(w, buildRouteForType(w, "maptype", "AtomicMap")...)
	addRoutes(w, buildRouteForType(w, "structtype", "GranularStruct")...)
	addRoutes(w, buildRouteForType(w, "structtype", "AtomicStruct")...)
	addRoutes(w, buildRouteForType(w, "structtype", "DeclaredAtomicStruct")...)
	addRoutes(w, buildRouteForType(w, "defaults", "Defaulted")...)
	return []*restful.WebService{w}
}

func addRoutes(ws *restful.WebService, routes ...*restful.RouteBuilder) {
	for _, r := range routes {
		ws.Route(r)
	}
}

// Implements OpenAPICanonicalTypeNamer
var _ = util.OpenAPICanonicalTypeNamer(&typeNamer{})

type typeNamer struct {
	pkg  string
	name string
}

func (t *typeNamer) OpenAPICanonicalTypeName() string {
	return fmt.Sprintf("github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/test/integration/testdata/%s.%s", t.pkg, t.name)
}

func buildRouteForType(ws *restful.WebService, pkg, name string) []*restful.RouteBuilder {
	namer := typeNamer{
		pkg:  pkg,
		name: name,
	}

	routes := []*restful.RouteBuilder{
		ws.GET(fmt.Sprintf("test/%s/%s", pkg, strings.ToLower(name))).
			Operation(fmt.Sprintf("get-%s.%s", pkg, name)).
			Produces("application/json").
			To(func(*restful.Request, *restful.Response) {}).
			Writes(&namer),
		ws.POST(fmt.Sprintf("test/%s", pkg)).
			Operation(fmt.Sprintf("create-%s.%s", pkg, name)).
			Produces("application/json").
			To(func(*restful.Request, *restful.Response) {}).
			Returns(201, "Created", &namer).
			Writes(&namer),
	}

	if pkg == "dummytype" {
		statusErrType := typeNamer{
			pkg:  "dummytype",
			name: "StatusError",
		}

		for _, route := range routes {
			route.Returns(500, "Internal Service Error", &statusErrType)
		}
	}

	return routes
}
