// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// Package api is the api-gw adapter for nexus/openapi-builder. It is a
// thin shim that converts api-gw's `pkg/model` state into the builder's
// input types, runs Build, and exposes the resulting spec to the rest
// of the binary via the public `Schemas` map.
//
// All OpenAPI emission logic lives in `nexus/openapi-builder`. Bugs in
// that logic must be fixed there; this file should only ever evolve to
// reflect changes in the api-gw `pkg/model` schema or to add new
// adapter glue.
package api

import (
	"sync"

	"nexus-api-gw/pkg/config"
	"nexus-api-gw/pkg/model"
	"nexus-api-gw/pkg/utils"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"

	"nexus/openapi-builder/pkg/builder"
)

var (
	// Schemas is the per-datamodel OpenAPI spec served at
	// `/:datamodel/openapi.json`. It is rebuilt by `Recreate()` from
	// the current model state and read by the Echo handler.
	Schemas      = make(map[string]openapi3.T)
	schemasMutex = &sync.RWMutex{}

	// specBuilder is the single source of truth for OpenAPI emission.
	// Its state mirrors model.* maps; we Reset+repopulate on every
	// Recreate to guarantee that collision detection sees the full
	// current CRD set and there are no orphan tags or stale
	// operationIds.
	specBuilder = builder.New()

	appName = "nexus-api-gw-openapi"
	log     = logging.GetLogger(appName)
)

// New initializes (or resets) the `Schemas` entry for `datamodel` with
// an empty spec. Kept for backward compatibility with the existing
// ginkgo tests; production code should call Recreate.
//
// `ResetDomain` ensures consecutive `New(domain) -> AddPath(...)`
// sequences in tests start from a clean URI set. Without this the
// builder's merge-by-URI semantics would let URIs leak across tests.
func New(datamodel string) {
	title := "Nexus API GW APIs"
	if info, ok := model.DatamodelToDatamodelInfo[datamodel]; ok {
		title = info.Title
	}
	schemasMutex.Lock()
	defer schemasMutex.Unlock()
	specBuilder.ResetDomain(datamodel)
	specBuilder.SetDatamodelTitle(datamodel, builder.DatamodelTitle{Title: title})
	Schemas[datamodel] = *specBuilder.Build(datamodel)
}

// AddPath registers a URI under `datamodel` and refreshes
// `Schemas[datamodel]` accordingly. Kept for backward compatibility
// with the existing tests; production code goes through Recreate.
//
// The builder accumulates URIs per CRD across calls (merge-by-URI
// semantics in `AddCRDNode`), so callers pass only the URI being
// added — the builder retains every URI that was previously
// registered for the same crdType under this domain.
func AddPath(uri nexus.RestURIs, datamodel string) {
	crdType, _ := model.GetURIToCRDType(uri.Uri)
	crdInfo, _ := model.GetCRDTypeToNodeInfo(crdType)
	crdSpec, _ := model.GetCrdTypeToSpec(crdType)

	schemasMutex.Lock()
	defer schemasMutex.Unlock()

	specBuilder.SetOptions(currentOptions())
	specBuilder.AddCRDNode(datamodel, crdType,
		toBuilderNodeInfo(crdInfo, &crdSpec),
		toBuilderURIs([]nexus.RestURIs{uri}))

	Schemas[datamodel] = *specBuilder.Build(datamodel)
}

// AddExtensionPath installs an ExtensionRestAPI URI under `datamodel`
// and refreshes `Schemas[datamodel]`.
func AddExtensionPath(spec model.ExtensionRestAPISpec, datamodel string) {
	schemasMutex.Lock()
	defer schemasMutex.Unlock()

	specBuilder.SetOptions(currentOptions())
	specBuilder.AddExtension(datamodel, toBuilderExtension(spec))
	Schemas[datamodel] = *specBuilder.Build(datamodel)
}

// Recreate is the canonical full-rebuild entry point. It snapshots
// the current model state, resets the builder, re-adds every CRD
// node, extension, and datamodel title, then rebuilds the spec for
// every domain. The structural fix for orphan tags and over-qualified
// operationIds: collision detection always sees the full current set.
func Recreate() {
	log.Debug().Msg("Recreating openapi spec")

	allCRDURIs := model.GetAllCrdTypeToRestUris()
	allExtensions := model.GetAllExtensionRestAPISpecs()

	schemasMutex.Lock()
	defer schemasMutex.Unlock()

	specBuilder.Reset()
	specBuilder.SetOptions(currentOptions())

	domains := map[string]struct{}{}

	for crdType, uris := range allCRDURIs {
		info, _ := model.GetCRDTypeToNodeInfo(crdType)
		crdSpec, _ := model.GetCrdTypeToSpec(crdType)
		domain := utils.GetDatamodelName(crdType)
		domains[domain] = struct{}{}
		specBuilder.AddCRDNode(domain, crdType,
			toBuilderNodeInfo(info, &crdSpec),
			toBuilderURIs(uris))
	}

	for _, ext := range allExtensions {
		domain := ext.Datamodel
		if domain == "" {
			domain = "extension.nexus.com"
		}
		domains[domain] = struct{}{}
		specBuilder.AddExtension(domain, toBuilderExtension(ext))
	}

	// Carry over titles for every known datamodel.
	model.DatamodelToDatamodelInfoMutex.Lock()
	for dm, info := range model.DatamodelToDatamodelInfo {
		domains[dm] = struct{}{}
		specBuilder.SetDatamodelTitle(dm, builder.DatamodelTitle{Title: info.Title})
	}
	model.DatamodelToDatamodelInfoMutex.Unlock()

	// Rebuild every affected domain. Note: we rebuild domains seen
	// via CRDs/extensions/titles only. Pre-existing entries in
	// Schemas for domains that no longer have any input get cleared.
	newSchemas := make(map[string]openapi3.T, len(domains))
	for dm := range domains {
		newSchemas[dm] = *specBuilder.Build(dm)
	}
	Schemas = newSchemas
}

// RecreateExtension is preserved for backward compatibility with
// echo_server. It is a no-op now: `Recreate()` already includes
// extensions in the same snapshot, so a separate extension pass is
// no longer needed. Kept so the existing call sites compile
// unchanged.
func RecreateExtension() {
	log.Debug().Msg("RecreateExtension is a no-op; Recreate already includes extensions")
}

// DatamodelUpdateNotification listens on `model.DatamodelsChan` and
// refreshes the spec for the affected datamodel when a title arrives.
// Kept for compatibility with main.go which spawns it as a goroutine.
func DatamodelUpdateNotification() {
	log.Debug().Msg("Started datamodel update notification")
	for name := range model.DatamodelsChan {
		log.Debug().Msgf("Received datamodel update notification for %s", name)
		schemasMutex.Lock()

		model.DatamodelToDatamodelInfoMutex.Lock()
		info := model.DatamodelToDatamodelInfo[name]
		model.DatamodelToDatamodelInfoMutex.Unlock()

		specBuilder.SetDatamodelTitle(name, builder.DatamodelTitle{Title: info.Title})
		specBuilder.SetOptions(currentOptions())
		Schemas[name] = *specBuilder.Build(name)
		schemasMutex.Unlock()
		log.Info().Msgf("Updated title: %s for %s openapi spec", info.Title, name)
	}
}

// currentOptions reads the api-gw runtime config (header aliases) and
// returns a builder.Options snapshot. Called at every Add*/Recreate so
// config changes propagate without restart.
func currentOptions() builder.Options {
	opts := builder.Options{DefaultTitle: "Nexus API GW APIs"}
	if config.Cfg != nil && len(config.Cfg.HeaderAliases) > 0 {
		opts.HeaderAliases = make(map[string]string, len(config.Cfg.HeaderAliases))
		for k, v := range config.Cfg.HeaderAliases {
			opts.HeaderAliases[k] = v
		}
	}
	// api-gw runtime does not drop any parents (IgnoredParents stays
	// nil); the build-time openapi-generator is the only consumer
	// that uses IgnoredParents.
	return opts
}

// --- type conversion helpers ----------------------------------------------

func toBuilderNodeInfo(info model.NodeInfo, spec *apiextensionsv1.CustomResourceDefinitionSpec) builder.NodeInfo {
	out := builder.NodeInfo{
		Name:            info.Name,
		ParentHierarchy: info.ParentHierarchy,
		IsSingleton:     info.IsSingleton,
		Description:     info.Description,
		DeferredDelete:  info.DeferredDelete,
	}
	if info.Children != nil {
		out.Children = make(map[string]builder.NodeHelperChild, len(info.Children))
		for k, v := range info.Children {
			out.Children[k] = builder.NodeHelperChild{
				FieldName:    v.FieldName,
				FieldNameGvk: v.FieldNameGvk,
				IsNamed:      v.IsNamed,
			}
		}
	}
	if info.Links != nil {
		out.Links = make(map[string]builder.NodeHelperChild, len(info.Links))
		for k, v := range info.Links {
			out.Links[k] = builder.NodeHelperChild{
				FieldName:    v.FieldName,
				FieldNameGvk: v.FieldNameGvk,
				IsNamed:      v.IsNamed,
			}
		}
	}
	// Extract the OpenAPI v3 schema from the first declared CRD
	// version, matching the existing emission code's lookup path
	// (`crdSpec.Versions[0].Schema.OpenAPIV3Schema`). The builder
	// tolerates a nil Schema (no-op for component emission).
	if spec != nil && len(spec.Versions) > 0 &&
		spec.Versions[0].Schema != nil &&
		spec.Versions[0].Schema.OpenAPIV3Schema != nil {
		out.Schema = spec.Versions[0].Schema.OpenAPIV3Schema
	}
	return out
}

func toBuilderURIs(uris []nexus.RestURIs) []builder.RestURIs {
	if len(uris) == 0 {
		return nil
	}
	out := make([]builder.RestURIs, 0, len(uris))
	for _, u := range uris {
		methods := make(map[string]struct{}, len(u.Methods))
		for m := range u.Methods {
			methods[string(m)] = struct{}{}
		}
		typeOfURI := builder.DefaultURI
		if info, ok := model.GetURIInfo(u.Uri); ok {
			typeOfURI = builder.URIType(info.TypeOfURI)
		}
		out = append(out, builder.RestURIs{
			URI:        u.Uri,
			Methods:    methods,
			TypeOfURI:  typeOfURI,
			PathParams: u.PathParams,
			Headers:    u.Headers,
		})
	}
	return out
}

func toBuilderExtension(spec model.ExtensionRestAPISpec) builder.ExtensionSpec {
	return builder.ExtensionSpec{
		URI:             spec.URI,
		Methods:         spec.Methods,
		OpenAPIPathSpec: spec.OpenAPIPathSpec,
	}
}

// LoadCombinedSpec loads a statically-mounted combined OpenAPI document
// (produced offline and shipped with the api-gw container image) and
// registers it under the `edge-orchestrator.intel.com` datamodel key in
// the `Schemas` map. Called once from main.go at startup; the document
// is then served via the standard /:datamodel/openapi.json route. The
// builder pipeline is intentionally bypassed here because the combined
// spec is sourced from disk rather than from the in-cluster model
// state.
func LoadCombinedSpec() {
	const specFilePath = "/static/openapispecs/combined/combined_spec.yaml"

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(specFilePath)
	if err != nil {
		log.Error().Msgf("Failed to load combined OpenAPI spec from %s: %v", specFilePath, err)
		return
	}

	schemasMutex.Lock()
	Schemas["edge-orchestrator.intel.com"] = *doc
	schemasMutex.Unlock()

	log.Info().Msgf("Loaded combined OpenAPI spec: %s", doc.Info.Title)
}
