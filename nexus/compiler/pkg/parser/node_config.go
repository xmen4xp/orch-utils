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
		val := strings.Split(annotationValue, ":")
		if len(val) == 2 {
			return strings.TrimSpace(val[1]), true
		}

		return annotationValue, false
	}

	return annotationValue, false
}
