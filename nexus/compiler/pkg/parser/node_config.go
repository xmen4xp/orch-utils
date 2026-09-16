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
)

// NexusOnDeleteTag is the per-child struct tag that controls what happens to a
// child when its parent is deleted, e.g.:
//
//	AISlices aislice.AISlice `nexus:"child" nexus-on-delete:"restrict"`
const NexusOnDeleteTag = "nexus-on-delete"

// Supported values for the nexus-on-delete tag.
const (
	// DeletionPolicyCascade is the default: the child is deleted along with its
	// parent. Assumed when the tag is absent.
	DeletionPolicyCascade = "cascade"
	// DeletionPolicyRestrict rejects deletion of the parent while this child exists.
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
