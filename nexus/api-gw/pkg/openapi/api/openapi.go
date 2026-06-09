// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"nexus-api-gw/pkg/config"
	"nexus-api-gw/pkg/model"
	"nexus-api-gw/pkg/utils"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
)

var (
	Schemas      = make(map[string]openapi3.T)
	schemasMutex = &sync.RWMutex{}
	appName      = "nexus-api-gw-openapi"
	log          = logging.GetLogger(appName)
)

func New(datamodel string) {
	// Check if datamodel info is present.
	title := "Nexus API GW APIs"
	if info, ok := model.DatamodelToDatamodelInfo[datamodel]; ok {
		title = info.Title
	}
	schema := openapi3.T{
		OpenAPI: "3.1.1",
		Info: &openapi3.Info{
			Title:          title,
			Description:    "",
			TermsOfService: "",
			Contact:        nil,
			License:        nil,
			Version:        "1.0.0",
		},
		Servers: openapi3.Servers{
			&openapi3.Server{
				Description: "API Gateway",
				URL:         "/",
			},
			&openapi3.Server{
				Description: "Local",
				URL:         "http://localhost:5000",
			},
			&openapi3.Server{
				Description: "Local SSL",
				URL:         "https://localhost:5443",
			},
		},
		Paths: openapi3.NewPaths(),
		Components: &openapi3.Components{
			Schemas:       openapi3.Schemas{},
			RequestBodies: openapi3.RequestBodies{},
			Responses: openapi3.ResponseBodies{
				"DefaultResponse": &openapi3.ResponseRef{
					Value: openapi3.NewResponse().
						WithDescription("Default response").
						WithContent(openapi3.NewContentWithJSONSchema(
							openapi3.NewSchema().WithProperty("message", openapi3.NewStringSchema())),
						),
				},
				"NotFoundResponse": &openapi3.ResponseRef{
					Value: openapi3.NewResponse().
						WithDescription("Not Found").
						WithContent(openapi3.NewContentWithJSONSchema(
							openapi3.NewSchema().WithProperty("message", openapi3.NewStringSchema())),
						),
				},
			},
		},
	}
	log.Info().Msgf("Created schema for %s", datamodel)
	schemasMutex.Lock()
	Schemas[datamodel] = schema
	schemasMutex.Unlock()
}

func DatamodelUpdateNotification() {
	log.Debug().Msg("Started datamodel update notification")
	for name := range model.DatamodelsChan {
		log.Debug().Msgf("Received datamodel update notification for %s", name)
		schemasMutex.RLock()
		_, exists := Schemas[name]
		schemasMutex.RUnlock()
		if !exists {
			New(name)
		}

		schemasMutex.Lock()
		if schema, ok := Schemas[name]; ok {
			model.DatamodelToDatamodelInfoMutex.Lock()
			schema.Info.Title = model.DatamodelToDatamodelInfo[name].Title
			model.DatamodelToDatamodelInfoMutex.Unlock()
			log.Info().Msgf("Updated title: %s for %s openapi spec", schema.Info.Title, name)
		}
		schemasMutex.Unlock()
	}
}

// AddPath creates and adds paths for all the methods of a URI.
func AddPath(uri nexus.RestURIs, datamodel string) {
	crdType, _ := model.GetURIToCRDType(uri.Uri)
	crdInfo, _ := model.GetCRDTypeToNodeInfo(crdType)
	parseSpec(crdType, datamodel)

	params := parseURIParams(uri, crdInfo.ParentHierarchy)
	pathItem := &openapi3.PathItem{}

	for method := range uri.Methods {
		addOperationToPathItem(pathItem, string(method), uri, crdInfo, params)
	}

	log.Info().Msgf("adding %s path to %s", uri.Uri, datamodel)
	schemasMutex.Lock()
	// Register a top-level OpenAPI tag for this node (named after the kind,
	// e.g. "Project") with the nexus-description as its description, so that
	// tooling (Swagger UI etc.) can render the description alongside the tag.
	if crdInfo.Description != "" {
		parts := strings.Split(crdInfo.Name, ".")
		if len(parts) > 1 {
			tagName := parts[1]
			s := Schemas[datamodel]
			if !hasTag(s.Tags, tagName) {
				s.Tags = append(s.Tags, &openapi3.Tag{
					Name:        tagName,
					Description: crdInfo.Description,
				})
				Schemas[datamodel] = s
			}
		}
	}
	existingPathItem := Schemas[datamodel].Paths.Value(uri.Uri)
	if existingPathItem == nil {
		Schemas[datamodel].Paths.Set(uri.Uri, pathItem)
	} else {
		if pathItem.Get != nil {
			existingPathItem.Get = pathItem.Get
		}
		if pathItem.Put != nil {
			existingPathItem.Put = pathItem.Put
		}
		if pathItem.Post != nil {
			existingPathItem.Post = pathItem.Post
		}
		if pathItem.Delete != nil {
			existingPathItem.Delete = pathItem.Delete
		}
		if pathItem.Options != nil {
			existingPathItem.Options = pathItem.Options
		}
		if pathItem.Head != nil {
			existingPathItem.Head = pathItem.Head
		}
		if pathItem.Patch != nil {
			existingPathItem.Patch = pathItem.Patch
		}
		if pathItem.Trace != nil {
			existingPathItem.Trace = pathItem.Trace
		}
	}
	schemasMutex.Unlock()
}

func addOperationToPathItem(pathItem *openapi3.PathItem, method string, uri nexus.RestURIs,
	crdInfo model.NodeInfo, params []*openapi3.ParameterRef,
) {
	opID := getOperationID(method, uri.Uri, crdInfo)
	nameParts := strings.Split(crdInfo.Name, ".")

	switch method {
	case "LIST":
		addListOperation(pathItem, opID, nameParts, params, crdInfo)
	case http.MethodGet:
		addGetOperation(pathItem, opID, nameParts, params, uri, crdInfo)
	case http.MethodPut:
		addPutOperation(pathItem, opID, nameParts, params, uri, crdInfo)
	case http.MethodPatch:
		addPatchOperation(pathItem, opID, nameParts, params, uri, crdInfo)
	case http.MethodDelete:
		addDeleteOperation(pathItem, opID, nameParts, params)
	}
}

func addListOperation(pathItem *openapi3.PathItem, opID string, nameParts []string,
	params []*openapi3.ParameterRef, crdInfo model.NodeInfo,
) {
	operation := &openapi3.Operation{
		OperationID: opID,
		Tags:        []string{nameParts[1]},
		Parameters:  params,
		Responses:   openapi3.NewResponses(),
	}
	operation.Responses.Set("200", &openapi3.ResponseRef{
		Ref: "#/components/responses/List" + crdInfo.Name,
	})
	pathItem.Get = operation
}

func addGetOperation(pathItem *openapi3.PathItem, opID string, nameParts []string,
	params []*openapi3.ParameterRef, uri nexus.RestURIs, crdInfo model.NodeInfo,
) {
	operation := &openapi3.Operation{
		OperationID: opID,
		Tags:        []string{nameParts[1]},
		Parameters:  params,
	}
	if uriInfo, ok := model.GetURIInfo(uri.Uri); ok {
		switch uriInfo.TypeOfURI {
		case model.StatusURI:
			operation.Responses = openapi3.NewResponses()
			operation.Responses.Set("200", &openapi3.ResponseRef{
				Ref: "#/components/responses/Get" + crdInfo.Name + ".Status",
			})
		case model.SingleLinkURI:
			operation.Responses = openapi3.NewResponses()
			operation.Responses.Set("200", &openapi3.ResponseRef{
				Ref: "#/components/responses/Get" + crdInfo.Name + ".SingleLink",
			})
		case model.NamedLinkURI:
			operation.Responses = openapi3.NewResponses()
			operation.Responses.Set("200", &openapi3.ResponseRef{
				Ref: "#/components/responses/Get" + crdInfo.Name + ".NamedLink",
			})
		default:
			operation.Responses = openapi3.NewResponses()
			operation.Responses.Set("200", &openapi3.ResponseRef{
				Ref: "#/components/responses/Get" + crdInfo.Name,
			})
		}
	} else {
		operation.Responses = openapi3.NewResponses()
		operation.Responses.Set("200", &openapi3.ResponseRef{
			Ref: "#/components/responses/DefaultResponse",
		})
	}
	pathItem.Get = operation
}

func addPutOperation(pathItem *openapi3.PathItem, opID string, nameParts []string,
	params []*openapi3.ParameterRef, uri nexus.RestURIs, crdInfo model.NodeInfo,
) {
	operation := &openapi3.Operation{
		OperationID: opID,
		Tags:        []string{nameParts[1]},
	}
	if uriInfo, ok := model.GetURIInfo(uri.Uri); ok && uriInfo.TypeOfURI == model.StatusURI {
		operation.RequestBody = &openapi3.RequestBodyRef{
			Ref: "#/components/requestBodies/Create" + crdInfo.Name + ".Status",
		}
		operation.Responses = openapi3.NewResponses()
		operation.Responses.Set("200", &openapi3.ResponseRef{
			Ref: "#/components/responses/DefaultResponse",
		})
		operation.Parameters = params
	} else {
		p := constructUpdateParam()
		var putParams []*openapi3.ParameterRef
		putParams = append(putParams, params...)
		putParams = append(putParams, p)
		operation.Parameters = putParams

		operation.RequestBody = &openapi3.RequestBodyRef{
			Ref: "#/components/requestBodies/Create" + crdInfo.Name,
		}
		operation.Responses = openapi3.NewResponses()
		operation.Responses.Set("200", &openapi3.ResponseRef{
			Ref: "#/components/responses/DefaultResponse",
		})
	}
	pathItem.Put = operation
}

func addPatchOperation(pathItem *openapi3.PathItem, opID string, nameParts []string,
	params []*openapi3.ParameterRef, uri nexus.RestURIs, crdInfo model.NodeInfo,
) {
	operation := &openapi3.Operation{
		OperationID: opID,
		Tags:        []string{nameParts[1]},
		Parameters:  params,
	}
	operation.Responses = openapi3.NewResponses()
	operation.Responses.Set("200", &openapi3.ResponseRef{
		Ref: "#/components/responses/DefaultResponse",
	})
	operation.Responses.Set("404", &openapi3.ResponseRef{
		Ref: "#/components/responses/NotFoundResponse",
	})
	if uriInfo, ok := model.GetURIInfo(uri.Uri); ok && uriInfo.TypeOfURI == model.StatusURI {
		operation.RequestBody = &openapi3.RequestBodyRef{
			Ref: "#/components/requestBodies/Create" + crdInfo.Name + ".Status",
		}
	} else {
		operation.RequestBody = &openapi3.RequestBodyRef{
			Ref: "#/components/requestBodies/Create" + crdInfo.Name,
		}
	}
	pathItem.Patch = operation
}

func addDeleteOperation(pathItem *openapi3.PathItem, opID string, nameParts []string, params []*openapi3.ParameterRef) {
	operation := &openapi3.Operation{
		OperationID: opID,
		Tags:        []string{nameParts[1]},
		Responses:   openapi3.NewResponses(),
		Parameters:  params,
	}
	operation.Responses.Set("200", &openapi3.ResponseRef{
		Value: openapi3.NewResponse().WithDescription("No content"),
	})
	pathItem.Delete = operation
}

// parseSpec parses openapi schema spec and status subresource.
func parseSpec(crdType, datamodel string) {
	crdInfo, _ := model.GetCRDTypeToNodeInfo(crdType)
	crdSpec, _ := model.GetCrdTypeToSpec(crdType)

	log.Debug().Msgf("Received datamodel update notification for %s", datamodel)
	log.Debug().Msgf("CRD info %#v", crdInfo)
	log.Debug().Msgf("CRD spec %#v", crdSpec)

	getKey := makeKey(crdInfo.Name, "Get")
	postKey := makeKey(crdInfo.Name, "Post")
	listKey := makeKey(crdInfo.Name, "List")
	statusKey := makeKey(crdInfo.Name, "Status")
	singleLinkKey := makeKey(crdInfo.Name, "SingleLink")
	namedLinkKey := makeKey(crdInfo.Name, "NamedLink")

	openapiSchema := crdSpec.Versions[0].Schema.OpenAPIV3Schema
	specProps := openapiSchema.Properties["spec"].Properties
	jsonSpecSchema := openapi3.NewObjectSchema()
	parseFields(jsonSpecSchema, specProps)

	statusProps := openapiSchema.Properties["status"].Properties
	delete(statusProps, "nexus")
	jsonStatusSchema := openapi3.NewObjectSchema()
	parseFields(jsonStatusSchema, statusProps)

	schemasMutex.Lock()
	defer schemasMutex.Unlock()

	Schemas[datamodel].Components.Schemas[statusKey] = openapi3.NewSchemaRef("", jsonStatusSchema)

	jsonSpecAndStatusSchema := openapi3.NewObjectSchema()
	jsonSpecAndStatusSchema.WithProperty("spec", jsonSpecSchema)
	jsonSpecAndStatusSchema.WithProperty("status", jsonStatusSchema)

	Schemas[datamodel].Components.Schemas[postKey] = openapi3.NewSchemaRef("", jsonSpecSchema)
	Schemas[datamodel].Components.Schemas[getKey] = openapi3.NewSchemaRef("", jsonSpecAndStatusSchema)

	jsonListObjectSchema := openapi3.NewObjectSchema()
	jsonListObjectSchema.WithProperty("name", openapi3.NewStringSchema())
	jsonListObjectSchema.WithProperty("spec", jsonSpecSchema)
	jsonListObjectSchema.WithProperty("status", jsonStatusSchema)
	jsonListSchema := openapi3.NewArraySchema().WithItems(jsonListObjectSchema)

	Schemas[datamodel].Components.Schemas[listKey] = openapi3.NewSchemaRef("", jsonListSchema)

	// TODO: Schema for single link and named link need to be generated.
	jsonSingleLinkSchema := openapi3.NewObjectSchema()
	jsonNamedLinkSchema := openapi3.NewArraySchema().WithItems(jsonSingleLinkSchema)
	Schemas[datamodel].Components.Schemas[singleLinkKey] = openapi3.NewSchemaRef("", jsonSingleLinkSchema)
	Schemas[datamodel].Components.Schemas[namedLinkKey] = openapi3.NewSchemaRef("", jsonNamedLinkSchema)

	Schemas[datamodel].Components.RequestBodies["Create"+crdInfo.Name] = &openapi3.RequestBodyRef{
		Value: openapi3.NewRequestBody().
			WithDescription("Request used to create " + crdInfo.Name).
			WithRequired(true).
			WithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + postKey}),
	}

	Schemas[datamodel].Components.Responses["Get"+crdInfo.Name] = &openapi3.ResponseRef{
		Value: openapi3.NewResponse().
			WithDescription("Response returned back after getting " + crdInfo.Name + " object").
			WithContent(
				openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + getKey}),
			),
	}

	Schemas[datamodel].Components.RequestBodies["Create"+statusKey] = &openapi3.RequestBodyRef{
		Value: openapi3.NewRequestBody().
			WithDescription("Request used to create Status subresource of " + crdInfo.Name).
			WithRequired(false).
			WithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + statusKey}),
	}

	Schemas[datamodel].Components.Responses["Get"+statusKey] = &openapi3.ResponseRef{
		Value: openapi3.NewResponse().
			WithDescription("Response returned back after getting status subresource of " + crdInfo.Name + " object").
			WithContent(
				openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + statusKey}),
			),
	}

	Schemas[datamodel].Components.Responses["List"+crdInfo.Name] = &openapi3.ResponseRef{
		Value: openapi3.NewResponse().
			WithDescription("Response returned back after getting " + crdInfo.Name + " objects").
			WithContent(
				openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + listKey}),
			),
	}

	Schemas[datamodel].Components.Responses["Get"+singleLinkKey] = &openapi3.ResponseRef{
		Value: openapi3.NewResponse().
			WithDescription("Response returned back after getting " + crdInfo.Name + " objects").
			WithContent(
				openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + singleLinkKey}),
			),
	}

	Schemas[datamodel].Components.Responses["Get"+namedLinkKey] = &openapi3.ResponseRef{
		Value: openapi3.NewResponse().
			WithDescription("Response returned back after getting " + crdInfo.Name + " objects").
			WithContent(
				openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + namedLinkKey}),
			),
	}
}

// parseFields parses openapi schema fields and attaches description/example
// metadata propagated from the source datamodel (nexus-description /
// nexus-example comment annotations surfaced through the CRD JSON schema).
func parseFields(jsonSchema *openapi3.Schema, specProps map[string]v1.JSONSchemaProps) {
	for name, prop := range specProps {
		if strings.Contains(name, "Gvk") {
			continue
		}
		schema := buildPropSchema(prop)
		if schema == nil {
			continue
		}
		if prop.Description != "" {
			schema.Description = prop.Description
		}
		if prop.Example != nil {
			var ex interface{}
			if err := json.Unmarshal(prop.Example.Raw, &ex); err == nil {
				schema.Example = ex
			}
		}
		jsonSchema.WithProperty(name, schema)
	}
}

// buildPropSchema constructs an openapi3.Schema from a single v1.JSONSchemaProps.
// It handles primitives, objects (including maps via additionalProperties),
// and arrays. Description/example are NOT attached here — the caller does that
// so the same helper can be used recursively for nested array/map items.
func buildPropSchema(prop v1.JSONSchemaProps) *openapi3.Schema {
	switch prop.Type {
	case "string":
		switch prop.Format {
		case "byte":
			return openapi3.NewBytesSchema()
		case "date-time":
			return openapi3.NewDateTimeSchema()
		default:
			return openapi3.NewStringSchema()
		}
	case "boolean":
		return openapi3.NewBoolSchema()
	case "integer":
		switch prop.Format {
		case "int32":
			return openapi3.NewInt32Schema()
		case "int64":
			return openapi3.NewInt64Schema()
		default:
			return openapi3.NewIntegerSchema()
		}
	case "number":
		return openapi3.NewFloat64Schema()
	case "object":
		schema := openapi3.NewObjectSchema()
		if prop.AdditionalProperties != nil && prop.AdditionalProperties.Schema != nil {
			schema.WithAdditionalProperties(buildPropSchema(*prop.AdditionalProperties.Schema))
		} else {
			parseFields(schema, prop.Properties)
		}
		return schema
	case "array":
		if prop.Items != nil && prop.Items.Schema != nil {
			return openapi3.NewArraySchema().WithItems(buildPropSchema(*prop.Items.Schema))
		}
		return openapi3.NewArraySchema()
	default:
		log.Info().Msgf("Unknown type %s", prop.Type)
		return nil
	}
}

// hasTag returns true if the tag with the given name already exists in the slice.
func hasTag(tags openapi3.Tags, name string) bool {
	for _, t := range tags {
		if t.Name == name {
			return true
		}
	}
	return false
}

// parseURIParams parses the URI parameters and headers.
func parseURIParams(restURI nexus.RestURIs, hierarchy []string) []*openapi3.ParameterRef {
	r := regexp.MustCompile(`{([^{}]+)}`)
	params := r.FindAllStringSubmatch(restURI.Uri, -1)

	// Get a snapshot of node info to avoid concurrent map access
	allNodeInfo := model.GetAllCrdTypeToNodeInfo()

	parameters := make([]*openapi3.ParameterRef, 0, len(params)+len(hierarchy)+len(restURI.Headers))
	for _, param := range params {
		description := "Name of the " + param[1] + " node"
		for _, nodeInfo := range allNodeInfo {
			if nodeInfo.Name == param[1] {
				if nodeInfo.Description != "" {
					description = nodeInfo.Description
					break
				}
			}
		}
		parameters = append(parameters, &openapi3.ParameterRef{
			Value: openapi3.NewPathParameter(param[1]).
				WithRequired(true).
				WithSchema(openapi3.NewStringSchema()).
				WithDescription(description),
		})
	}

	// Add header parameters
	for _, headerName := range restURI.Headers {
		description := "Header for " + headerName
		for _, nodeInfo := range allNodeInfo {
			if nodeInfo.Name == headerName {
				if nodeInfo.Description != "" {
					description = nodeInfo.Description
					break
				}
			}
		}
		parameters = append(parameters, &openapi3.ParameterRef{
			Value: openapi3.NewHeaderParameter(headerName).
				WithRequired(true).
				WithSchema(openapi3.NewStringSchema()).
				WithDescription(description),
		})
	}

	for _, parent := range hierarchy {
		crdInfo, ok := model.GetCRDTypeToNodeInfo(parent)
		if !ok || crdInfo.IsSingleton {
			continue
		}
		var description string
		if crdInfo.Description != "" {
			description = crdInfo.Description
		} else {
			description = "Name of the " + crdInfo.Name + " node"
		}

		// Skip if parent is already in URI path or headers
		if paramExist(crdInfo.Name, params) || headerParamExist(crdInfo.Name, restURI.Headers) {
			continue
		}

		// If a header alias is configured for this parent (via
		// config.headerAliases, e.g. orgs.Org -> x-org-id), emit it as a
		// required header parameter instead of a query parameter, mirroring
		// the behaviour of the standalone openapi-generator.
		if headerName := headerAliasFor(crdInfo.Name); headerName != "" {
			if rawHeaderNameExist(headerName, restURI.Headers) {
				continue
			}
			parameters = append(parameters, &openapi3.ParameterRef{
				Value: openapi3.NewHeaderParameter(headerName).
					WithRequired(true).
					WithSchema(openapi3.NewStringSchema()).
					WithDescription(description),
			})
			continue
		}

		parameters = append(parameters, &openapi3.ParameterRef{
			Value: openapi3.NewQueryParameter(crdInfo.Name).
				WithRequired(true).
				WithSchema(openapi3.NewStringSchema()).
				WithDescription(description),
		})
	}
	return parameters
}

// headerAliasFor returns the configured HTTP header name for a given
// parent node type (e.g. "orgs.Org" -> "x-org-id"), or "" if no alias is
// configured. Reads from config.Cfg.HeaderAliases which is populated from
// the api-gw helm values.
func headerAliasFor(nodeType string) string {
	if config.Cfg == nil || config.Cfg.HeaderAliases == nil {
		return ""
	}
	return config.Cfg.HeaderAliases[nodeType]
}

// rawHeaderNameExist returns true if any entry in headers matches the
// given raw HTTP header name (case-insensitive). Used to avoid emitting a
// duplicate header parameter when the spec already declares the alias name.
func rawHeaderNameExist(headerName string, headers []string) bool {
	for _, h := range headers {
		if strings.EqualFold(h, headerName) {
			return true
		}
	}
	return false
}

func constructUpdateParam() *openapi3.ParameterRef {
	return &openapi3.ParameterRef{
		Value: openapi3.NewQueryParameter("update_if_exists").
			WithRequired(false).
			WithSchema(openapi3.NewBoolSchema()).
			WithDescription("If set to false, disables update of preexisting object. Default value is true"),
	}
}

func paramExist(param string, params [][]string) bool {
	for _, p := range params {
		if p[1] == param {
			return true
		}
	}
	return false
}

func headerParamExist(param string, headers []string) bool {
	for _, h := range headers {
		if h == param {
			return true
		}
	}
	return false
}

func Recreate() {
	log.Debug().Msg("Recreating openapi spec")

	// Get a snapshot of the map to avoid concurrent access
	allCrdTypeToRestUris := model.GetAllCrdTypeToRestUris()

	for crdType, uris := range allCrdTypeToRestUris {
		datamodel := utils.GetDatamodelName(crdType)

		// Only create schema if it doesn't exist - don't clear existing paths
		schemasMutex.RLock()
		_, exists := Schemas[datamodel]
		schemasMutex.RUnlock()
		if !exists {
			log.Debug().Msgf("Creating new openapi schema for %s", datamodel)
			New(datamodel)
		}

		for _, uri := range uris {
			log.Debug().Msgf("Adding path %s for %s Datamodel name %s", uri.Uri, crdType, datamodel)
			AddPath(uri, datamodel)
		}
	}
}

// RecreateExtension adds ExtensionRestAPI paths to the OpenAPI spec.
func RecreateExtension() {
	log.Debug().Msg("Recreating extension openapi spec")
	specs := model.GetAllExtensionRestAPISpecs()

	for _, spec := range specs {
		datamodel := spec.Datamodel
		if datamodel == "" {
			datamodel = "extension.nexus.com"
		}

		// Ensure schema exists for this datamodel
		schemasMutex.RLock()
		_, exists := Schemas[datamodel]
		schemasMutex.RUnlock()
		if !exists {
			New(datamodel)
		}

		log.Debug().Msgf("Adding extension path %s for datamodel %s", spec.URI, datamodel)
		AddExtensionPath(spec, datamodel)
	}
}

// AddExtensionPath adds an ExtensionRestAPI path to the OpenAPI schema.
func AddExtensionPath(spec model.ExtensionRestAPISpec, datamodel string) {
	schemasMutex.RLock()
	_, exists := Schemas[datamodel]
	schemasMutex.RUnlock()
	if !exists {
		New(datamodel)
	}

	// Parse the OpenAPIPathSpec YAML to extract path operations
	pathItem := parseOpenAPIPathSpec(spec.OpenAPIPathSpec)
	if pathItem == nil {
		log.Warn().Msgf("Failed to parse OpenAPIPathSpec for %s", spec.URI)
		return
	}

	log.Info().Msgf("Adding extension path %s to %s", spec.URI, datamodel)
	schemasMutex.Lock()
	existingPathItem := Schemas[datamodel].Paths.Value(spec.URI)
	if existingPathItem == nil {
		Schemas[datamodel].Paths.Set(spec.URI, pathItem)
	} else {
		if pathItem.Get != nil {
			existingPathItem.Get = pathItem.Get
		}
		if pathItem.Put != nil {
			existingPathItem.Put = pathItem.Put
		}
		if pathItem.Post != nil {
			existingPathItem.Post = pathItem.Post
		}
		if pathItem.Delete != nil {
			existingPathItem.Delete = pathItem.Delete
		}
		if pathItem.Options != nil {
			existingPathItem.Options = pathItem.Options
		}
		if pathItem.Head != nil {
			existingPathItem.Head = pathItem.Head
		}
		if pathItem.Patch != nil {
			existingPathItem.Patch = pathItem.Patch
		}
		if pathItem.Trace != nil {
			existingPathItem.Trace = pathItem.Trace
		}
	}
	schemasMutex.Unlock()
}

// parseOpenAPIPathSpec parses an OpenAPI path spec YAML string into a PathItem.
func parseOpenAPIPathSpec(openAPIPathSpec string) *openapi3.PathItem {
	if openAPIPathSpec == "" {
		return nil
	}

	pathItem := &openapi3.PathItem{}

	// Simple YAML parsing for common operations
	spec := strings.ToLower(openAPIPathSpec)

	if strings.Contains(spec, "get:") {
		pathItem.Get = parseOperation(openAPIPathSpec, "get")
	}
	if strings.Contains(spec, "post:") {
		pathItem.Post = parseOperation(openAPIPathSpec, "post")
	}
	if strings.Contains(spec, "put:") {
		pathItem.Put = parseOperation(openAPIPathSpec, "put")
	}
	if strings.Contains(spec, "patch:") {
		pathItem.Patch = parseOperation(openAPIPathSpec, "patch")
	}
	if strings.Contains(spec, "delete:") {
		pathItem.Delete = parseOperation(openAPIPathSpec, "delete")
	}

	return pathItem
}

// parseOperation extracts an operation from the OpenAPI path spec using proper YAML parsing.
func parseOperation(openAPIPathSpec, method string) *openapi3.Operation {
	// Use kin-openapi to parse the full OpenAPI path spec
	pathItem := &openapi3.PathItem{}
	if err := pathItem.UnmarshalJSON([]byte(convertYAMLToJSON(openAPIPathSpec))); err != nil {
		// Fallback to simple parsing if full parsing fails
		log.Debug().Msgf("Full OpenAPI parsing failed, using simple parsing: %v", err)
		return parseOperationSimple(openAPIPathSpec, method)
	}

	// Get the operation for the specified method
	var parsedOp *openapi3.Operation
	switch strings.ToLower(method) {
	case "get":
		parsedOp = pathItem.Get
	case "post":
		parsedOp = pathItem.Post
	case "put":
		parsedOp = pathItem.Put
	case "patch":
		parsedOp = pathItem.Patch
	case "delete":
		parsedOp = pathItem.Delete
	}

	if parsedOp != nil {
		return parsedOp
	}

	// Fallback to simple parsing
	return parseOperationSimple(openAPIPathSpec, method)
}

// convertYAMLToJSON converts YAML to JSON for OpenAPI parsing.
func convertYAMLToJSON(yamlStr string) string {
	var data interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &data); err != nil {
		return "{}"
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(jsonBytes)
}

// parseOperationSimple is the fallback simple parser for basic fields.
func parseOperationSimple(openAPIPathSpec, method string) *openapi3.Operation {
	op := &openapi3.Operation{
		Responses: openapi3.NewResponses(),
	}

	// Extract tags if present
	if idx := strings.Index(strings.ToLower(openAPIPathSpec), "tags:"); idx != -1 {
		tagStart := idx + 5
		remaining := openAPIPathSpec[tagStart:]
		if dashIdx := strings.Index(remaining, "- "); dashIdx != -1 {
			tagLine := remaining[dashIdx+2:]
			if newlineIdx := strings.Index(tagLine, "\n"); newlineIdx != -1 {
				tag := strings.TrimSpace(tagLine[:newlineIdx])
				op.Tags = []string{tag}
			}
		}
	}

	// Extract operationId if present
	if idx := strings.Index(strings.ToLower(openAPIPathSpec), "operationid:"); idx != -1 {
		opIdStart := idx + 12
		remaining := openAPIPathSpec[opIdStart:]
		if newlineIdx := strings.Index(remaining, "\n"); newlineIdx != -1 {
			opId := strings.TrimSpace(remaining[:newlineIdx])
			op.OperationID = opId
		}
	}

	// Extract summary if present
	if idx := strings.Index(strings.ToLower(openAPIPathSpec), "summary:"); idx != -1 {
		summaryStart := idx + 8
		remaining := openAPIPathSpec[summaryStart:]
		if newlineIdx := strings.Index(remaining, "\n"); newlineIdx != -1 {
			summary := strings.TrimSpace(remaining[:newlineIdx])
			op.Summary = summary
		}
	}

	// Add default 200 response
	op.Responses.Set("200", &openapi3.ResponseRef{
		Value: openapi3.NewResponse().WithDescription("Success"),
	})

	return op
}

func LoadCombinedSpec() {
	// Path to the OpenAPI specification JSON file
	specFilePath := "/static/openapispecs/combined/combined_spec.yaml"

	// Create a new OpenAPI loader
	loader := openapi3.NewLoader()

	// Load the OpenAPI specification from the file
	doc, err := loader.LoadFromFile(specFilePath)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Failed to load OpenAPI spec: %v", err))
		return
	}

	// Print the title of the OpenAPI specification
	fmt.Printf("OpenAPI Title: %s\n", doc.Info.Title)

	Schemas["edge-orchestrator.intel.com"] = *doc
}

func makeKey(crd, keyType string) string {
	return crd + "." + keyType
}
