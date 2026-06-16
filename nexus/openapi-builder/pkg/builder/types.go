// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// Package builder is the single source of truth for OpenAPI spec emission
// from the Nexus datamodel. It is a pure library — no I/O, no k8s client,
// no HTTP framework — consumed by both `nexus/openapi-generator` (build
// time) and `nexus/api-gw` (runtime).
//
// Inputs are fed via Add* methods. Spec emission happens in `Build(domain)`,
// which takes a snapshot of the current input set, runs collision detection
// scoped to that domain, and returns a freshly-assembled `*openapi3.T`.
// Build is deterministic for any input set, in any insertion order.
package builder

import apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

// URIType classifies a REST URI for the purposes of operationId emission
// and component-schema selection.
type URIType int

const (
	// DefaultURI is a CRUD/list endpoint on the node itself.
	DefaultURI URIType = iota
	// SingleLinkURI follows a singular link annotation to a referenced node.
	SingleLinkURI
	// NamedLinkURI follows a named link (map) annotation to a referenced node.
	NamedLinkURI
	// StatusURI targets the `/status` subresource.
	StatusURI
)

// NodeHelperChild describes a child or link declared in a parent's nexus
// annotation. FieldName is the Go struct field name (e.g. "Org"); IsNamed
// distinguishes singular links/children from named (map) ones.
type NodeHelperChild struct {
	FieldName    string
	FieldNameGvk string
	IsNamed      bool
}

// NodeInfo carries the per-CRD metadata that the builder needs to emit
// tags, operationIds, and parameters for one node's URIs. It is a flat
// snapshot — the builder does not navigate from one NodeInfo to another.
type NodeInfo struct {
	// Name is the datamodel-qualified node name in `package.Kind` form
	// (e.g. "orgs.Org", "aislice.AISlice"). The Kind portion drives tag
	// and operationId generation; the package portion is used to
	// disambiguate when two CRDs in the same datamodel share a Kind.
	Name string

	// ParentHierarchy lists the ParentHierarchy[].Name values from the
	// nexus annotation, used to dedupe URI parameters inherited from
	// ancestors.
	ParentHierarchy []string

	// Children and Links carry the nexus child/link declarations. Used
	// when emitting SingleLink/NamedLink URIs to determine the target
	// node's Kind for operationId construction.
	Children map[string]NodeHelperChild
	Links    map[string]NodeHelperChild

	// IsSingleton is true when the node is a singleton (no name path
	// parameter on its leaf URI).
	IsSingleton bool

	// Description is the nexus-description text propagated to the
	// emitted OpenAPI tag and the schema components.
	Description string

	// DeferredDelete signals that DELETE on this node returns 202 with
	// a status body rather than 204.
	DeferredDelete bool

	// Schema is the CRD's openapi v3 JSON schema, used to emit
	// component schemas (spec/status/get/list/post) and request/response
	// bodies. May be nil for synthetic nodes — in that case no
	// components are emitted and only generic responses are referenced
	// from the operations.
	Schema *apiextensionsv1.JSONSchemaProps
}

// RestURIs is one URI declared on a node, with the HTTP methods exposed on
// it and the URI's classification (DefaultURI / StatusURI / ...).
//
// Methods is intentionally just a set of HTTP method strings. The
// per-method response-code map carried by upstream `nexus.RestURIs` is
// not used during OpenAPI emission — the builder emits a fixed response
// envelope per (method, URIType) combination.
type RestURIs struct {
	URI       string
	Methods   map[string]struct{}
	TypeOfURI URIType

	// PathParams maps URI-token aliases to canonical groupKinds (e.g.
	// "datacenter" -> "datacenters.DataCenters"). Empty map means tokens
	// in the URI are already canonical.
	PathParams map[string]string

	// Headers lists explicit header parameters declared on the URI by
	// the datamodel author. Distinct from `Options.HeaderAliases` which
	// promotes hierarchy parents into header parameters.
	Headers []string
}

// ExtensionSpec describes a declarative ExtensionRestAPI URI (one entry
// per CR). The OpenAPIPathSpec is a raw YAML fragment describing the
// path's OpenAPI3 operations; the builder parses and inlines it.
type ExtensionSpec struct {
	URI             string
	Methods         []string
	OpenAPIPathSpec string
	Description     string
	// Tag overrides the auto-derived tag name. Empty = derive from URI.
	Tag string
}

// DatamodelTitle is the optional spec.info.title override for a datamodel.
// When unset, the builder uses the default "Nexus API GW APIs".
type DatamodelTitle struct {
	Title string
}
