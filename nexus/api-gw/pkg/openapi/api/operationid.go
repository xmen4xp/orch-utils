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
	kind := qualifiedKind(nameParts)
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

// qualifiedKind returns the Kind portion of a CRD's NodeInfo.Name
// (the part after the first dot), prefixed with a Title-cased package
// segment iff more than one CRD shares the same Kind. Returns "" when
// nameParts has no Kind portion.
//
// Examples:
//
//	["orgs", "Org"]                       (no collision) -> "Org"
//	["aislice", "AISlice"]                (collision)    -> "AisliceAISlice"
//	["discoveredaislice", "AISlice"]      (collision)    -> "DiscoveredaisliceAISlice"
func qualifiedKind(nameParts []string) string {
	if len(nameParts) < 2 {
		return ""
	}
	kind := nameParts[1]
	if model.IsCollidingKind(kind) {
		return titleFirst(nameParts[0]) + kind
	}
	return kind
}

// titleFirst upper-cases the first rune of s. Used instead of the
// deprecated strings.Title for our simple ASCII package-name inputs.
func titleFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
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
