// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"net/http"

	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"

	"nexus/openapi-generator/pkg/model"
)

// ConstructNewURIs walks the children, links, and GET URIs of one
// node and emits the derived child/link/status URIs into `newUris`
// and `urisMap`. It is invoked once per node by main.go before any
// AddPath calls.
//
// Behaviour matches the previous in-line implementation exactly:
//
//   - For every URI declared on the node, register it in `urisMap`
//     as DefaultURI.
//   - For every GET URI, derive the corresponding `/status` URI
//     (StatusURI) and one URI per child/link (SingleLinkURI for
//     non-named children/links, NamedLinkURI for named ones).
//
// All derived URIs carry only GET to preserve the existing build-time
// emission shape; runtime PUT/PATCH on /status is added by api-gw
// separately.
func ConstructNewURIs(n model.NexusAnnotation, urisMap map[string]model.RestURIInfo, newUris *[]nexus.RestURIs) {
	for _, uri := range n.NexusRestAPIGen.Uris {
		urisMap[uri.Uri] = model.RestURIInfo{TypeOfURI: model.DefaultURI}
		for method := range uri.Methods {
			if method != http.MethodGet {
				continue
			}
			statusURIPath := uri.Uri + "/status"
			addStatusURI(statusURIPath, model.StatusURI, urisMap, newUris)

			for _, c := range []map[string]model.NodeHelperChild{n.Children, n.Links} {
				processChildOrLink(c, uri, urisMap, newUris)
			}
		}
	}
}

func processChildOrLink(nodes map[string]model.NodeHelperChild, uri nexus.RestURIs,
	urisMap map[string]model.RestURIInfo, newUris *[]nexus.RestURIs) {
	for _, n := range nodes {
		uriPath := uri.Uri + "/" + n.FieldName
		t := model.SingleLinkURI
		if n.IsNamed {
			t = model.NamedLinkURI
		}
		addURI(uriPath, t, urisMap, newUris)
	}
}

// addURI registers a single-link or named-link URI with GET only.
func addURI(uriPath string, typeOfURI model.URIType, urisMap map[string]model.RestURIInfo, uris *[]nexus.RestURIs) {
	newURI := nexus.RestURIs{
		Uri: uriPath,
		Methods: map[nexus.HTTPMethod]nexus.HTTPCodesResponse{
			http.MethodGet: nexus.DefaultHTTPGETResponses,
		},
	}
	urisMap[uriPath] = model.RestURIInfo{TypeOfURI: typeOfURI}
	*uris = append(*uris, newURI)
}

// addStatusURI registers a /status URI with GET only. PUT and PATCH
// are intentionally NOT emitted at build time to preserve the
// existing compiler spec shape — the runtime api-gw consumer adds
// those methods separately.
func addStatusURI(uriPath string, typeOfURI model.URIType, urisMap map[string]model.RestURIInfo, uris *[]nexus.RestURIs) {
	newURI := nexus.RestURIs{
		Uri: uriPath,
		Methods: map[nexus.HTTPMethod]nexus.HTTPCodesResponse{
			http.MethodGet: nexus.DefaultHTTPGETResponses,
		},
	}
	urisMap[uriPath] = model.RestURIInfo{TypeOfURI: typeOfURI}
	*uris = append(*uris, newURI)
}
