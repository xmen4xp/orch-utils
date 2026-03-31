// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/hex"
	"fmt"
	"hash"
	"net/http"
	"regexp"
	"strings"

	"nexus-api-gw/pkg/model"
	"nexus-api-gw/pkg/utils"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
	"golang.org/x/crypto/sha3"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
)

var (
	Schemas = make(map[string]openapi3.T)
	appName = "nexus-api-gw-openapi"
	log     = logging.GetLogger(appName)
)

func New(datamodel string) {
	// Check if datamodel info is present.
	title := "Nexus API GW APIs"
	description := ""
	version := "1.0.0"

	info, hasDatamodelInfo := model.DatamodelToDatamodelInfo[datamodel]
	if hasDatamodelInfo {
		if info.Title != "" {
			title = info.Title
		}
		if info.Description != "" {
			description = info.Description
		}
		if info.Version != "" {
			version = info.Version
		}
	}

	schema := openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:          title,
			Description:    description,
			TermsOfService: "",
			Contact:        nil,
			License:        nil,
			Version:        version,
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
			Schemas:         openapi3.Schemas{},
			RequestBodies:   openapi3.RequestBodies{},
			SecuritySchemes: openapi3.SecuritySchemes{},
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

	// Apply security schemes from DatamodelInfo (from CRD)
	if hasDatamodelInfo {
		for schemeName, scheme := range info.SecuritySchemes {
			secScheme := &openapi3.SecurityScheme{
				Type:         scheme.Type,
				Scheme:       scheme.Scheme,
				BearerFormat: scheme.BearerFormat,
				In:           scheme.In,
				Name:         scheme.Name,
				Description:  scheme.Description,
			}
			schema.Components.SecuritySchemes[schemeName] = &openapi3.SecuritySchemeRef{
				Value: secScheme,
			}
		}

		// Apply global security requirements
		if len(info.Security) > 0 {
			schema.Security = make(openapi3.SecurityRequirements, 0, len(info.Security))
			for _, secReq := range info.Security {
				schema.Security = append(schema.Security, secReq)
			}
		}
	}

	log.Info().Msgf("Created schema for %s with %d security schemes", datamodel, len(schema.Components.SecuritySchemes))
	Schemas[datamodel] = schema
}

func DatamodelUpdateNotification() {
	log.Debug().Msg("Started datamodel update notification")
	for name := range model.DatamodelsChan {
		log.Debug().Msgf("Received datamodel update notification for %s", name)
		if _, ok := Schemas[name]; !ok {
			New(name)
		}

		if schema, ok := Schemas[name]; ok {
			model.DatamodelToDatamodelInfoMutex.Lock()
			info := model.DatamodelToDatamodelInfo[name]

			// Update Info section
			if info.Title != "" {
				schema.Info.Title = info.Title
			}
			if info.Description != "" {
				schema.Info.Description = info.Description
			}
			if info.Version != "" {
				schema.Info.Version = info.Version
			}

			// Update security schemes
			if schema.Components.SecuritySchemes == nil {
				schema.Components.SecuritySchemes = openapi3.SecuritySchemes{}
			}
			for schemeName, scheme := range info.SecuritySchemes {
				secScheme := &openapi3.SecurityScheme{
					Type:         scheme.Type,
					Scheme:       scheme.Scheme,
					BearerFormat: scheme.BearerFormat,
					In:           scheme.In,
					Name:         scheme.Name,
					Description:  scheme.Description,
				}
				schema.Components.SecuritySchemes[schemeName] = &openapi3.SecuritySchemeRef{
					Value: secScheme,
				}
			}

			// Update global security requirements
			if len(info.Security) > 0 {
				schema.Security = make(openapi3.SecurityRequirements, 0, len(info.Security))
				for _, secReq := range info.Security {
					schema.Security = append(schema.Security, secReq)
				}
			}

			model.DatamodelToDatamodelInfoMutex.Unlock()
			log.Info().Msgf("Updated openapi spec for %s: title=%s, security_schemes=%d",
				name, schema.Info.Title, len(schema.Components.SecuritySchemes))
		}
	}
}

// AddPath creates and adds paths for all the methods of a URI.
func AddPath(uri nexus.RestURIs, datamodel string) {
	crdType := model.URIToCRDType[uri.Uri]
	crdInfo := model.CrdTypeToNodeInfo[crdType]
	parseSpec(crdType, datamodel)

	h := sha3.New256()
	params := parseURIParams(uri, crdInfo.ParentHierarchy, datamodel)
	pathItem := &openapi3.PathItem{}

	for method := range uri.Methods {
		addOperationToPathItem(pathItem, string(method), uri, crdInfo, params, h, datamodel)
	}

	log.Info().Msgf("adding %s path to %s", uri.Uri, datamodel)
	Schemas[datamodel].Paths.Set(uri.Uri, pathItem)
}

func addOperationToPathItem(pathItem *openapi3.PathItem, method string, uri nexus.RestURIs,
	crdInfo model.NodeInfo, params []*openapi3.ParameterRef, h hash.Hash, datamodel string,
) {
	formedStr := fmt.Sprintf("%s%s", method, uri.Uri)
	h.Write([]byte(formedStr))
	fmt.Fprintf(h, "%s%s", method, uri.Uri)
	opID := hex.EncodeToString(h.Sum(nil))
	nameParts := strings.Split(crdInfo.Name, ".")

	// Get operation-level security override if specified
	var opSecurity *openapi3.SecurityRequirements
	if uri.Security != nil {
		opSecurity = &openapi3.SecurityRequirements{}
		for _, secReq := range *uri.Security {
			*opSecurity = append(*opSecurity, secReq)
		}
	}

	switch method {
	case "LIST":
		addListOperation(pathItem, opID, nameParts, params, crdInfo, opSecurity)
	case http.MethodGet:
		addGetOperation(pathItem, opID, nameParts, params, uri, crdInfo, opSecurity)
	case http.MethodPut:
		addPutOperation(pathItem, opID, nameParts, params, uri, crdInfo, opSecurity)
	case http.MethodPatch:
		addPatchOperation(pathItem, opID, nameParts, params, uri, crdInfo, opSecurity)
	case http.MethodDelete:
		addDeleteOperation(pathItem, opID, nameParts, params, opSecurity)
	}
}

func addListOperation(pathItem *openapi3.PathItem, opID string, nameParts []string,
	params []*openapi3.ParameterRef, crdInfo model.NodeInfo, security *openapi3.SecurityRequirements,
) {
	operation := &openapi3.Operation{
		OperationID: opID,
		Tags:        []string{nameParts[1]},
		Parameters:  params,
		Responses:   openapi3.NewResponses(),
	}
	if security != nil {
		operation.Security = security
	}
	operation.Responses.Set("200", &openapi3.ResponseRef{
		Ref: "#/components/responses/List" + crdInfo.Name,
	})
	pathItem.Get = operation
}

func addGetOperation(pathItem *openapi3.PathItem, opID string, nameParts []string,
	params []*openapi3.ParameterRef, uri nexus.RestURIs, crdInfo model.NodeInfo, security *openapi3.SecurityRequirements,
) {
	operation := &openapi3.Operation{
		OperationID: opID,
		Tags:        []string{nameParts[1]},
		Parameters:  params,
	}
	if security != nil {
		operation.Security = security
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
	params []*openapi3.ParameterRef, uri nexus.RestURIs, crdInfo model.NodeInfo, security *openapi3.SecurityRequirements,
) {
	operation := &openapi3.Operation{
		OperationID: opID,
		Tags:        []string{nameParts[1]},
	}
	if security != nil {
		operation.Security = security
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
	params []*openapi3.ParameterRef, uri nexus.RestURIs, crdInfo model.NodeInfo, security *openapi3.SecurityRequirements,
) {
	operation := &openapi3.Operation{
		OperationID: opID,
		Tags:        []string{nameParts[1]},
		Parameters:  params,
	}
	if security != nil {
		operation.Security = security
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

func addDeleteOperation(pathItem *openapi3.PathItem, opID string, nameParts []string, params []*openapi3.ParameterRef, security *openapi3.SecurityRequirements) {
	operation := &openapi3.Operation{
		OperationID: opID,
		Tags:        []string{nameParts[1]},
		Responses:   openapi3.NewResponses(),
		Parameters:  params,
	}
	if security != nil {
		operation.Security = security
	}
	operation.Responses.Set("200", &openapi3.ResponseRef{
		Value: openapi3.NewResponse().WithDescription("No content"),
	})
	pathItem.Delete = operation
}

// parseSpec parses openapi schema spec and status subresource.
func parseSpec(crdType, datamodel string) {
	crdInfo := model.CrdTypeToNodeInfo[crdType]
	crdSpec := model.CrdTypeToSpec[crdType]

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

// ParseFields parses openapi schema fields.
// parseFields parses openapi schema fields.
func parseFields(jsonSchema *openapi3.Schema, specProps map[string]v1.JSONSchemaProps) {
	for name, prop := range specProps {
		if strings.Contains(name, "Gvk") {
			continue
		}
		addPropertyToSchema(jsonSchema, name, prop)
	}
}

func addPropertyToSchema(jsonSchema *openapi3.Schema, name string, prop v1.JSONSchemaProps) {
	switch prop.Type {
	case "string":
		addStringProperty(jsonSchema, name, prop)
	case "boolean":
		jsonSchema.WithProperty(name, openapi3.NewBoolSchema())
	case "object":
		schema := openapi3.NewSchema()
		parseFields(schema, prop.Properties)
		jsonSchema.WithProperty(name, schema)
	case "integer":
		addIntegerProperty(jsonSchema, name, prop)
	case "number":
		jsonSchema.WithProperty(name, openapi3.NewFloat64Schema())
	case "array":
		schema := openapi3.NewSchema()
		parseFields(schema, prop.Items.Schema.Properties)
		arraySchema := openapi3.NewArraySchema().WithItems(schema)
		jsonSchema.WithProperty(name, arraySchema)
	default:
		log.Info().Msgf("Unknown type %s", prop.Type)
	}
}

func addStringProperty(jsonSchema *openapi3.Schema, name string, prop v1.JSONSchemaProps) {
	switch prop.Format {
	case "byte":
		jsonSchema.WithProperty(name, openapi3.NewBytesSchema())
	case "date-time":
		jsonSchema.WithProperty(name, openapi3.NewDateTimeSchema())
	default:
		jsonSchema.WithProperty(name, openapi3.NewStringSchema())
	}
}

func addIntegerProperty(jsonSchema *openapi3.Schema, name string, prop v1.JSONSchemaProps) {
	switch prop.Format {
	case "int32":
		jsonSchema.WithProperty(name, openapi3.NewInt32Schema())
	case "int64":
		jsonSchema.WithProperty(name, openapi3.NewInt64Schema())
	default:
		jsonSchema.WithProperty(name, openapi3.NewIntegerSchema())
	}
}

// parseURIParams parses the URI parameters and headers.
func parseURIParams(restURI nexus.RestURIs, hierarchy []string, datamodel string) []*openapi3.ParameterRef {
	r := regexp.MustCompile(`{([^{}]+)}`)
	params := r.FindAllStringSubmatch(restURI.Uri, -1)

	parameters := make([]*openapi3.ParameterRef, 0, len(params)+len(hierarchy)+len(restURI.HeaderParams))
	for _, param := range params {
		description := "Name of the " + param[1] + " node"
		for _, nodeInfo := range model.CrdTypeToNodeInfo {
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

	// Add header parameters only if they're not covered by security schemes
	for headerName, nodeType := range restURI.HeaderParams {
		// Skip if this header is covered by a security scheme
		if isHeaderInSecuritySchemes(headerName, datamodel) {
			continue
		}
		description := "Header for " + nodeType
		for _, nodeInfo := range model.CrdTypeToNodeInfo {
			if nodeInfo.Name == nodeType {
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
		crdInfo := model.CrdTypeToNodeInfo[parent]
		if crdInfo.IsSingleton {
			continue
		}
		var description string
		if crdInfo.Description != "" {
			description = crdInfo.Description
		} else {
			description = "Name of the " + crdInfo.Name + " node"
		}

		// Skip if parent is already in URI path, PathParams map, or headers
		if !paramExist(crdInfo.Name, params) && !pathParamMapsToNodeType(crdInfo.Name, restURI.PathParams) && !headerParamExist(crdInfo.Name, restURI.HeaderParams) {
			parameters = append(parameters, &openapi3.ParameterRef{
				Value: openapi3.NewQueryParameter(crdInfo.Name).
					WithRequired(true).
					WithSchema(openapi3.NewStringSchema()).
					WithDescription(description),
			})
		}
	}
	return parameters
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

func headerParamExist(param string, headers map[string]string) bool {
	for _, nodeType := range headers {
		if nodeType == param {
			return true
		}
	}
	return false
}

// pathParamMapsToNodeType checks if any PathParams value maps to the given node type
func pathParamMapsToNodeType(nodeType string, pathParams map[string]string) bool {
	for _, mappedNodeType := range pathParams {
		if mappedNodeType == nodeType {
			return true
		}
	}
	return false
}

// isHeaderInSecuritySchemes checks if a header is already covered by security schemes in the specified datamodel
func isHeaderInSecuritySchemes(headerName string, datamodel string) bool {
	model.DatamodelToDatamodelInfoMutex.RLock()
	defer model.DatamodelToDatamodelInfoMutex.RUnlock()

	dmInfo, ok := model.DatamodelToDatamodelInfo[datamodel]
	if !ok {
		return false
	}

	for _, scheme := range dmInfo.SecuritySchemes {
		if scheme.Type == "apiKey" && scheme.In == "header" && scheme.Name == headerName {
			return true
		}
	}
	return false
}

func Recreate() {
	log.Debug().Msg("Recreating openapi spec")
	for crdType := range model.CrdTypeToRestUris {
		log.Debug().Msgf("Recreating openapi spec for %s Datamodel name %s", crdType, utils.GetDatamodelName(crdType))
		New(utils.GetDatamodelName(crdType))
	}

	for crdType, uris := range model.CrdTypeToRestUris {
		datamodel := utils.GetDatamodelName(crdType)
		for _, uri := range uris {
			log.Debug().Msgf("Adding path %s for %s Datamodel name %s", uri.Uri, crdType, datamodel)
			AddPath(uri, datamodel)
		}
	}
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

// getGlobalSecurityHeaders returns a map of header names that are covered by global security schemes
func getGlobalSecurityHeaders(datamodel string) map[string]bool {
	headers := make(map[string]bool)

	model.DatamodelToDatamodelInfoMutex.RLock()
	defer model.DatamodelToDatamodelInfoMutex.RUnlock()

	info, ok := model.DatamodelToDatamodelInfo[datamodel]
	if !ok {
		return headers
	}

	// Check all security schemes used in global security requirements
	for _, secReq := range info.Security {
		for schemeName := range secReq {
			if scheme, exists := info.SecuritySchemes[schemeName]; exists {
				// Only collect apiKey type schemes in headers
				if scheme.Type == "apiKey" && scheme.In == "header" && scheme.Name != "" {
					headers[scheme.Name] = true
				}
			}
		}
	}

	return headers
}
