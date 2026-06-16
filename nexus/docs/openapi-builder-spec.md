# OpenAPI Builder — Design Specification

## 1. Problem Statement

The Nexus framework emits an OpenAPI specification for every datamodel from two independent code paths:

| Path | Where | When | Output |
|------|-------|------|--------|
| Build time | `nexus/openapi-generator` (standalone binary) | At datamodel build, invoked by the consumer repo's datamodel-build target | A static JSON file (`build/openapi/<datamodel>.json`) committed to the consumer repo |
| Runtime | `nexus/api-gw` (in-process module) | At every CRD reconcile event in the cluster | A live JSON document served at `/{datamodel}/openapi.json` |

Both paths historically maintained their own copies of every emission rule — collision detection, `qualifiedKind` derivation, operationId generation, tag registration, schema construction, URI-parameter extraction, and extension-path inlining. The two implementations drifted over time, producing the following customer-visible defects:

- **Orphan tags.** The api-gw runtime accumulated state across reconcile events. A CRD that arrived before its `Kind` had a colliding sibling was emitted with an unqualified tag. When the colliding sibling later arrived, the runtime emitted a second qualified tag without removing the first, leaving an empty bare tag in the served spec (e.g. `User` and `UsersUser` both present, the former empty).
- **Over-qualified Kinds.** The collision detector scanned every registered CRD across every datamodel. A `User` CRD in datamodel `app.example.com` collided with an unrelated `User` CRD in datamodel `admin.example.com`, forcing the `app.example.com` spec to emit `UsersUser` even though no real collision existed within that datamodel.
- **Build-time/runtime drift.** A small change to one emission path (e.g. switching a `/status` URI from `GET`-only to `GET+PUT+PATCH`) had to be mirrored manually in the other. The team had no structural mechanism to keep the two in lockstep.

This specification defines **`nexus/openapi-builder`**, a single pure-library module that owns all OpenAPI emission logic. Both `nexus/api-gw` and `nexus/openapi-generator` collapse into thin adapters that translate their respective input sources into the builder's input types, call `Build(domain)`, and forward the result to their respective output sinks.

## 2. Design Overview

The solution introduces one new Go module and refactors two existing modules:

| Component | Purpose |
|-----------|---------|
| `nexus/openapi-builder` (new module) | Pure library — owns all OpenAPI emission rules. No I/O, no k8s client, no HTTP framework. |
| `SpecBuilder` (type) | Per-process container for input state. Supports `Add*` / `Remove*` / `Reset*` mutations and a snapshot-semantic `Build(domain)` |
| `Options` (type) | Per-spec configuration — header aliases, ignored parents, default title, servers |
| `nexus/api-gw/pkg/openapi/api` (adapter) | Reads `pkg/model` state, feeds the builder, exposes the result via `Schemas` map and `/{datamodel}/openapi.json` |
| `nexus/openapi-generator/pkg/api` (adapter) | Reads CRD YAML inputs, feeds the builder, serialises the result to a JSON file via the existing CLI |

### Emission flow

```
Adapter inputs (model.* or CRD YAMLs)
       │
       ▼
SpecBuilder.AddCRDNode  (per CRD type)
SpecBuilder.AddExtension (per ExtensionRestAPI CR)
SpecBuilder.SetDatamodelTitle (per datamodel)
       │
       ▼
SpecBuilder.Build(domain)            ─── snapshot ───
       │   1. RLock; copy all inputs for `domain` into local slices/maps.
       │   2. Sort inputs (deterministic order).
       │   3. Compute per-domain Kind histogram (KindCounts).
       │   4. Emit components per node (Schemas, RequestBodies, Responses).
       │   5. Emit tags (deduped, qualifiedKind-aware).
       │   6. Emit paths (per URI, per method).
       │   7. Inline extension paths.
       ▼
*openapi3.T (fresh, internally consistent)
```

Every `Build` call assembles a complete spec from the current input snapshot. There is no incremental state shared across calls. Collision detection always sees the full domain — never a partial view — so a Kind cannot be unqualified on one call and qualified on the next.

## 3. Public API

The builder exposes one type and a small set of methods:

```go
package builder

// Lifecycle.
func New() *SpecBuilder
func NewWithOptions(opts Options) *SpecBuilder
func (b *SpecBuilder) SetOptions(opts Options)
func (b *SpecBuilder) SetLogger(l Logger)

// Inputs (idempotent on the natural key).
func (b *SpecBuilder) AddCRDNode(domain, crdType string, info NodeInfo, uris []RestURIs)
func (b *SpecBuilder) RemoveCRDNode(domain, crdType string)
func (b *SpecBuilder) AddExtension(domain string, ext ExtensionSpec)
func (b *SpecBuilder) RemoveExtension(domain, uri string)
func (b *SpecBuilder) SetDatamodelTitle(domain string, t DatamodelTitle)

// Snapshot + clear.
func (b *SpecBuilder) Reset()
func (b *SpecBuilder) ResetDomain(domain string)
func (b *SpecBuilder) Domains() []string

// Emission.
func (b *SpecBuilder) Build(domain string) *openapi3.T
```

### Input types

The input types are intentionally decoupled from any consumer's `model` package so the builder takes no transitive dependencies on either api-gw or the openapi-generator. They live in `pkg/builder/types.go`:

```go
type NodeInfo struct {
    Name            string                                       // "users.User" — package.Kind
    ParentHierarchy []string
    Children        map[string]NodeHelperChild
    Links           map[string]NodeHelperChild
    IsSingleton     bool
    Description     string
    DeferredDelete  bool
    Schema          *apiextensionsv1.JSONSchemaProps             // CRD's openAPIV3Schema; nil OK
}

type RestURIs struct {
    URI        string
    Methods    map[string]struct{}                                // method set; values unused
    TypeOfURI  URIType                                            // Default / Status / SingleLink / NamedLink
    PathParams map[string]string                                  // alias → canonical groupKind
    Headers    []string                                           // explicit header parameters
}

type ExtensionSpec struct {
    URI             string
    Methods         []string
    OpenAPIPathSpec string                                        // inline YAML
    Description     string
    Tag             string
}

type DatamodelTitle struct {
    Title string
}

type Options struct {
    HeaderAliases  map[string]string                              // node name → HTTP header
    IgnoredParents map[string]struct{}                            // node names to drop entirely
    DefaultTitle   string
    Servers        openapi3.Servers
}
```

### Key behaviours

- **`AddCRDNode` is idempotent** on `(domain, crdType)`. A second call with the same key overwrites the prior entry rather than appending.
- **`Build` is a pure snapshot.** Concurrent `Add*` calls observed before the snapshot are included; calls observed after the snapshot are not. Re-calling `Build(domain)` after the same input set always produces byte-identical JSON output (deterministic ordering of nodes, extensions, components).
- **`SpecBuilder` is goroutine-safe.** A single `sync.RWMutex` guards the input maps; emission proceeds under RLock against a local copy of inputs.
- **`Logger` is optional.** When unset, all builder log lines (e.g. malformed extension YAML warnings) are silently dropped. Adapters supply their own logger.

## 4. Core Components

### 4.1 Collision Detection — `qualifiedKind`

A CRD's `NodeInfo.Name` is `package.Kind` (e.g. `users.User`). The tag and operationId emitted for the CRD use the `Kind` portion alone — unless another CRD **in the same domain** declares the same Kind, in which case the package segment is prepended for disambiguation.

```go
func qualifiedKind(nameParts []string, counts KindCounts) string {
    kind := nameParts[1]
    if counts[kind] > 1 {
        return titleFirst(nameParts[0]) + kind
    }
    return kind
}
```

`KindCounts` is computed once per `Build(domain)` call from the snapshot's NodeInfos for that domain only. This is the structural fix for the over-qualified-Kind defect: collision detection cannot see CRDs from other domains because they are never passed in.

| Domain inputs | Kind histogram | Result for `users.User` |
|---|---|---|
| `users.User`, `orgs.Org` | `{User: 1, Org: 1}` | `User` (bare) |
| `pkga.Widget`, `pkgb.Widget` | `{Widget: 2}` | `PkgaWidget` and `PkgbWidget` |
| `users.User` in datamodel A **and separately** `user.User` in datamodel B | Two independent histograms | `User` in each spec |

### 4.2 OperationId Derivation — `GetOperationID`

| URI type | Method | OperationId |
|----------|--------|-------------|
| Default | LIST | `list<KindPlural>` |
| Default | GET | `get<Kind>` |
| Default | PUT / PATCH / DELETE | `<verb><Kind>` |
| Status | any verb | `<verb><Kind>Status` |
| SingleLink / NamedLink | GET | `get<ParentKind><FieldName>` |

`<Kind>` is the `qualifiedKind` result, so colliding Kinds get the package prefix on the operationId as well as the tag. `<KindPlural>` appends `s` unless the Kind already ends in `s` (e.g. `Clusters`, `DataCenters`, `Nodes` stay as-is). `<FieldName>` is the trailing static segment of the link URI.

### 4.3 Component Emission — `buildComponentsForNode`

For every node with a non-nil `Schema`, the builder emits six schema components, two request bodies, and five responses keyed off the full `NodeInfo.Name`:

| Component key | Content |
|---|---|
| `<name>.Get` | `{ spec: <specSchema>, status: <statusSchema> }` |
| `<name>.Post` | `<specSchema>` only — used as the create request body |
| `<name>.List` | Array of `{ name, spec, status }` |
| `<name>.Status` | `<statusSchema>` only (the `nexus` sub-property is stripped) |
| `<name>.SingleLink` | Empty placeholder object |
| `<name>.NamedLink` | Array of `<name>.SingleLink` placeholders |

| Request body | Body schema | Used by |
|---|---|---|
| `Create<name>` | `<name>.Post` | PUT/PATCH on Default URI |
| `Create<name>.Status` | `<name>.Status` | PUT/PATCH on Status URI |

| Response | Content schema | Used by |
|---|---|---|
| `Get<name>` | `<name>.Get` | GET on Default URI |
| `Get<name>.Status` | `<name>.Status` | GET on Status URI |
| `Get<name>.SingleLink` | `<name>.SingleLink` | GET on SingleLink URI |
| `Get<name>.NamedLink` | `<name>.NamedLink` | GET on NamedLink URI |
| `List<name>` | `<name>.List` | LIST on Default URI |

`parseFields` walks `JSONSchemaProps.Properties` recursively, mapping each property to an `openapi3.Schema` and propagating any `Description` / `Example` annotations from the source datamodel comments. Fields containing `Gvk` in the name are skipped (they are GVK metadata Nexus injects into every type and are not customer-facing).

### 4.4 URI Parameter Emission — `buildURIParams`

For each URI, the builder emits:

1. **Path parameters** — one per `{token}` in the URI template. Description is sourced from the resolved canonical node's `NodeInfo.Description` when available, otherwise defaults to `Name of the <token> node`.
2. **Header parameters** — one per entry in `RestURIs.Headers`. These are explicit headers declared by the datamodel author.
3. **Hierarchy parameters** — one per non-singleton ancestor CRD in `NodeInfo.ParentHierarchy`. Each is emitted as a query parameter by default. Two `Options` fields modify this behaviour:

   | Option | Effect |
   |--------|--------|
   | `HeaderAliases[name] = "x-foo"` | Emit `name` as required header `x-foo` instead of a query parameter |
   | `IgnoredParents[name] = struct{}{}` | Drop `name` from the parameter list entirely |

   When both `HeaderAliases[name]` and `IgnoredParents[name]` are set, the HeaderAlias wins — the parent is promoted to a header parameter.

Path tokens are resolved to canonical groupKinds via `RestURIs.PathParams` before description lookup; the OpenAPI parameter name remains the original alias. PUT operations on Default URIs additionally append an `update_if_exists` boolean query parameter.

### 4.5 Path Item Emission — `buildPathItem`

`buildPathItem` consumes one `RestURIs` value, looks up its tag via `qualifiedKind`, and emits one `openapi3.Operation` per method in the URI's `Methods` set. Each operation carries:

| Field | Value |
|-------|-------|
| `OperationId` | From `GetOperationID` |
| `Tags` | `[qualifiedKind]` — single-element |
| `Parameters` | From `buildURIParams` (plus the `update_if_exists` query for PUT on Default URIs) |
| `RequestBody` | `$ref` to `Create<name>` or `Create<name>.Status` (PUT/PATCH only) |
| `Responses` | `200` ref to the appropriate component response; PATCH also gets `404` to `NotFoundResponse` |

When two `RestURIs` values share the same URI string, `mergePathItem` overlays non-nil methods from each — preserving the existing AddPath behaviour where a status URI's `PUT` does not clobber a sibling-emitted `GET`.

### 4.6 Extension Path Emission — `parseExtensionPathItem`

`ExtensionSpec.OpenAPIPathSpec` is a verbatim YAML fragment describing an `openapi3.PathItem`. The builder parses it via `sigs.k8s.io/yaml`, unmarshals into `*openapi3.PathItem`, and installs it on the spec's `Paths` under `ExtensionSpec.URI`. Extension paths **overwrite** any auto-emitted path at the same URI (matching the existing runtime behaviour where declarative paths win over datamodel-derived ones). Malformed extension YAML is logged via the optional `Logger` and skipped — never panicked.

## 5. Adapter Integration

Each consumer collapses into a thin shim over the builder. Both adapters maintain the same public surface they had before the refactor so downstream callers compile unchanged.

### 5.1 api-gw — Runtime Adapter

Lives in `nexus/api-gw/pkg/openapi/api/openapi.go`.

| Public symbol | Behaviour |
|---|---|
| `Schemas map[string]openapi3.T` | Per-datamodel spec, refreshed on every Add/Recreate call. Read by `pkg/openapi/combined` and by the Echo handler at `/{datamodel}/openapi.json`. |
| `New(datamodel)` | Resets the builder's domain state for `datamodel`, sets the title, rebuilds and publishes to `Schemas`. |
| `AddPath(uri, datamodel)` | Adds the CRD's full URI list to the builder, rebuilds and publishes. |
| `AddExtensionPath(spec, datamodel)` | Adds the extension spec to the builder, rebuilds and publishes. |
| `Recreate()` | Snapshots all CRDs + extensions + titles from `pkg/model`, resets the builder, re-adds everything, rebuilds every affected domain. This is the canonical full-rebuild entry point invoked by `echo_server` on CRD reconcile events. |
| `RecreateExtension()` | No-op (kept for backward compatibility). Extensions are now part of the same snapshot as CRDs. |
| `DatamodelUpdateNotification()` | Listens on `model.DatamodelsChan`; refreshes the title and rebuilds. |

`Options` are sourced from `config.Cfg.HeaderAliases` at every Add/Recreate call so config changes propagate without restart. The api-gw runtime sets no `IgnoredParents` — that field is exclusive to the build-time consumer.

### 5.2 openapi-generator — Build-Time Adapter

Lives in `nexus/openapi-generator/pkg/api/api.go`.

| Public symbol | Behaviour |
|---|---|
| `Schemas map[string]openapi3.T` | Read by `main.go` after the last `AddPath` call; serialised to disk via `json.MarshalIndent`. |
| `New(datamodel)` | Resets domain state, sets title, publishes empty spec. |
| `AddPath(uri, datamodel)` | Adds the URI to the builder, rebuilds. |
| `AddExtensionPath(fileBytes, datamodel)` | Parses an ExtensionRestAPI CR YAML, adds it to the builder, rebuilds. Mixed directories are tolerated — non-`ExtensionRestAPI` payloads are silently skipped. |
| `ConstructNewURIs(annotation, urisMap, newUris)` | Build-time URI discovery — unchanged. Walks children/links/status URIs derived from a node's annotation and registers them in `urisMap`. |

`Options` are sourced from `model.OpenApiIgnoredParentPathParams` and `model.OpenApiNodeToHeaderMapping`, populated from the build-time datamodel config file.

### 5.3 Type Conversion

Each adapter ships a `convert.go` (openapi-generator) or in-line helpers (api-gw) that translate the consumer's model types into the builder's input types. The translation is mechanical: copy the fields, map enum values, copy `*apiextensionsv1.CustomResourceDefinitionSpec.Versions[0].Schema.OpenAPIV3Schema` into `NodeInfo.Schema`.

## 6. Module Layout

```
nexus/openapi-builder/                          NEW Go module
├── go.mod                                       module nexus/openapi-builder
├── go.sum
├── README.md
└── pkg/builder/
    ├── types.go                                 input types
    ├── qualifiedkind.go                         qualifiedKind / KindCounts / kindPlural / titleFirst / lastStaticSegment
    ├── qualifiedkind_test.go
    ├── operationid.go                           GetOperationID
    ├── operationid_test.go
    ├── schema.go                                buildComponentsForNode / parseFields / buildPropSchema
    ├── schema_test.go
    ├── uriparams.go                             buildURIParams
    ├── pathitem.go                              buildPathItem + per-method emitters + mergePathItem + hasTag
    ├── extension.go                             parseExtensionPathItem
    ├── builder.go                               SpecBuilder + Options + Build
    └── builder_test.go
```

The module is a peer of the existing `nexus/api-gw`, `nexus/openapi-generator`, `nexus/compiler`, etc. — every component under `nexus/` is its own Go module. Consumers reference it via `replace nexus/openapi-builder => ../openapi-builder` in their `go.mod` files, matching the existing pattern used for `nexus/admin/api` and the nexus framework itself.

### Dependency surface

```go
require (
    github.com/getkin/kin-openapi v0.131.0     // openapi3 types
    k8s.io/apiextensions-apiserver v0.32.3     // JSONSchemaProps
    sigs.k8s.io/yaml v1.6.0                    // extension YAML parsing
)
```

No `client-go`, no `echo`, no `logrus`, no `controller-runtime`. The `Logger` interface accepts any type with a `Warnf(format string, args ...interface{})` method.

## 7. Snapshot Semantics — Why The Bugs Cannot Recur

### 7.1 Orphan Tags

The historical defect was that api-gw's `Recreate()` was *not* idempotent: it called `New(datamodel)` only when the datamodel was missing from `Schemas`, and `AddPath` appended to the existing tag list without checking whether existing tags were still correct under the current CRD set. A CRD that arrived alone got an unqualified tag; when its sibling later arrived, the unqualified tag was never removed.

The builder eliminates this class of bug at the API layer. There is no `AddPath`-equivalent method that mutates a previously-published spec. The only mutations live on `SpecBuilder` (which holds *inputs*, not *outputs*) and `Build(domain)` is a one-shot snapshot that constructs a fresh `*openapi3.T` from those inputs. The `Schemas` map exposed by the adapter is overwritten wholesale on every call. Orphan tags are physically impossible.

Regression test: `TestBuild_SnapshotDeterminism` in `pkg/builder/builder_test.go` adds a noise node, removes it, and asserts that `Build` output is byte-identical to a build without the noise node. Any code path that left residue would break this test.

### 7.2 Over-Qualified Kinds

The historical defect was that the global `IsCollidingKind(kind)` scanned every entry in the process-wide `CrdTypeToNodeInfo` map, conflating CRDs from unrelated datamodels.

The builder eliminates this at the type-system level. `qualifiedKind` takes a `KindCounts` parameter — a per-domain Kind histogram — and there is no way to compute or supply one that includes CRDs from another domain. `Build(domain)` computes `KindCounts` inside its snapshot from that domain's NodeInfos alone, then passes it down to `GetOperationID` and tag emission. Cross-domain leakage is structurally impossible.

Regression test: `TestQualifiedKind_PerDomainScoping` and `TestBuild_PerDomainCollisionScoping` exercise the exact `User` vs `User` cross-datamodel scenario and assert both specs emit a bare `User` tag.

## 8. Testing

| Test file | Coverage |
|-----------|----------|
| `qualifiedkind_test.go` | All collision/no-collision permutations + the per-domain scoping regression test + plural / title-case / link-suffix helpers |
| `operationid_test.go` | Full Cartesian product of method × URI type, with and without collision |
| `schema_test.go` | All primitive / object / array types, nested structures, Gvk skipping, `nexus` stripping from status, description/example propagation |
| `builder_test.go` | Empty domain, basic CRUD, per-domain collision scoping, same-domain qualification, snapshot determinism across insertion orders, status URI verbs, extension installation, Reset / ResetDomain |

All tests run under `go test ./...` in the `nexus/openapi-builder` module and pass with `-race`.

### Adapter-level tests

Both adapters keep their pre-existing test suites unchanged (the api-gw ginkgo suite, the openapi-generator integration tests). These now exercise the adapter glue without re-testing emission rules, which are owned by the builder's own tests.

## 9. Deployment

### Image rebuild

After the refactor merges, the CI pipeline rebuilds the standard nexus images including `nexus-api-gw` and `nexus-openapi-generator`. The Go module wiring (replace directives + go.sum updates in both consumers) is handled by the existing build pipeline; no new wiring is required.

### Image consumption

| Image | Required? |
|-------|-----------|
| `nexus-api-gw` | **Yes** — applies the runtime fixes (orphan tags, over-qualified Kinds). |
| `nexus-openapi-generator` | Optional — build-time spec shape is identical to today's output. Bump at your convenience. |

### Backward compatibility

- `nexus/openapi-generator` CLI flags (`--yamls-path`, `--datamodel-path`, `--datamodel-name`, `--output-file-path`) are unchanged. Consumer-repo datamodel-build invocations, CI workflows, and dependency-pinning schemas all stay the same.
- `nexus/api-gw` HTTP route `/:datamodel/openapi.json` is unchanged.
- The exported `api.Schemas` map continues to be read by `pkg/openapi/combined` and by the Echo handler.

## 10. /status Endpoint Behaviour

| Spec | `/status` methods emitted |
|------|---------------------------|
| Build-time (`<datamodel>.json`) | `GET` only — preserved by the openapi-generator adapter's URI discovery (`addStatusURI`). |
| Runtime (api-gw `/{datamodel}/openapi.json`) | `GET + PUT + PATCH` — preserved by the api-gw runtime input feed (`controllers/process.go`'s `addStatusURI` registers `GET + PUT`; `echo_server.go` mirrors `PUT → PATCH` before reconciliation). |

The builder is input-driven: it emits exactly the methods present in `RestURIs.Methods` for a given URI. Both consumers feed their existing method sets unchanged, so neither spec changes shape.

## 11. Limitations and Future Work

- **No POST**: Mirrors the existing emission rules — POST is reserved for Nexus-managed object creation via PUT. Adding POST support requires both compiler and builder changes; out of scope for this refactor.
- **Placeholder SingleLink / NamedLink schemas**: The component emission for link traversal targets uses empty placeholder objects, matching today's behaviour. A future change to emit the actual referenced node's schema is a builder-only change.
- **No spec caching**: `Build(domain)` is called on every Add / Recreate / HTTP request. At today's scale this is sub-millisecond; if profiling later shows it as a hotspot, a per-domain build-result cache invalidated on input mutation can be added inside the builder.
- **`Logger` is `Warnf`-only**: Adapters needing finer log-level control should wrap their own logger to satisfy the interface. The minimal surface is deliberate — it lets the builder stay agnostic to slog / zerolog / logrus / klog choices in the consumers.
