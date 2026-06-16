// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package api

import (
	yamlv1 "github.com/ghodss/yaml"
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"nexus/openapi-builder/pkg/builder"
	"nexus/openapi-generator/pkg/model"
)

// toBuilderNodeInfo converts a model.NodeInfo + CRD spec into the
// builder's NodeInfo type. The builder owns no model package; this
// is the only adapter-to-library type bridge.
func toBuilderNodeInfo(info model.NodeInfo, spec apiextensionsv1.CustomResourceDefinitionSpec) builder.NodeInfo {
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
			out.Children[k] = builder.NodeHelperChild{FieldName: v.FieldName, FieldNameGvk: v.FieldNameGvk, IsNamed: v.IsNamed}
		}
	}
	if info.Links != nil {
		out.Links = make(map[string]builder.NodeHelperChild, len(info.Links))
		for k, v := range info.Links {
			out.Links[k] = builder.NodeHelperChild{FieldName: v.FieldName, FieldNameGvk: v.FieldNameGvk, IsNamed: v.IsNamed}
		}
	}
	if len(spec.Versions) > 0 && spec.Versions[0].Schema != nil &&
		spec.Versions[0].Schema.OpenAPIV3Schema != nil {
		out.Schema = spec.Versions[0].Schema.OpenAPIV3Schema
	}
	return out
}

// toBuilderURIs converts a slice of nexus.RestURIs into the builder's
// RestURIs type. URI type classification is read from
// `model.UriToUriInfo`, populated by ConstructMapUriToUriInfo.
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
		if info, ok := model.UriToUriInfo[u.Uri]; ok {
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

// ghYAMLToJSON is a thin wrapper around ghodss/yaml's YAMLToJSON used
// by parseExtensionEnvelope. Centralising the import keeps api.go's
// import set small.
func ghYAMLToJSON(in []byte) ([]byte, error) {
	return yamlv1.YAMLToJSON(in)
}
