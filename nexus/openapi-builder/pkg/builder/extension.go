// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"encoding/json"

	"github.com/getkin/kin-openapi/openapi3"
	"sigs.k8s.io/yaml"
)

// parseExtensionPathItem converts a raw OpenAPIPathSpec YAML fragment
// into an openapi3.PathItem suitable for installing into a spec under
// the given URI. Returns (nil, error) for invalid YAML or invalid
// PathItem content. Callers should log the error and skip the URI; the
// builder does not panic on malformed extensions.
func parseExtensionPathItem(openAPIPathSpec string) (*openapi3.PathItem, error) {
	if openAPIPathSpec == "" {
		return nil, errEmptyExtensionSpec
	}
	pathJSON, err := yaml.YAMLToJSON([]byte(openAPIPathSpec))
	if err != nil {
		return nil, err
	}
	pathItem := &openapi3.PathItem{}
	if err := pathItem.UnmarshalJSON(pathJSON); err != nil {
		return nil, err
	}
	return pathItem, nil
}

// parseExtensionCREnvelope decodes a raw ExtensionRestAPI CR YAML/JSON
// payload (as produced by the datamodel build pipeline) and extracts
// the URI + OpenAPIPathSpec for the builder to consume. Useful for the
// openapi-generator adapter which reads CRs from disk; api-gw consumes
// already-parsed `ExtensionSpec` instances from its model package.
func parseExtensionCREnvelope(fileBytes []byte) (ExtensionSpec, bool, error) {
	var env struct {
		Kind string `json:"kind"`
		Spec struct {
			URI             string   `json:"uri"`
			OpenAPIPathSpec string   `json:"openAPIPathSpec"`
			Methods         []string `json:"methods"`
			Description     string   `json:"description"`
		} `json:"spec"`
	}
	asJSON, err := yaml.YAMLToJSON(fileBytes)
	if err != nil {
		return ExtensionSpec{}, false, err
	}
	if err := json.Unmarshal(asJSON, &env); err != nil {
		return ExtensionSpec{}, false, err
	}
	if env.Kind != "ExtensionRestAPI" {
		return ExtensionSpec{}, false, nil // not an extension manifest; skip
	}
	if env.Spec.URI == "" || env.Spec.OpenAPIPathSpec == "" {
		return ExtensionSpec{}, false, errIncompleteExtensionSpec
	}
	return ExtensionSpec{
		URI:             env.Spec.URI,
		OpenAPIPathSpec: env.Spec.OpenAPIPathSpec,
		Methods:         env.Spec.Methods,
		Description:     env.Spec.Description,
	}, true, nil
}

// errEmptyExtensionSpec is returned by parseExtensionPathItem when the
// OpenAPIPathSpec string is empty.
var errEmptyExtensionSpec = pathSpecError("openAPIPathSpec is empty")

// errIncompleteExtensionSpec is returned by parseExtensionCREnvelope
// when the CR omits either spec.uri or spec.openAPIPathSpec.
var errIncompleteExtensionSpec = pathSpecError("ExtensionRestAPI CR missing spec.uri or spec.openAPIPathSpec")

type pathSpecError string

func (e pathSpecError) Error() string { return string(e) }
