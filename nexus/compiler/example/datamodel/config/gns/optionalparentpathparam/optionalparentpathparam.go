// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package optionalparentpathparam

import (
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
)

var OptionalParentPathParamRestAPISpec = nexus.RestAPISpec{
	Uris: []nexus.RestURIs{
		{
			Uri:     "/v1/gns/{gns.Gns}/optionalparentpathparam/{optionalparentpathparam.OptionalParentPathParam}",
			Methods: nexus.DefaultHTTPMethodsResponses,
		},
		{
			Uri: "/v1/optionalparentpathparams",
			QueryParams: []string{
				"gns.Gns",
			},
			Methods: nexus.HTTPListResponse,
		},
	},
}

// nexus-rest-api-gen:OptionalParentPathParamRestAPISpec
type OptionalParentPathParam struct {
	nexus.Node
}
