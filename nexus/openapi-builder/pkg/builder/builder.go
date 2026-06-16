// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"sort"
	"strings"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
)

// Options configures the builder's per-spec behaviour. Both fields are
// optional; the zero value of Options is a valid configuration.
type Options struct {
	// HeaderAliases maps node names (e.g. "orgs.Org") to HTTP header
	// names (e.g. "x-org-id"). When set, parent path-hierarchy params
	// for these nodes are emitted as required header parameters
	// instead of query parameters.
	HeaderAliases map[string]string

	// IgnoredParents lists node names whose path-hierarchy params are
	// dropped from the spec entirely (typically because the parent's
	// identity is conveyed via tenant context rather than the URL).
	// If a node appears in BOTH IgnoredParents and HeaderAliases, it
	// is emitted as a header parameter (the IgnoredParents drop is
	// overridden by the HeaderAlias promotion).
	IgnoredParents map[string]struct{}

	// DefaultTitle is the spec.info.title used when a domain has no
	// title configured via SetDatamodelTitle. Defaults to
	// "Nexus API GW APIs" when empty.
	DefaultTitle string

	// Servers is the list of OpenAPI servers added to every emitted
	// spec. When empty, a single "/" server is emitted as the default.
	Servers openapi3.Servers
}

// Logger is the minimal logging interface the builder uses for non-fatal
// emission warnings (e.g. malformed extension YAML). Adapters supply
// their own logger to avoid pinning a logging library into this module.
//
// A nil logger silently discards messages.
type Logger interface {
	Warnf(format string, args ...interface{})
}

type noopLogger struct{}

func (noopLogger) Warnf(string, ...interface{}) {}

// SpecBuilder is the entry point. Construct one per process via New(),
// feed inputs via Add* methods, and call Build(domain) whenever a
// fresh spec is needed. SpecBuilder is goroutine-safe.
type SpecBuilder struct {
	mu      sync.RWMutex
	opts    Options
	logger  Logger
	domains map[string]*domainInputs
}

type domainInputs struct {
	title      string
	nodes      map[string]nodeRecord    // crdType → node + URIs
	extensions map[string]ExtensionSpec // URI → extension
}

type nodeRecord struct {
	info NodeInfo
	uris []RestURIs
}

// New returns a SpecBuilder with default options (no header aliases,
// no ignored parents, default title, single "/" server) and a noop
// logger.
func New() *SpecBuilder {
	return NewWithOptions(Options{})
}

// NewWithOptions returns a SpecBuilder configured with the given
// Options and a noop logger.
func NewWithOptions(opts Options) *SpecBuilder {
	return &SpecBuilder{
		opts:    opts,
		logger:  noopLogger{},
		domains: map[string]*domainInputs{},
	}
}

// SetLogger replaces the builder's logger. Safe to call at any time.
func (b *SpecBuilder) SetLogger(l Logger) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if l == nil {
		l = noopLogger{}
	}
	b.logger = l
}

// SetOptions replaces the builder's emission options. Affects future
// Build() calls only; in-flight Build() calls retain their snapshot
// options.
func (b *SpecBuilder) SetOptions(opts Options) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.opts = opts
}

// AddCRDNode registers a CRD node's metadata and URI list under the
// given domain. The URI list is **merged by URI string** with any
// previously registered URIs for the same (domain, crdType): a URI
// already present is replaced (last-write-wins), a URI not present
// is appended. The latest call's `info` always overwrites the prior
// info.
//
// This shape supports two consumption patterns with no caller-side
// state:
//   - Snapshot adapters (api-gw) call once per CRD with the full URI
//     list after a Reset/ResetDomain — no duplicates, so merge is a
//     plain assignment.
//   - Incremental adapters (openapi-generator) call once per URI as
//     they walk the input. The builder accumulates URIs across calls
//     and preserves each URI's original methods (GET / LIST / PUT /
//     PATCH / DELETE) verbatim — never re-synthesised by the caller.
func (b *SpecBuilder) AddCRDNode(domain, crdType string, info NodeInfo, uris []RestURIs) {
	b.mu.Lock()
	defer b.mu.Unlock()
	d := b.getOrCreateDomainLocked(domain)
	// mergeURIsByURI handles both first-call (prev.uris is nil) and
	// subsequent-call cases identically — a nil base just means the
	// result is a copy of `uris`.
	d.nodes[crdType] = nodeRecord{
		info: info,
		uris: mergeURIsByURI(d.nodes[crdType].uris, uris),
	}
}

// mergeURIsByURI returns base with each entry of incoming applied:
// an entry whose URI matches an existing one replaces that entry in
// place (preserving order); a new URI is appended at the end. The
// returned slice is a fresh allocation — callers may mutate freely.
func mergeURIsByURI(base, incoming []RestURIs) []RestURIs {
	out := append([]RestURIs(nil), base...)
	for _, u := range incoming {
		replaced := false
		for i, e := range out {
			if e.URI == u.URI {
				out[i] = u
				replaced = true
				break
			}
		}
		if !replaced {
			out = append(out, u)
		}
	}
	return out
}

// RemoveCRDNode removes a CRD node from the given domain. No-op if
// the node is not registered.
func (b *SpecBuilder) RemoveCRDNode(domain, crdType string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if d, ok := b.domains[domain]; ok {
		delete(d.nodes, crdType)
	}
}

// AddExtension registers (or replaces) a declarative extension URI
// under the given domain. Idempotent on (domain, URI).
func (b *SpecBuilder) AddExtension(domain string, ext ExtensionSpec) {
	b.mu.Lock()
	defer b.mu.Unlock()
	d := b.getOrCreateDomainLocked(domain)
	d.extensions[ext.URI] = ext
}

// RemoveExtension removes a declarative extension URI from the given
// domain. No-op if not registered.
func (b *SpecBuilder) RemoveExtension(domain, uri string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if d, ok := b.domains[domain]; ok {
		delete(d.extensions, uri)
	}
}

// SetDatamodelTitle overrides the spec.info.title for the given
// domain. Empty title reverts to Options.DefaultTitle.
func (b *SpecBuilder) SetDatamodelTitle(domain string, t DatamodelTitle) {
	b.mu.Lock()
	defer b.mu.Unlock()
	d := b.getOrCreateDomainLocked(domain)
	d.title = t.Title
}

// Reset clears all input state across every domain.
func (b *SpecBuilder) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.domains = map[string]*domainInputs{}
}

// ResetDomain clears all input state for a single domain.
func (b *SpecBuilder) ResetDomain(domain string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.domains, domain)
}

// Domains returns the sorted list of known domains. Useful for
// adapters that iterate and emit a spec per domain.
func (b *SpecBuilder) Domains() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]string, 0, len(b.domains))
	for d := range b.domains {
		out = append(out, d)
	}
	sort.Strings(out)
	return out
}

// getOrCreateDomainLocked must be called with b.mu held.
func (b *SpecBuilder) getOrCreateDomainLocked(domain string) *domainInputs {
	d, ok := b.domains[domain]
	if !ok {
		d = &domainInputs{
			nodes:      map[string]nodeRecord{},
			extensions: map[string]ExtensionSpec{},
		}
		b.domains[domain] = d
	}
	return d
}

// Build returns a freshly-assembled OpenAPI spec for the given
// domain. Snapshot semantics: collision detection runs over the
// CURRENT input set for `domain` only, so the spec is internally
// consistent regardless of how inputs accumulated.
//
// Build returns an empty (but valid) spec for an unknown domain;
// callers can distinguish via Domains() if needed.
func (b *SpecBuilder) Build(domain string) *openapi3.T {
	// Snapshot inputs under RLock so we can release the lock before
	// doing the (potentially expensive) emission work.
	b.mu.RLock()
	opts := b.opts
	logger := b.logger
	title := opts.DefaultTitle
	if title == "" {
		title = "Nexus API GW APIs"
	}

	// keyedNodeRecord pairs a snapshot nodeRecord with its CRD-type
	// map key so emission can index by CRD type without re-locking.
	type keyedNodeRecord struct {
		crdType string
		rec     nodeRecord
	}

	var (
		nodes      []keyedNodeRecord
		extensions []ExtensionSpec
	)
	if d, ok := b.domains[domain]; ok {
		if d.title != "" {
			title = d.title
		}
		nodes = make([]keyedNodeRecord, 0, len(d.nodes))
		for crdType, n := range d.nodes {
			nodes = append(nodes, keyedNodeRecord{crdType: crdType, rec: n})
		}
		extensions = make([]ExtensionSpec, 0, len(d.extensions))
		for _, e := range d.extensions {
			extensions = append(extensions, e)
		}
	}
	b.mu.RUnlock()

	// Deterministic order — emission depends only on the input set,
	// not the map iteration order.
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].rec.info.Name < nodes[j].rec.info.Name })
	sort.Slice(extensions, func(i, j int) bool { return extensions[i].URI < extensions[j].URI })

	// Snapshot indices used during emission. All three are derived
	// from the same single-window snapshot above — they cannot drift
	// out of sync with the collision histogram.
	nodesByCRDType := make(map[string]NodeInfo, len(nodes))
	nodesByName := make(map[string]NodeInfo, len(nodes))
	infosForCounts := make([]NodeInfo, 0, len(nodes))
	for _, n := range nodes {
		nodesByCRDType[n.crdType] = n.rec.info
		nodesByName[n.rec.info.Name] = n.rec.info
		infosForCounts = append(infosForCounts, n.rec.info)
	}

	counts := kindCountsFromNodes(infosForCounts)

	spec := newOpenAPISpec(title, opts.Servers)

	// Emit components and paths per node.
	for _, n := range nodes {
		// Skip nodes with no registered URIs. At runtime the api-gw
		// informer registers every CRD in the cluster (including
		// internal-only ones like rt*.hd.cisco.com whose nexus
		// annotation declares zero REST URIs); emitting their schemas
		// would leak internal types into the served spec. The
		// invariant is "a node without URIs cannot produce any
		// operation", and an unreferenced schema bag carries no
		// information for API consumers.
		if len(n.rec.uris) == 0 {
			continue
		}
		buildComponentsForNode(spec.Components, n.rec.info)
		nameParts := strings.Split(n.rec.info.Name, ".")
		tagName := qualifiedKind(nameParts, counts)
		if n.rec.info.Description != "" && tagName != "" && !hasTag(spec.Tags, tagName) {
			spec.Tags = append(spec.Tags, &openapi3.Tag{
				Name:        tagName,
				Description: n.rec.info.Description,
			})
		}

		for _, uri := range n.rec.uris {
			params := buildURIParams(uri, uri.Headers, n.rec.info.ParentHierarchy,
				nodesByCRDType, nodesByName, opts)
			pathItem := buildPathItem(uri, n.rec.info, params, counts)
			mergePathItem(spec.Paths, uri.URI, pathItem)
		}
	}

	// Emit extension paths last. They overwrite any auto-emitted path
	// at the same URI (matches existing runtime behaviour where the
	// declarative path wins over the datamodel-derived one).
	for _, ext := range extensions {
		pathItem, err := parseExtensionPathItem(ext.OpenAPIPathSpec)
		if err != nil {
			logger.Warnf("openapi-builder: skipping extension %q: %v", ext.URI, err)
			continue
		}
		spec.Paths.Set(ext.URI, pathItem)
	}

	// Prune root-level tag entries that no emitted operation references.
	// A CRD may declare Description but have its paths overwritten by an
	// extension, or simply have no registered URIs — in either case the
	// document-level Tag entry is an orphan and must not be emitted.
	spec.Tags = pruneUnreferencedTags(spec.Tags, spec.Paths)

	return spec
}

// pruneUnreferencedTags returns the subset of `tags` that is referenced
// by at least one operation in `paths`. Preserves input ordering.
func pruneUnreferencedTags(tags openapi3.Tags, paths *openapi3.Paths) openapi3.Tags {
	if len(tags) == 0 || paths == nil {
		return tags
	}
	used := make(map[string]struct{})
	for _, item := range paths.Map() {
		if item == nil {
			continue
		}
		for _, op := range item.Operations() {
			if op == nil {
				continue
			}
			for _, t := range op.Tags {
				used[t] = struct{}{}
			}
		}
	}
	out := tags[:0]
	for _, t := range tags {
		if _, ok := used[t.Name]; ok {
			out = append(out, t)
		}
	}
	return out
}

// newOpenAPISpec constructs the base OpenAPI document with the
// default components (`DefaultResponse`, `NotFoundResponse`) referenced
// throughout the spec.
func newOpenAPISpec(title string, servers openapi3.Servers) *openapi3.T {
	if len(servers) == 0 {
		servers = openapi3.Servers{
			&openapi3.Server{Description: "API Gateway", URL: "/"},
		}
	}
	return &openapi3.T{
		OpenAPI: "3.1.1",
		Info: &openapi3.Info{
			Title:   title,
			Version: "1.0.0",
		},
		Servers: servers,
		Paths:   openapi3.NewPaths(),
		Components: &openapi3.Components{
			Schemas:       openapi3.Schemas{},
			RequestBodies: openapi3.RequestBodies{},
			Responses: openapi3.ResponseBodies{
				"DefaultResponse": &openapi3.ResponseRef{
					Value: openapi3.NewResponse().
						WithDescription("Default response").
						WithContent(openapi3.NewContentWithJSONSchema(
							openapi3.NewSchema().WithProperty("message", openapi3.NewStringSchema()),
						)),
				},
				"NotFoundResponse": &openapi3.ResponseRef{
					Value: openapi3.NewResponse().
						WithDescription("Not Found").
						WithContent(openapi3.NewContentWithJSONSchema(
							openapi3.NewSchema().WithProperty("message", openapi3.NewStringSchema()),
						)),
				},
			},
		},
	}
}

// mergePathItem installs `incoming`'s non-nil operations into the
// path item at `uri` in `paths`. If no path item exists at `uri`, the
// incoming one is set verbatim. If one exists, operations from
// `incoming` overwrite the existing operations on the same methods.
// This matches the existing AddPath merge semantics in both consumers.
func mergePathItem(paths *openapi3.Paths, uri string, incoming *openapi3.PathItem) {
	existing := paths.Value(uri)
	if existing == nil {
		paths.Set(uri, incoming)
		return
	}
	if incoming.Get != nil {
		existing.Get = incoming.Get
	}
	if incoming.Put != nil {
		existing.Put = incoming.Put
	}
	if incoming.Post != nil {
		existing.Post = incoming.Post
	}
	if incoming.Delete != nil {
		existing.Delete = incoming.Delete
	}
	if incoming.Options != nil {
		existing.Options = incoming.Options
	}
	if incoming.Head != nil {
		existing.Head = incoming.Head
	}
	if incoming.Patch != nil {
		existing.Patch = incoming.Patch
	}
	if incoming.Trace != nil {
		existing.Trace = incoming.Trace
	}
}
