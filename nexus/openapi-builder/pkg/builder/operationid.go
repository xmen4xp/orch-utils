// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"net/http"
	"strings"
)

// GetOperationID returns the OpenAPI operationId for a datamodel-derived
// route. It is exported so that adapters/tests can exercise it directly;
// the builder uses it internally during Build.
//
// Scheme:
//
//   - LIST                      -> list<KindPlural>      (Kind-derived plural; if Kind already ends in "s", use Kind verbatim)
//   - GET item                  -> get<Kind>
//   - PUT/PATCH/DELETE item     -> <verb><Kind>
//   - StatusURI (any verb)      -> <verb><Kind>Status
//   - SingleLink / NamedLink    -> get<ParentKind><FieldName>  (FieldName = last URI segment)
//
// `counts` is the per-domain Kind histogram used to disambiguate
// colliding Kinds via qualifiedKind. `info` is the NodeInfo of the CRD
// the URI belongs to. `uri` is the templated URI string. `uriType`
// classifies the URI; pass DefaultURI for plain CRUD endpoints.
func GetOperationID(method, uri string, info NodeInfo, uriType URIType, counts KindCounts) string {
	nameParts := strings.Split(info.Name, ".")
	kind := qualifiedKind(nameParts, counts)

	switch method {
	case "LIST":
		return "list" + kindPlural(kind)
	case http.MethodGet:
		switch uriType {
		case StatusURI:
			return "get" + kind + "Status"
		case SingleLinkURI, NamedLinkURI:
			return "get" + kind + lastStaticSegment(uri)
		}
		return "get" + kind
	case http.MethodPut:
		if uriType == StatusURI {
			return "put" + kind + "Status"
		}
		return "put" + kind
	case http.MethodPatch:
		if uriType == StatusURI {
			return "patch" + kind + "Status"
		}
		return "patch" + kind
	case http.MethodDelete:
		return "delete" + kind
	}
	return strings.ToLower(method) + kind
}
