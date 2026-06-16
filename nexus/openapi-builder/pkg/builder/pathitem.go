// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// buildPathItem assembles the openapi3.PathItem for one URI. It emits
// one Operation per method present in `uri.Methods`. The operation's
// Tag is computed by qualifiedKind using the per-domain `counts`; the
// operationId is computed by GetOperationID for parity.
//
// `params` is the pre-built parameter list shared by all methods on
// this URI. Methods that need extra parameters (PUT on a non-status
// URI appends `update_if_exists`) extend `params` locally.
func buildPathItem(uri RestURIs, info NodeInfo, params []*openapi3.ParameterRef, counts KindCounts) *openapi3.PathItem {
	pathItem := &openapi3.PathItem{}
	nameParts := strings.Split(info.Name, ".")
	tag := qualifiedKind(nameParts, counts)

	for method := range uri.Methods {
		opID := GetOperationID(method, uri.URI, info, uri.TypeOfURI, counts)
		switch method {
		case "LIST":
			pathItem.Get = newListOperation(opID, tag, params, info)
		case http.MethodGet:
			pathItem.Get = newGetOperation(opID, tag, params, info, uri.TypeOfURI)
		case http.MethodPut:
			pathItem.Put = newPutOperation(opID, tag, params, info, uri.TypeOfURI)
		case http.MethodPatch:
			pathItem.Patch = newPatchOperation(opID, tag, params, info, uri.TypeOfURI)
		case http.MethodDelete:
			pathItem.Delete = newDeleteOperation(opID, tag, params)
		}
	}

	return pathItem
}

func newListOperation(opID, tag string, params []*openapi3.ParameterRef, info NodeInfo) *openapi3.Operation {
	op := &openapi3.Operation{
		OperationID: opID,
		Tags:        []string{tag},
		Parameters:  params,
		Responses:   openapi3.NewResponses(),
	}
	op.Responses.Set("200", &openapi3.ResponseRef{
		Ref: "#/components/responses/List" + info.Name,
	})
	return op
}

func newGetOperation(opID, tag string, params []*openapi3.ParameterRef, info NodeInfo, uriType URIType) *openapi3.Operation {
	op := &openapi3.Operation{
		OperationID: opID,
		Tags:        []string{tag},
		Parameters:  params,
		Responses:   openapi3.NewResponses(),
	}
	switch uriType {
	case StatusURI:
		op.Responses.Set("200", &openapi3.ResponseRef{
			Ref: "#/components/responses/Get" + info.Name + ".Status",
		})
	case SingleLinkURI:
		op.Responses.Set("200", &openapi3.ResponseRef{
			Ref: "#/components/responses/Get" + info.Name + ".SingleLink",
		})
	case NamedLinkURI:
		op.Responses.Set("200", &openapi3.ResponseRef{
			Ref: "#/components/responses/Get" + info.Name + ".NamedLink",
		})
	default:
		op.Responses.Set("200", &openapi3.ResponseRef{
			Ref: "#/components/responses/Get" + info.Name,
		})
	}
	return op
}

func newPutOperation(opID, tag string, params []*openapi3.ParameterRef, info NodeInfo, uriType URIType) *openapi3.Operation {
	op := &openapi3.Operation{
		OperationID: opID,
		Tags:        []string{tag},
		Responses:   openapi3.NewResponses(),
	}
	if uriType == StatusURI {
		op.RequestBody = &openapi3.RequestBodyRef{
			Ref: "#/components/requestBodies/Create" + info.Name + ".Status",
		}
		op.Responses.Set("200", &openapi3.ResponseRef{
			Ref: "#/components/responses/DefaultResponse",
		})
		op.Parameters = params
		return op
	}
	// Default URI: extend params with the update_if_exists query param.
	putParams := make([]*openapi3.ParameterRef, 0, len(params)+1)
	putParams = append(putParams, params...)
	putParams = append(putParams, constructUpdateParam())
	op.Parameters = putParams
	op.RequestBody = &openapi3.RequestBodyRef{
		Ref: "#/components/requestBodies/Create" + info.Name,
	}
	op.Responses.Set("200", &openapi3.ResponseRef{
		Ref: "#/components/responses/DefaultResponse",
	})
	return op
}

func newPatchOperation(opID, tag string, params []*openapi3.ParameterRef, info NodeInfo, uriType URIType) *openapi3.Operation {
	op := &openapi3.Operation{
		OperationID: opID,
		Tags:        []string{tag},
		Parameters:  params,
		Responses:   openapi3.NewResponses(),
	}
	op.Responses.Set("200", &openapi3.ResponseRef{
		Ref: "#/components/responses/DefaultResponse",
	})
	op.Responses.Set("404", &openapi3.ResponseRef{
		Ref: "#/components/responses/NotFoundResponse",
	})
	if uriType == StatusURI {
		op.RequestBody = &openapi3.RequestBodyRef{
			Ref: "#/components/requestBodies/Create" + info.Name + ".Status",
		}
	} else {
		op.RequestBody = &openapi3.RequestBodyRef{
			Ref: "#/components/requestBodies/Create" + info.Name,
		}
	}
	return op
}

func newDeleteOperation(opID, tag string, params []*openapi3.ParameterRef) *openapi3.Operation {
	op := &openapi3.Operation{
		OperationID: opID,
		Tags:        []string{tag},
		Parameters:  params,
		Responses:   openapi3.NewResponses(),
	}
	op.Responses.Set("200", &openapi3.ResponseRef{
		Value: openapi3.NewResponse().WithDescription("No content"),
	})
	return op
}

// hasTag returns true if a tag with the given name already exists in
// the slice. Used by Build to dedup tag registrations across multiple
// URIs that share a node.
func hasTag(tags openapi3.Tags, name string) bool {
	for _, t := range tags {
		if t.Name == name {
			return true
		}
	}
	return false
}
