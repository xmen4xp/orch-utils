// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"net/http"
	"strings"

	"nexus-api-gw/pkg/model"
)

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
	kind := ""
	if len(nameParts) > 1 {
		kind = nameParts[1]
	}
	uriInfo, _ := model.GetURIInfo(uri)

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
// or "" if the URI ends in a path parameter "{...}". Used only for link
// traversal URIs where the trailing segment is the Go field name being
// navigated to.
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
