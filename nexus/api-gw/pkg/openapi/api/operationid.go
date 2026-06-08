// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"net/http"
	"strings"

	"nexus-api-gw/pkg/model"
)

// getOperationID returns an OpenAPI operationId of the form
// "<httpVerb><Kind>" (singular) or "<httpVerb><Plural>" (LIST) for
// datamodel-derived routes. Sub-URIs get suffixes:
//   - StatusURI:       <verb><Kind>Status
//   - SingleLink/Named: get<ParentKind><FieldName> (last URI segment is the Go field name)
//
// Verb map: LIST->get, GET->get, PUT->put, PATCH->patch, DELETE->delete.
// Plural is derived from the LIST URI's last static segment so that
// developer-chosen plurals (e.g. Org -> "organizations") are preserved while
// the singular Kind casing (e.g. "AISlice") is retained where it overlaps.
func getOperationID(method, uri string, crdInfo model.NodeInfo) string {
	nameParts := strings.Split(crdInfo.Name, ".")
	kind := ""
	if len(nameParts) > 1 {
		kind = nameParts[1]
	}
	uriInfo, _ := model.GetURIInfo(uri)

	switch method {
	case "LIST":
		return "get" + derivePlural(kind, uri)
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

// derivePlural produces a PascalCase plural noun for a Kind by combining the
// Kind's original casing with the URI's plural segment. Rules:
//  1. If the URI plural equals lowercase(kind), use kind verbatim (already plural).
//  2. If the URI plural starts with lowercase(kind), splice kind onto the tail.
//  3. Otherwise, Title-case the URI plural segment.
//
// Falls back to kind+"s" when the URI has no usable plural segment.
func derivePlural(kind, uri string) string {
	plural := lastStaticSegment(uri)
	if plural == "" {
		return kind + "s"
	}
	lowerKind := strings.ToLower(kind)
	if plural == lowerKind {
		return kind
	}
	if strings.HasPrefix(plural, lowerKind) {
		return kind + plural[len(lowerKind):]
	}
	return strings.ToUpper(plural[:1]) + plural[1:]
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
