// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"go/ast"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// versionSuffixRe matches a trailing version suffix on a Go type name, e.g.
// "OrgV2" -> suffix "V2". Used only as a convenience default for nexus-kind
// when it isn't explicitly annotated.
var versionSuffixRe = regexp.MustCompile(`V(\d+)$`)

// resolveNodeIdentity determines the Kind/Version/storage-version-ness of a
// nexus node struct from its nexus-version / nexus-kind / nexus-storage-version
// doc comment annotations (see node_config.go), falling back to today's
// implicit behavior when none are present:
//
//   - No annotations at all -> version "v1", kind == struct name, storage
//     version true. This is exactly today's (pre-versioning) behavior, so
//     every existing DSL type is unaffected.
//   - nexus-version present but nexus-kind absent -> kind defaults to the
//     struct name with a trailing "V<N>" suffix stripped (e.g. "OrgV2" ->
//     "Org"), so a second version of a Kind can be declared without having
//     to repeat the kind name in an annotation.
//   - storage-version-ness defaults to true only for version "v1"; any other
//     version must opt in explicitly via `// nexus-storage-version: true` if
//     it is meant to become the new storage version.
// ResolveNodeIdentity is the exported form of resolveNodeIdentity, for use by
// other packages (e.g. the CRD/openapi generators) that need to group struct
// versions by Kind without re-implementing annotation resolution.
func ResolveNodeIdentity(pkg Package, structName string) (kind, version string, isStorageVersion bool) {
	return resolveNodeIdentity(pkg, structName)
}

func resolveNodeIdentity(pkg Package, structName string) (kind, version string, isStorageVersion bool) {
	version = "v1"
	if v, ok := GetNexusVersionAnnotation(pkg, structName); ok && strings.TrimSpace(v) != "" {
		version = strings.TrimSpace(v)
	}

	kind = structName
	if k, ok := GetNexusKindAnnotation(pkg, structName); ok && strings.TrimSpace(k) != "" {
		kind = strings.TrimSpace(k)
	} else if version != "v1" {
		if stripped := versionSuffixRe.ReplaceAllString(structName, ""); stripped != "" && stripped != structName {
			kind = stripped
		}
	}

	isStorageVersion = version == "v1"
	if sv, ok := GetNexusStorageVersionAnnotation(pkg, structName); ok && strings.TrimSpace(sv) != "" {
		parsed, err := strconv.ParseBool(strings.TrimSpace(sv))
		if err != nil {
			log.Fatalf("Invalid nexus-storage-version annotation value %q on type %s: must be true or false", sv, structName)
		}
		isStorageVersion = parsed
	}

	return kind, version, isStorageVersion
}

// scanConversionImplementors walks every file in a Go AST package and returns
// the set of struct names that implement both ConvertTo(conversion.Hub) error
// and ConvertFrom(conversion.Hub) error methods, mirroring the
// sigs.k8s.io/controller-runtime conversion.Convertible interface. These are
// the methods a non-storage-version node must provide to convert to/from its
// Kind's storage version.
func scanConversionImplementors(pkg *ast.Package) map[string]bool {
	methodsByReceiver := map[string]map[string]bool{}

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Recv == nil || len(funcDecl.Recv.List) != 1 {
				continue
			}
			if funcDecl.Name.Name != "ConvertTo" && funcDecl.Name.Name != "ConvertFrom" {
				continue
			}

			recvType := funcDecl.Recv.List[0].Type
			if starExpr, ok := recvType.(*ast.StarExpr); ok {
				recvType = starExpr.X
			}
			ident, ok := recvType.(*ast.Ident)
			if !ok {
				continue
			}

			if methodsByReceiver[ident.Name] == nil {
				methodsByReceiver[ident.Name] = map[string]bool{}
			}
			methodsByReceiver[ident.Name][funcDecl.Name.Name] = true
		}
	}

	implementors := map[string]bool{}
	for structName, methods := range methodsByReceiver {
		implementors[structName] = methods["ConvertTo"] && methods["ConvertFrom"]
	}
	return implementors
}

// mergeNodeVersion folds `incoming`, a newly-parsed version of the same Kind
// as `canonical`, into the graph. Exactly one version of a Kind may be the
// storage version; whichever one is gets to be the canonical graph node (the
// one used for children/parents/REST-hierarchy resolution), and the other is
// recorded in canonical.OtherVersions. Parse order (which version is seen
// first while walking the DSL tree) does not matter.
func mergeNodeVersion(canonical *Node, incoming Node) {
	if canonical.IsStorageVersion && incoming.IsStorageVersion {
		log.Fatalf("Kind %q has more than one storage version declared: %q (version=%s) and %q (version=%s). "+
			"Exactly one version of a Kind must be the storage version.",
			canonical.Kind, canonical.Name, canonical.Version, incoming.Name, incoming.Version)
		// log.Fatalf exits the process in production. If ExitFunc has been
		// overridden (e.g. in tests) and execution reaches here anyway, keep
		// the original canonical node intact rather than silently replacing
		// its identity, so callers that already resolved references to it by
		// name don't fail with an unrelated nil-pointer error.
		incoming.IsStorageVersion = false
	}

	if incoming.IsStorageVersion {
		demoted := *canonical
		demoted.OtherVersions = nil
		incoming.OtherVersions = append(append([]Node{}, canonical.OtherVersions...), demoted)
		*canonical = incoming
		return
	}

	canonical.OtherVersions = append(canonical.OtherVersions, incoming)
}

// validateNodeVersions runs once, after the whole DSL tree has been walked
// and all versions of every Kind have been merged, and enforces the
// invariants multi-version support depends on:
//   - every Kind resolves to exactly one storage version.
//   - every non-storage version implements ConvertTo/ConvertFrom (i.e. can
//     convert to/from the storage version).
func validateNodeVersions(nodes map[string]Node) {
	for crdName, node := range nodes {
		if !node.IsStorageVersion {
			log.Fatalf("Kind %q (crd %q) has no storage version. Exactly one version of a Kind "+
				"(by default, the version named \"v1\") must be the storage version.", node.Kind, crdName)
		}
		for _, other := range node.OtherVersions {
			if !other.ConversionImplemented {
				log.Fatalf("Version %q of Kind %q (type %s) must implement ConvertTo(conversion.Hub) error and "+
					"ConvertFrom(conversion.Hub) error to convert to/from its storage version %q (type %s)",
					other.Version, node.Kind, other.Name, node.Version, node.Name)
			}
		}
	}
}
