<!--
SPDX-FileCopyrightText: (C) 2025 Intel Corporation
SPDX-License-Identifier: Apache-2.0
-->

# openapi-builder

Single source of truth for OpenAPI spec emission from the Nexus datamodel.

This module is a pure library (no I/O, no k8s client, no HTTP framework) consumed by:

- `nexus/openapi-generator` — build-time, reads CRD YAMLs from disk and writes a static JSON spec.
- `nexus/api-gw` — runtime, reads CRDs from `model.CrdTypeToNodeInfo` and serves the spec at `/{datamodel}/openapi.json`.

Both consumers feed the builder with the same input types (`NodeInfo`, `RestURIs`, `ExtensionSpec`) and call `Build(domain)` to obtain a deterministic spec. Collision detection (`qualifiedKind`) is scoped per-domain.

## Why a separate module

`nexus/` is a multi-module tree; every sibling (`api-gw`, `openapi-generator`, `compiler`, ...) is its own Go module. Keeping the builder as its own module follows that pattern, gives it an independent version surface, and makes drift between consumers structurally impossible — both must use the same version of the library.

See `nexus/docs/openapi-unification-plan.md` for the design rationale and migration plan.
