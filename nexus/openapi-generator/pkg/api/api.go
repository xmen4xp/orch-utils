// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	yamlv1 "github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"nexus/openapi-generator/pkg/model"
)

var Schemas = make(map[string]openapi3.T)

func New(datamodel string) {
	// Check if datamodel info is present
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
		Paths: &openapi3.Paths{},
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
	Schemas[datamodel] = schema
}

// getOperationID returns an OpenAPI operationId for a datamodel-derived route.
//
// Scheme:
//   - LIST                      -> list<KindPlural>      (Kind-derived plural; if Kind already ends in "s", use Kind verbatim)
//   - GET item                  -> get<Kind>
//   - PUT/PATCH/DELETE item     -> <verb><Kind>
//   - StatusURI (any verb)      -> <verb><Kind>Status
//   - SingleLink / NamedLink    -> get<ParentKind><FieldName>  (FieldName = last URI segment)
//
// No parent prefix is used for CRUD/LIST — only link traversal carries it because
// link traversal is semantically a navigation from the parent.
func getOperationID(method, uri string, crdInfo model.NodeInfo) string {
	nameParts := strings.Split(crdInfo.Name, ".")
	kind := qualifiedKind(nameParts)
	uriInfo, _ := model.GetUriInfo(uri)

	switch method {
	case "LIST":
		return "list" + kindPlural(kind)
	case http.MethodGet:
		switch uriInfo.TypeOfURI {
		case model.StatusURI:
			return "get" + kind + "Status"
		case model.SingleLinkURI, model.NamedLinkURI:
			return "get" + kind + lastStaticSegment(uri)
		}
		return "get" + kind
	case http.MethodPut:
		if uriInfo.TypeOfURI == model.StatusURI {
			return "put" + kind + "Status"
		}
		return "put" + kind
	case http.MethodPatch:
		if uriInfo.TypeOfURI == model.StatusURI {
			return "patch" + kind + "Status"
		}
		return "patch" + kind
	case http.MethodDelete:
		return "delete" + kind
	}
	// Fallback: should not happen for datamodel-derived routes.
	return strings.ToLower(method) + kind
}

// kindPlural returns the plural form of a Kind name, derived solely from the
// Kind itself (no URI-segment heuristics). If the Kind already ends in "s"
// (e.g. "Clusters", "DataCenters", "Nodes"), it is returned unchanged.
// Otherwise an "s" is appended (e.g. "Org" -> "Orgs", "AISlice" -> "AISlices").
func kindPlural(kind string) string {
	if strings.HasSuffix(kind, "s") {
		return kind
	}
	return kind + "s"
}

// lastStaticSegment returns the trailing non-parameter segment of a URI,
// or "" if the URI ends in a path parameter "{...}".
func lastStaticSegment(uri string) string {
	parts := strings.Split(strings.TrimRight(uri, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]
	if strings.HasPrefix(last, "{") {
		return ""
	}
	return last
}

// AddExtensionPath merges a single ExtensionRestAPI YAML manifest into the
// OpenAPI spec for the given datamodel. The manifest is expected to be a
// nexus ExtensionRestAPI CR with a `spec.uri` and a `spec.openAPIPathSpec`
// holding the inline OpenAPI PathItem definition (verbatim YAML).
//
// This is the build-time counterpart to api-gw's parseOpenAPIPathSpec — it
// guarantees that infra-host.com.json contains the same custom routes
// (metrics, recommendations, etc.) that api-gw serves at runtime.
//
// Files that are not ExtensionRestAPI manifests are silently skipped so the
// caller can pass a mixed directory.
func AddExtensionPath(fileBytes []byte, datamodel string) {
	// Decode the CR envelope to read spec.uri and spec.openAPIPathSpec.
	var env struct {
		Kind string `json:"kind"`
		Spec struct {
			URI             string `json:"uri"`
			OpenAPIPathSpec string `json:"openAPIPathSpec"`
		} `json:"spec"`
	}
	crJSON, err := yamlv1.YAMLToJSON(fileBytes)
	if err != nil {
		log.Warnf("extension: unable to parse YAML: %v", err)
		return
	}
	if err := json.Unmarshal(crJSON, &env); err != nil {
		log.Warnf("extension: unable to unmarshal CR envelope: %v", err)
		return
	}
	if env.Kind != "ExtensionRestAPI" {
		return
	}
	if env.Spec.URI == "" || env.Spec.OpenAPIPathSpec == "" {
		log.Warnf("extension: skipping %q — missing spec.uri or spec.openAPIPathSpec", env.Spec.URI)
		return
	}

	// Convert the inline OpenAPI PathItem YAML to a PathItem object.
	pathJSON, err := yamlv1.YAMLToJSON([]byte(env.Spec.OpenAPIPathSpec))
	if err != nil {
		log.Warnf("extension: %q: openAPIPathSpec is not valid YAML: %v", env.Spec.URI, err)
		return
	}
	pathItem := &openapi3.PathItem{}
	if err := pathItem.UnmarshalJSON(pathJSON); err != nil {
		log.Warnf("extension: %q: openAPIPathSpec is not a valid PathItem: %v", env.Spec.URI, err)
		return
	}

	schema, ok := Schemas[datamodel]
	if !ok {
		log.Warnf("extension: %q: datamodel %q not initialized; call New() first", env.Spec.URI, datamodel)
		return
	}
	if schema.Paths == nil {
		schema.Paths = &openapi3.Paths{}
	}
	schema.Paths.Set(env.Spec.URI, pathItem)
	Schemas[datamodel] = schema
	log.Infof("adding extension %s path to %s", env.Spec.URI, datamodel)
}

// AddPath creates and adds paths for all the methods of a URI
func AddPath(uri nexus.RestURIs, datamodel string) {
	crdType := model.UriToCRDType[uri.Uri]
	crdInfo := model.CrdTypeToNodeInfo[crdType]
	parseSpec(crdType, datamodel)

	params := parseUriParams(uri, crdInfo.ParentHierarchy)
	pathItem := &openapi3.PathItem{}
	for method := range uri.Methods {
		opId := getOperationID(string(method), uri.Uri, crdInfo)
		nameParts := strings.Split(crdInfo.Name, ".")
		switch method {
		case "LIST":
			resp := &openapi3.Responses{}
			resp.Set("200", &openapi3.ResponseRef{
				Ref: "#/components/responses/List" + crdInfo.Name,
			})
			operation := &openapi3.Operation{
				OperationID: opId,
				Tags:        []string{qualifiedKind(nameParts)},
				Parameters:  params,
				Responses:   resp,
			}
			pathItem.Get = operation
		case http.MethodGet:
			operation := &openapi3.Operation{
				OperationID: opId,
				Tags:        []string{qualifiedKind(nameParts)},
				Parameters:  params,
			}
			if uriInfo, ok := model.GetUriInfo(uri.Uri); ok {
				switch uriInfo.TypeOfURI {
				case model.StatusURI:
					resp := &openapi3.Responses{}
					resp.Set("200", &openapi3.ResponseRef{
						Ref: "#/components/responses/Get" + crdInfo.Name + ".Status",
					})
					operation.Responses = resp
				case model.SingleLinkURI:
					resp := &openapi3.Responses{}
					resp.Set("200", &openapi3.ResponseRef{
						Ref: "#/components/responses/Get" + crdInfo.Name + ".SingleLink",
					})
					operation.Responses = resp
				case model.NamedLinkURI:
					resp := &openapi3.Responses{}
					resp.Set("200", &openapi3.ResponseRef{
						Ref: "#/components/responses/Get" + crdInfo.Name + ".NamedLink",
					})
					operation.Responses = resp
				default:
					resp := &openapi3.Responses{}
					resp.Set("200", &openapi3.ResponseRef{
						Ref: "#/components/responses/Get" + crdInfo.Name,
					})
					operation.Responses = resp
				}
			} else {
				resp := &openapi3.Responses{}
				resp.Set("200", &openapi3.ResponseRef{
					Ref: "#/components/responses/DefaultResponse",
				})
				operation.Responses = resp
			}
			pathItem.Get = operation
		case http.MethodPut:
			operation := &openapi3.Operation{
				OperationID: opId,
				Tags:        []string{qualifiedKind(nameParts)},
			}
			if uriInfo, ok := model.GetUriInfo(uri.Uri); ok && uriInfo.TypeOfURI == model.StatusURI {
				operation.RequestBody = &openapi3.RequestBodyRef{
					Ref: "#/components/requestBodies/Create" + crdInfo.Name + ".Status",
				}
				resp := &openapi3.Responses{}
				resp.Set("200", &openapi3.ResponseRef{
					Ref: "#/components/responses/DefaultResponse",
				})
				operation.Responses = resp
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
				resp := &openapi3.Responses{}
				resp.Set("200", &openapi3.ResponseRef{
					Ref: "#/components/responses/DefaultResponse",
				})
				operation.Responses = resp
			}
			pathItem.Put = operation
		case http.MethodPatch:
			operation := &openapi3.Operation{
				OperationID: opId,
				Tags:        []string{qualifiedKind(nameParts)},
				Parameters:  params,
			}
			resp := &openapi3.Responses{}
			resp.Set("200", &openapi3.ResponseRef{
				Ref: "#/components/responses/DefaultResponse",
			})
			resp.Set("404", &openapi3.ResponseRef{
				Ref: "#/components/responses/NotFoundResponse",
			})
			operation.Responses = resp
			if uriInfo, ok := model.GetUriInfo(uri.Uri); ok && uriInfo.TypeOfURI == model.StatusURI {
				operation.RequestBody = &openapi3.RequestBodyRef{
					Ref: "#/components/requestBodies/Create" + crdInfo.Name + ".Status",
				}
			} else {
				operation.RequestBody = &openapi3.RequestBodyRef{
					Ref: "#/components/requestBodies/Create" + crdInfo.Name,
				}
			}
			pathItem.Patch = operation
		case http.MethodDelete:
			resp := &openapi3.Responses{}
			resp.Set("200", &openapi3.ResponseRef{
				Value: openapi3.NewResponse().WithDescription("No content"),
			})
			operation := &openapi3.Operation{
				OperationID: opId,
				Tags:        []string{qualifiedKind(nameParts)},
				Responses:   resp,
				Parameters:  params,
			}
			pathItem.Delete = operation
		}
	}

	if crdInfo.Description != "" {
		tagName := qualifiedKind(strings.Split(crdInfo.Name, "."))
		if !hasTag(Schemas[datamodel].Tags, tagName) {
			s := Schemas[datamodel]
			s.Tags = append(s.Tags, &openapi3.Tag{
				Name:        tagName,
				Description: crdInfo.Description,
			})
			Schemas[datamodel] = s
		}
	}
	log.Infof("adding %s path to %s", uri.Uri, datamodel)
	Schemas[datamodel].Paths.Set(uri.Uri, pathItem)
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

// parseSpec parses openapi schema spec and status subresource
func parseSpec(crdType string, datamodel string) {
	crdInfo := model.CrdTypeToNodeInfo[crdType]
	crdSpec := model.CrdTypeToSpec[crdType]

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

	// TODO: Schema for single link and named link need to be generated
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

// parseFields parses openapi schema fields
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
// It handles primitives, objects (including maps via additionalProperties), and arrays.
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
		log.Infof("Unknown type %s", prop.Type)
		return nil
	}
}

// parseUriParams parses the URI parameters and headers.
func parseUriParams(restURI nexus.RestURIs, hierarchy []string) (parameters []*openapi3.ParameterRef) {
	r := regexp.MustCompile(`{([^{}]+)}`)
	params := r.FindAllStringSubmatch(restURI.Uri, -1)

	// Resolve each URI token to its canonical groupKind via PathParams. The
	// alias is what appears in the URL (and becomes the OpenAPI parameter
	// name); the canonical is used to look up Node descriptions and to satisfy
	// downstream paramExist checks that key off CRD names.
	resolvedParams := resolveUriParams(params, restURI.PathParams)

	for i, param := range params {
		alias := param[1]
		canonical := resolvedParams[i][1]
		description := "Name of the " + alias + " node"
		for _, nodeInfo := range model.CrdTypeToNodeInfo {
			if nodeInfo.Name == canonical {
				if nodeInfo.Description != "" {
					description = nodeInfo.Description
					break
				}
			}
		}
		parameters = append(parameters, &openapi3.ParameterRef{
			Value: openapi3.NewPathParameter(alias).
				WithRequired(true).
				WithSchema(openapi3.NewStringSchema()).
				WithDescription(description),
		})
	}

	// Add header parameters
	for _, headerName := range restURI.Headers {
		description := "Header for " + headerName
		for _, nodeInfo := range model.CrdTypeToNodeInfo {
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

		if _, ok := model.OpenApiIgnoredParentPathParams[crdInfo.Name]; ok {
			// Ignored parents are normally dropped from the spec entirely. If a
			// nodeToHeaderMapping entry exists for this parent, surface it as a
			// required header parameter instead (e.g. orgs.Org -> x-org-id).
			headerName, hasHeader := model.OpenApiNodeToHeaderMapping[crdInfo.Name]
			if !hasHeader {
				continue
			}
			// If the parent is already declared in the URI as a path parameter,
			// or already present as a header on the spec, do not emit a duplicate.
			if paramExist(crdInfo.Name, resolvedParams) || headerParamExist(crdInfo.Name, restURI.Headers) || rawHeaderNameExist(headerName, restURI.Headers) {
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

		// Skip if parent is already in URI path or headers
		if !paramExist(crdInfo.Name, resolvedParams) && !headerParamExist(crdInfo.Name, restURI.Headers) {
			parameters = append(parameters, &openapi3.ParameterRef{
				Value: openapi3.NewQueryParameter(crdInfo.Name).
					WithRequired(true).
					WithSchema(openapi3.NewStringSchema()).
					WithDescription(description),
			})
		}
	}
	return
}

func constructUpdateParam() *openapi3.ParameterRef {
	return &openapi3.ParameterRef{
		Value: openapi3.NewQueryParameter("update_if_exists").
			WithRequired(false).
			WithSchema(openapi3.NewBoolSchema()).
			WithDescription("If set to false, disables update of preexisting object. Default value is true"),
	}
}

// resolveUriParams rewrites each URI token to its canonical groupKind via the
// per-URI PathParams map. Output shape matches r.FindAllStringSubmatch output
// (so [][1] is the resolved name). Tokens not present in PathParams are left
// as-is, preserving backward compatibility with URIs that use canonical form
// directly.
func resolveUriParams(rawParams [][]string, pathParams map[string]string) [][]string {
	if len(pathParams) == 0 {
		return rawParams
	}
	out := make([][]string, len(rawParams))
	for i, p := range rawParams {
		if len(p) < 2 {
			out[i] = p
			continue
		}
		if canonical, ok := pathParams[p[1]]; ok {
			resolved := make([]string, len(p))
			copy(resolved, p)
			resolved[1] = canonical
			out[i] = resolved
			continue
		}
		out[i] = p
	}
	return out
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

// rawHeaderNameExist reports whether any entry in headers (case-insensitively)
// matches the given raw HTTP header name. Used to avoid emitting a duplicate
// header parameter when the spec already declares the alias name explicitly.
func rawHeaderNameExist(headerName string, headers []string) bool {
	for _, h := range headers {
		if strings.EqualFold(h, headerName) {
			return true
		}
	}
	return false
}

func makeKey(crd, keyType string) string {
	return crd + "." + keyType
}

// ConstructNewURIs constructs the new URIs from ['status', 'children', 'links'] and store it in cache.
func ConstructNewURIs(n model.NexusAnnotation, urisMap map[string]model.RestURIInfo, newUris *[]nexus.RestURIs) {
	for _, uri := range n.NexusRestAPIGen.Uris {
		urisMap[uri.Uri] = model.RestURIInfo{
			TypeOfURI: model.DefaultURI,
		}
		for method := range uri.Methods {
			if method == http.MethodGet {
				statusUriPath := uri.Uri + "/status"
				addStatusUri(statusUriPath, model.StatusURI, urisMap, newUris)

				for _, c := range []map[string]model.NodeHelperChild{n.Children, n.Links} {
					processChildOrLink(c, uri, urisMap, newUris)
				}
			}
		}
	}
}

func processChildOrLink(nodes map[string]model.NodeHelperChild, uri nexus.RestURIs, urisMap map[string]model.RestURIInfo, newUris *[]nexus.RestURIs) {
	for _, n := range nodes {
		uriPath := uri.Uri + "/" + n.FieldName
		var t model.URIType
		if n.IsNamed {
			t = model.NamedLinkURI
		} else {
			t = model.SingleLinkURI
		}
		addUri(uriPath, t, urisMap, newUris)
	}
}

// addUri adds the uriPath </root/{orgchart.Root}/leader/{management.Leader}/HR> to the urisMap and to the uris list.
func addUri(uriPath string, typeOfUri model.URIType, urisMap map[string]model.RestURIInfo, uris *[]nexus.RestURIs) {
	newUri := nexus.RestURIs{
		Uri: uriPath,
		Methods: map[nexus.HTTPMethod]nexus.HTTPCodesResponse{
			http.MethodGet: nexus.DefaultHTTPGETResponses,
		},
	}
	urisMap[uriPath] = model.RestURIInfo{
		TypeOfURI: typeOfUri,
	}
	*uris = append(*uris, newUri)
}

func addStatusUri(uriPath string, typeOfUri model.URIType, urisMap map[string]model.RestURIInfo, uris *[]nexus.RestURIs) {
	newUri := nexus.RestURIs{
		Uri: uriPath,
		Methods: map[nexus.HTTPMethod]nexus.HTTPCodesResponse{
			http.MethodGet: nexus.DefaultHTTPGETResponses,
			// http.MethodPut: nexus.DefaultHTTPPUTResponses,
		},
	}
	urisMap[uriPath] = model.RestURIInfo{
		TypeOfURI: typeOfUri,
	}
	*uris = append(*uris, newUri)
}
