// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// Package api is the openapi-generator adapter for
// `nexus/openapi-builder`. It owns build-time URI discovery
// (ConstructNewURIs and friends) and converts the resulting model
// state into builder inputs, then exposes the built spec via the
// public `Schemas` map for `main.go` to serialise.
//
// All OpenAPI emission logic lives in `nexus/openapi-builder`. Bugs
// in that logic must be fixed there.
package api

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"

	"nexus/openapi-builder/pkg/builder"
	"nexus/openapi-generator/pkg/model"
)

var (
	// Schemas holds the per-datamodel OpenAPI spec rendered by Build.
	// It is rebuilt by every Add* call so `main.go` can read a current
	// view immediately after writing the last URI.
	Schemas      = make(map[string]openapi3.T)
	schemasMutex sync.Mutex

	// specBuilder is the single source of truth for OpenAPI emission.
	specBuilder = builder.New()
)

// New initializes (or resets) the spec for `datamodel`. Subsequent
// AddPath / AddExtensionPath calls extend that spec.
func New(datamodel string) {
	title := "Nexus API GW APIs"
	if info, ok := model.DatamodelToDatamodelInfo[datamodel]; ok && info.Title != "" {
		title = info.Title
	}
	schemasMutex.Lock()
	defer schemasMutex.Unlock()
	specBuilder.SetOptions(currentOptions())
	specBuilder.ResetDomain(datamodel)
	specBuilder.SetDatamodelTitle(datamodel, builder.DatamodelTitle{Title: title})
	Schemas[datamodel] = *specBuilder.Build(datamodel)
}

// AddPath registers a single URI under `datamodel` and refreshes
// `Schemas[datamodel]` to reflect the addition. Tests and main.go
// call this once per URI from a deterministic ordered list, so the
// resulting spec is reproducible.
func AddPath(uri nexus.RestURIs, datamodel string) {
	crdType, ok := model.UriToCRDType[uri.Uri]
	if !ok {
		log.Warnf("openapi-generator: no CRD type registered for URI %q", uri.Uri)
		return
	}
	crdInfo := model.CrdTypeToNodeInfo[crdType]
	crdSpec := model.CrdTypeToSpec[crdType]

	schemasMutex.Lock()
	defer schemasMutex.Unlock()
	specBuilder.SetOptions(currentOptions())

	// Aggregate every URI that has been registered against this CRD
	// type so far so the builder sees the complete URI list per node.
	// Without this, calling AddCRDNode for one URI would replace the
	// builder's prior URI list for the same CRD.
	uris := uriSetFor(crdType)
	uris = appendURIIfMissing(uris, uri)

	specBuilder.AddCRDNode(datamodel, crdType,
		toBuilderNodeInfo(crdInfo, crdSpec), toBuilderURIs(uris))
	Schemas[datamodel] = *specBuilder.Build(datamodel)
}

// AddExtensionPath merges an ExtensionRestAPI CR manifest into the
// spec for the given datamodel. Files that are not ExtensionRestAPI
// manifests are silently skipped so the caller can pass a mixed
// directory.
func AddExtensionPath(fileBytes []byte, datamodel string) {
	ext, ok, err := parseExtensionEnvelope(fileBytes)
	if err != nil {
		log.Warnf("extension: %v", err)
		return
	}
	if !ok {
		return
	}

	schemasMutex.Lock()
	defer schemasMutex.Unlock()
	specBuilder.SetOptions(currentOptions())
	specBuilder.AddExtension(datamodel, ext)
	Schemas[datamodel] = *specBuilder.Build(datamodel)
	log.Infof("adding extension %s path to %s", ext.URI, datamodel)
}

// currentOptions reads build-time config (ignoredParentPathParams +
// nodeToHeaderMapping) and converts it into builder.Options.
func currentOptions() builder.Options {
	opts := builder.Options{
		DefaultTitle: "Nexus API GW APIs",
		Servers: openapi3.Servers{
			&openapi3.Server{Description: "API Gateway", URL: "/"},
			&openapi3.Server{Description: "Local", URL: "http://localhost:5000"},
			&openapi3.Server{Description: "Local SSL", URL: "https://localhost:5443"},
		},
	}
	if len(model.OpenApiIgnoredParentPathParams) > 0 {
		opts.IgnoredParents = make(map[string]struct{}, len(model.OpenApiIgnoredParentPathParams))
		for k := range model.OpenApiIgnoredParentPathParams {
			opts.IgnoredParents[k] = struct{}{}
		}
	}
	if len(model.OpenApiNodeToHeaderMapping) > 0 {
		opts.HeaderAliases = make(map[string]string, len(model.OpenApiNodeToHeaderMapping))
		for k, v := range model.OpenApiNodeToHeaderMapping {
			opts.HeaderAliases[k] = v
		}
	}
	return opts
}

// uriSetFor returns every nexus.RestURIs whose URI maps to `crdType`
// in `model.UriToCRDType`. Used so AddPath can pass the full URI
// list to the builder when the caller has not pre-populated a CRD
// type→URIs map.
func uriSetFor(crdType string) []nexus.RestURIs {
	var out []nexus.RestURIs
	for uriPath, ct := range model.UriToCRDType {
		if ct != crdType {
			continue
		}
		typeOfURI := model.DefaultURI
		if info, ok := model.UriToUriInfo[uriPath]; ok {
			typeOfURI = info.TypeOfURI
		}
		out = append(out, restURIWithMethods(uriPath, typeOfURI))
	}
	return out
}

// restURIWithMethods returns a nexus.RestURIs synthesised from a URI
// path + URI type. Methods default to GET unless the URI is a /status
// path (in which case GET + PUT, mirroring addStatusUri's emission).
func restURIWithMethods(uriPath string, typeOfURI model.URIType) nexus.RestURIs {
	out := nexus.RestURIs{Uri: uriPath, Methods: map[nexus.HTTPMethod]nexus.HTTPCodesResponse{}}
	if typeOfURI == model.StatusURI {
		out.Methods[http.MethodGet] = nexus.DefaultHTTPGETResponses
		// PUT/PATCH on /status are intentionally NOT auto-added at
		// build-time to preserve the existing compiler spec shape.
		// The runtime api-gw adds them via separate logic.
	} else {
		out.Methods[http.MethodGet] = nexus.DefaultHTTPGETResponses
	}
	return out
}

// appendURIIfMissing appends `uri` to `uris` unless an entry with the
// same URI string is already present. Methods/responses from the
// caller take precedence so the test-passed URI's methods land
// in the builder verbatim.
func appendURIIfMissing(uris []nexus.RestURIs, uri nexus.RestURIs) []nexus.RestURIs {
	for i, existing := range uris {
		if existing.Uri == uri.Uri {
			uris[i] = uri
			return uris
		}
	}
	return append(uris, uri)
}

// parseExtensionEnvelope decodes an ExtensionRestAPI CR YAML payload
// into a builder.ExtensionSpec. Returns ok=false (no error) when the
// payload's `kind` is not `ExtensionRestAPI` so the caller can
// safely iterate over a mixed directory.
func parseExtensionEnvelope(fileBytes []byte) (builder.ExtensionSpec, bool, error) {
	// Decode via JSON for compatibility with both YAML and JSON
	// inputs. ghodss/yaml is the canonical translator used elsewhere
	// in this repo.
	var env struct {
		Kind string `json:"kind"`
		Spec struct {
			URI             string   `json:"uri"`
			OpenAPIPathSpec string   `json:"openAPIPathSpec"`
			Methods         []string `json:"methods"`
			Description     string   `json:"description"`
		} `json:"spec"`
	}
	asJSON, err := ghYAMLToJSON(fileBytes)
	if err != nil {
		return builder.ExtensionSpec{}, false, err
	}
	if err := json.Unmarshal(asJSON, &env); err != nil {
		return builder.ExtensionSpec{}, false, err
	}
	if env.Kind != "ExtensionRestAPI" {
		return builder.ExtensionSpec{}, false, nil
	}
	if env.Spec.URI == "" || env.Spec.OpenAPIPathSpec == "" {
		log.Warnf("extension: skipping %q — missing spec.uri or spec.openAPIPathSpec", env.Spec.URI)
		return builder.ExtensionSpec{}, false, nil
	}
	return builder.ExtensionSpec{
		URI:             env.Spec.URI,
		Methods:         env.Spec.Methods,
		OpenAPIPathSpec: env.Spec.OpenAPIPathSpec,
		Description:     env.Spec.Description,
	}, true, nil
}
