// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"go/doc"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	NexusRestApiGenAnnotation       = "nexus-rest-api-gen"
	NexusExtensionRestAPIAnnotation = "nexus-extension-rest-api"
	NexusDescriptionAnnotation      = "nexus-description"
	NexusGraphqlAnnotation          = "nexus-graphql-query"
	NexusSecretSpecAnnotation       = "nexus-secret-spec"
	NexusGraphqlSpecAnnotation      = "nexus-graphql-spec"
	NexusDeferredDeleteAnnotation   = "nexus-deferred-delete"
	NexusDeletionPolicyAnnotation   = "nexus-deletion-policy"
	// NexusDeletionRestrictChildrenAnnotation optionally scopes a restrict
	// policy to specific child kinds (comma-separated). When absent, a restrict
	// policy applies to all children.
	NexusDeletionRestrictChildrenAnnotation = "nexus-deletion-restrict-children"
)

// Supported values for the nexus-deletion-policy annotation.
const (
	// DeletionPolicyCascade is the default policy: all children are deleted
	// before the node itself is deleted. This matches the historical behavior
	// and is assumed when the annotation is absent.
	DeletionPolicyCascade = "cascade"
	// DeletionPolicyRestrict rejects deletion of a node while it still has children.
	DeletionPolicyRestrict = "restrict"
)

func GetNexusSecretSpecAnnotation(pkg Package, name string) (string, bool) {
	return getNexusAnnotation(pkg, name, NexusSecretSpecAnnotation)
}

func GetNexusRestAPIGenAnnotation(pkg Package, name string) (string, bool) {
	anno, ok := getNexusAnnotation(pkg, name, NexusRestApiGenAnnotation)
	if ok && !pkg.IsVarPresent(anno) {
		log.Fatalf("Error: var %+s is not present", anno)
	}
	return anno, ok
}

func GetNexusDescriptionAnnotation(pkg Package, name string) (string, bool) {
	return getNexusAnnotation(pkg, name, NexusDescriptionAnnotation)
}

func GetNexusGraphqlAnnotation(pkg Package, name string) (string, bool) {
	return getNexusAnnotation(pkg, name, NexusGraphqlAnnotation)
}

func GetNexusDeferredDeleteAnnotation(pkg Package, name string) (string, bool) {
	return getNexusAnnotation(pkg, name, NexusDeferredDeleteAnnotation)
}

func GetNexusDeletionPolicyAnnotation(pkg Package, name string) (string, bool) {
	return getNexusAnnotation(pkg, name, NexusDeletionPolicyAnnotation)
}

// GetNexusDeletionRestrictChildrenAnnotation returns the list of child kinds a
// restrict policy is scoped to. Returns (nil, false) when the annotation is
// absent or empty, which callers should treat as "all children".
func GetNexusDeletionRestrictChildrenAnnotation(pkg Package, name string) ([]string, bool) {
	anno, ok := getNexusAnnotation(pkg, name, NexusDeletionRestrictChildrenAnnotation)
	if !ok || strings.TrimSpace(anno) == "" {
		return nil, false
	}
	var children []string
	for _, part := range strings.Split(anno, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			children = append(children, trimmed)
		}
	}
	return children, len(children) > 0
}

func GetNexusGraphqlSpecAnnotation(pkg Package, name string) (string, bool) {
	return getNexusAnnotation(pkg, name, NexusGraphqlSpecAnnotation)
}

func GetNexusExtensionRestAPIAnnotations(pkg Package, name string) ([]string, bool) {
	anno, ok := getNexusAnnotation(pkg, name, NexusExtensionRestAPIAnnotation)
	if !ok || anno == "" {
		return nil, false
	}

	// Split by comma and trim whitespace
	parts := strings.Split(anno, ",")
	var annotations []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		if !pkg.IsVarPresent(trimmed) {
			log.Fatalf("Error: var %s is not present", trimmed)
		}
		annotations = append(annotations, trimmed)
	}

	return annotations, len(annotations) > 0
}

func getNexusAnnotation(pkg Package, name string, annotationName string) (string, bool) {
	var annotationValue string

	d := doc.New(&pkg.Pkg, pkg.Name, 4)
	for _, t := range d.Types {
		if t.Name == name {
			if strings.Contains(t.Doc, annotationName) {
				re := regexp.MustCompile(annotationName + ".*")
				annotationValue = re.FindString(t.Doc)
			}
		}
	}

	if annotationValue != "" {
		// Split on the FIRST colon only — annotation values (notably
		// nexus-description) may legitimately contain additional colons in
		// their prose, e.g. "AlertRule defines an alert evaluation: severity, ...".
		val := strings.SplitN(annotationValue, ":", 2)
		if len(val) == 2 {
			return strings.TrimSpace(val[1]), true
		}

		return annotationValue, false
	}

	return annotationValue, false
}
