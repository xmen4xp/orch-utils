// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package openapi_generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/common"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

const defaultNamePrefix = "nexustempmodule/apis"

type definition struct {
	schema       *extensionsv1.JSONSchemaProps
	dependencies []string
	resolved     bool
}

func (d *definition) fixEnumsTypes() {
	fixEnumsTypesInSchema(d.schema)
}

// fixEnumsTypesInSchema checks if schema desription has openAPIEnum annotation
// in it, and replaces it's type from `integer` to `string`. This is needed to
// properly handle enums from proto. See README for more info.
func fixEnumsTypesInSchema(schema *extensionsv1.JSONSchemaProps) {
	if strings.Contains(schema.Description, openAPIEnum) {
		schema.Description = strings.Trim(
			strings.ReplaceAll(schema.Description, openAPIEnum, ""), " ")
		if schema.Type == "integer" {
			schema.Type = "string"
			schema.Format = ""
			schema.Default = nil
		}
		if schema.Type == "array" {
			if schema.Items != nil {
				if schema.Items.Schema != nil {
					if schema.Items.Schema.Type == "integer" {
						schema.Items.Schema.Type = "string"
						schema.Items.Schema.Format = ""
						schema.Items.Schema.Default = nil
					}
					fixEnumsTypesInSchema(schema.Items.Schema)
				}
				if len(schema.Items.JSONSchemas) != 0 {
					toReplace := make([]extensionsv1.JSONSchemaProps, len(schema.Items.JSONSchemas))
					for i, s := range schema.Items.JSONSchemas {
						if s.Type == "integer" {
							s.Type = "string"
							s.Format = ""
							s.Default = nil
						}
						fixEnumsTypesInSchema(&s)
						toReplace[i] = s
					}
					schema.Items.JSONSchemas = toReplace
				}
			}
		}
	}
	for name := range schema.Properties {
		prop := schema.Properties[name]
		fixEnumsTypesInSchema(&prop)
		schema.Properties[name] = prop
	}
	additionalProperties := schema.AdditionalProperties
	if additionalProperties != nil {
		if additionalProperties.Schema != nil {
			fixEnumsTypesInSchema(additionalProperties.Schema)
		}
	}
	additionalItems := schema.AdditionalItems
	if additionalItems != nil {
		if additionalItems.Schema != nil {
			fixEnumsTypesInSchema(additionalItems.Schema)
		}
	}
}

// NewGenerator returns a new Generator instance and an error.
// rawDefinitions is a value returned by GetOpenAPIDefinitions functions created
// by openapi-gen from k8s.io/kube-openapi.
func NewGenerator(rawDefinitions map[string]common.OpenAPIDefinition) (*Generator, error) {
	for k, v := range defaultSchemas() {
		rawDefinitions[strings.ToLower(k)] = v
	}
	definitions := make(map[string]*definition, len(rawDefinitions))
	for name, rawDef := range rawDefinitions {
		definition, err := convertDefinition(rawDef)
		if err != nil {
			return nil, fmt.Errorf("converting schema %q: %v", name, err)
		}
		definition.fixEnumsTypes()
		definitions[strings.ToLower(name)] = definition
	}
	return &Generator{
		definitions:        definitions,
		missingDefinitions: map[string]struct{}{},
		namePrefix:         defaultNamePrefix,
	}, nil
}

func convertDefinition(input common.OpenAPIDefinition) (*definition, error) {
	serialized, err := json.Marshal(input.Schema.SchemaProps)
	if err != nil {
		return nil, fmt.Errorf("serializing schema: %v", err)
	}
	var schemaProps extensionsv1.JSONSchemaProps
	err = json.Unmarshal(serialized, &schemaProps)
	if err != nil {
		return nil, fmt.Errorf("deserializing schema: %v", err)
	}
	dependencies := make([]string, len(input.Dependencies))
	for i, dep := range input.Dependencies {
		dependencies[i] = strings.ToLower(dep)
	}
	return &definition{
		schema:       &schemaProps,
		dependencies: dependencies,
	}, nil
}

type Generator struct {
	definitions        map[string]*definition
	missingDefinitions map[string]struct{}
	namePrefix         string
}

// MissingDefinitions is a simple getter for g.missingDefinitions.
func (g *Generator) MissingDefinitions() map[string]struct{} {
	return g.missingDefinitions
}

// SetNamePrefix sets the namePrefix used in name creation.
// This should not be needed in day to day usage, useful for testing.
func (g *Generator) SetNamePrefix(namePrefix string) {
	g.namePrefix = namePrefix
}

// AddDefinition is a helper for adding default definitions.
// This should not be needed in day to day usage, useful for testing.
func (g *Generator) AddDefinition(name string, rawDef common.OpenAPIDefinition) error {
	def, err := convertDefinition(rawDef)
	if err != nil {
		return err
	}
	g.definitions[name] = def
	return nil
}

// ResolveRefs calls resolveRefsForPackage for each package which has not been
// resolved yet
func (g *Generator) ResolveRefs() error {
	for pkg, schema := range g.definitions {
		if schema.resolved {
			continue
		}
		err := g.resolveRefsForPackage(pkg)
		if err != nil {
			return fmt.Errorf("resolving refs for package %q: %v", pkg, err)
		}
	}
	return nil
}

func (g *Generator) resolveRefsForPackage(pkg string) error {
	pkgSchema := g.getDefinition(pkg)
	if pkgSchema.resolved {
		return nil
	}

	// We need to have all the subtrees resolved not to miss something
	for _, dep := range pkgSchema.dependencies {
		err := g.resolveRefsForPackage(dep)
		if err != nil {
			return fmt.Errorf("resolving dependency %v for schema %v: %v", dep, pkg, err)
		}
	}

	fmt.Printf("Resolving refs for %v\n", pkg)

	g.resolveRefsInProperty(pkgSchema.schema)
	g.resolveRefsInProperties(pkgSchema.schema)
	g.resolveRefsInAdditionalProperties(pkgSchema.schema)

	pkgSchema.resolved = true
	return nil
}

func (g *Generator) getDefinition(pkg string) *definition {
	pkg = strings.ToLower(pkg)
	_, ok := g.definitions[pkg]
	if !ok {
		// g.missingDefinitions[pkg] = struct{}{}
		fmt.Printf("Missing schema for %v, using default\n", pkg)
		g.definitions[pkg] = g.definitions["default"]
	}
	def := g.definitions[pkg]
	g.addKubernetesExtensionsFlags(def.schema)
	return def
}

// addKubernetesExtensionsFlags adds the following flags:
//   - x-kubernetes-preserve-unknown-fields - required properties of type object
//     with no properties specified
//   - x-kubernetes-int-or-string - required for properties which use anyOf with
//     integer or string values
func (g *Generator) addKubernetesExtensionsFlags(schema *extensionsv1.JSONSchemaProps) {
	if len(schema.Properties) == 0 && schema.AdditionalProperties == nil && schema.Type == "object" {
		t := true
		schema.XPreserveUnknownFields = &t
	}
	// any is an alias for struct{}
	// - x-kubernetes-preserve-unknown-fields: true
	if schema.Title == "any" {
		t := true
		schema.XPreserveUnknownFields = &t
		schema.Title = "" // reset Title
	}

	if len(schema.AnyOf) > 0 {
		schema.XIntOrString = true
	}
	for name, prop := range schema.Properties {
		g.addKubernetesExtensionsFlags(&prop)
		schema.Properties[name] = prop
	}
	if schema.Items != nil {
		if schema.Items.Schema != nil {
			g.addKubernetesExtensionsFlags(schema.Items.Schema)
		}
		for arrName, arrSchema := range schema.Items.JSONSchemas {
			g.addKubernetesExtensionsFlags(&arrSchema)
			schema.Items.JSONSchemas[arrName] = arrSchema
		}
	}
}

// resolveRefsInProperties checks if any of the properties' schemas are refs
// and replaces them with proper schema
func (g *Generator) resolveRefsInProperties(schema *extensionsv1.JSONSchemaProps) {
	toInline := map[string]extensionsv1.JSONSchemaProps{}
	toDelete := map[string]struct{}{}
	for property, propSchema := range schema.Properties {
		if propSchema.Ref != nil {
			refSchema := g.getDefinition(*propSchema.Ref).schema
			if _, ok := refSchema.Properties[inline]; ok {
				toDelete[inline] = struct{}{}
				for k, v := range refSchema.Properties {
					toInline[k] = v
				}
				toDelete[property] = struct{}{}
			} else {
				schema.Properties[property] = *refSchema
			}
		}
		g.resolveRefsInProperty(&propSchema)
		g.resolveRefsInProperties(&propSchema)
		g.resolveRefsInAdditionalProperties(&propSchema)
	}
	for k, v := range toInline {
		schema.Properties[k] = v
	}
	for v := range toDelete {
		delete(schema.Properties, v)
	}
	required := []string{}
	for _, v := range schema.Required {
		if _, ok := toDelete[v]; ok {
			continue
		}
		required = append(required, v)
	}
	schema.Required = required
}

// resolveRefsInAdditionalProperties checks if any of the additionalProperties'
// schemas are refs and replaces them with proper schema
func (g *Generator) resolveRefsInAdditionalProperties(schema *extensionsv1.JSONSchemaProps) {
	additionalProperties := schema.AdditionalProperties
	if additionalProperties != nil {
		if additionalProperties.Schema != nil {
			if additionalProperties.Schema.Ref != nil {
				additionalProperties.Schema = g.getDefinition(*additionalProperties.Schema.Ref).schema
			}
			g.resolveRefsInProperty(additionalProperties.Schema)
			g.resolveRefsInProperties(additionalProperties.Schema)
			g.resolveRefsInAdditionalProperties(additionalProperties.Schema)
		}
	}

	additionalItems := schema.AdditionalItems
	if additionalItems != nil {
		if additionalItems.Schema != nil {
			if additionalItems.Schema.Ref != nil {
				additionalItems.Schema = g.getDefinition(*additionalItems.Schema.Ref).schema
			}
			g.resolveRefsInProperty(additionalItems.Schema)
			g.resolveRefsInProperties(additionalItems.Schema)
			g.resolveRefsInAdditionalProperties(additionalItems.Schema)
		}
	}
}

// resolveRefsInProperty checks if property.Items schemas are refs and replaces them
// with proper schema
func (g *Generator) resolveRefsInProperty(propSchema *extensionsv1.JSONSchemaProps) {
	if propSchema.Items != nil {
		// in Items we can have either single schema defined in Schema, or a list of
		// them in JSONSchemas
		if propSchema.Items.Schema != nil {
			if propSchema.Items.Schema.Ref != nil {
				*propSchema.Items.Schema = *g.getDefinition(*propSchema.Items.Schema.Ref).schema
			}
		}
		newSchemas := make([]extensionsv1.JSONSchemaProps, len(propSchema.Items.JSONSchemas))
		for i, schema := range propSchema.Items.JSONSchemas {
			if schema.Ref == nil {
				newSchemas[i] = schema
				continue
			}
			newSchemas[i] = *g.getDefinition(*schema.Ref).schema
		}
		propSchema.Items.JSONSchemas = newSchemas
	}
}

func (g *Generator) UpdateYAMLs(yamlsPath string) error {
	return filepath.Walk(yamlsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walking files: %v", err)
		}

		if info.IsDir() {
			fmt.Printf("Skipping dir %q\n", path)
			return nil
		}

		if !strings.HasSuffix(path, ".yaml") {
			fmt.Printf("Skipping non-yaml file %q\n", path)
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading file %q: %v", path, err)
		}

		parts := strings.Split(string(content), "---")
		crds := make([]extensionsv1.CustomResourceDefinition, len(parts))
		for _, part := range parts {
			if part == "" {
				continue
			}
			var crd extensionsv1.CustomResourceDefinition
			err := yaml.Unmarshal([]byte(part), &crd)
			if err != nil {
				return fmt.Errorf("unmarshalling: %v", err)
			}
			if len(crd.Spec.Versions) == 0 {
				return fmt.Errorf("crd %v has no versions", crd.Name)
			}
			// TODO another hack. I don't know how to prevent Status field being
			// generated in YAML file, so we add StoredVersions by hand
			var storedVersions []string
			for _, v := range crd.Spec.Versions {
				if v.Storage {
					storedVersions = append(storedVersions, v.Name)
				}
			}
			if len(storedVersions) != 1 {
				return fmt.Errorf("crd %v must have exactly one storage version, found %d", crd.Name, len(storedVersions))
			}
			crd.Status.StoredVersions = storedVersions
			err = g.addCustomResourceValidation(&crd)
			if err != nil {
				return err
			}
			crds = append(crds, crd)
		}

		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return fmt.Errorf("opening file %q for write: %v", path, err)
		}
		defer f.Close()
		header := `# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0
`
		for _, crd := range crds {
			// TODO IDK why this happens. It this out for now.
			if crd.Name == "" {
				continue
			}
			fmt.Printf("Writing schema %q to %q\n", crd.Name, path)
			_, err := f.Write([]byte("---\n"))
			if err != nil {
				return fmt.Errorf("writing separator: %v", err)
			}
			serialized, err := yaml.Marshal(crd)
			if err != nil {
				return fmt.Errorf("serializing crd %q: %v", crd.Name, err)
			}
			// Prepend the header
			finalOutput := []byte(header + string(serialized))
			_, err = f.Write(finalOutput)
			if err != nil {
				return fmt.Errorf("writing crd %q: %v", crd.Name, err)
			}
		}
		return nil
	})
}

func (g *Generator) addCustomResourceValidation(crd *extensionsv1.CustomResourceDefinition) error {
	crd.Spec.PreserveUnknownFields = false
	// Each version now has its own generated Go type (apis/<group>/<version>
	// package - see pkg/generator/types_generator.go and
	// RenderTypesTemplate), so each gets its own schema instead of sharing
	// the storage version's. getDefinition() falls back to a generic
	// "default" schema (with a warning) if a version's openapi-gen output
	// can't be found, so this is safe even for datamodels whose build
	// pipeline doesn't run openapi-gen against every version's package yet.
	schemas := make([]*extensionsv1.JSONSchemaProps, len(crd.Spec.Versions))
	storageIdx := -1
	for i := range crd.Spec.Versions {
		name := g.createName(crd.Spec.Group, crd.Spec.Versions[i].Name, crd.Spec.Names.Kind)
		schemaProps := g.getDefinition(name).schema
		toReplace := []string{}
		for _, propName := range schemaProps.Required {
			// We want only "spec" and "data" fields required in schemas. "status" and all
			// others should remain optional.
			if propName == "spec" || propName == "data" {
				toReplace = append(toReplace, propName)
			}
		}
		schemaProps.Required = toReplace
		// We need to ensure that crd.spec.versions[*].schema.openapiV3Schema["metadata"]
		// does not contain any fields except for `type: object`. This is required by K8s
		// API and we added x-kubernetes-preserve-unknown-fields: true` here in `getDefinition`,
		// so we are fixing our wrongly added flag.
		if len(schemaProps.Properties) == 0 {
			schemaProps.Properties = map[string]extensionsv1.JSONSchemaProps{}
		}
		schemaProps.Properties["metadata"] = extensionsv1.JSONSchemaProps{
			Type: "object",
		}

		schemas[i] = schemaProps
		if crd.Spec.Versions[i].Storage {
			storageIdx = i
		}
	}

	// A multi-version Kind's conversion webhook (see
	// orch-utils/nexus/api-gw/pkg/server/echoserver/conversion_handler.go)
	// is a generic JSON pass-through - it doesn't know this datamodel's
	// field semantics, so it can't re-map a field that exists only in a
	// non-storage version into the storage version's shape. That wouldn't
	// matter if the pass-through's output only needed to round-trip through
	// *validation*, but the API server also *prunes* any field not declared
	// in the storage version's schema before persisting - independent of
	// and after whatever the webhook returns. So a field declared only on
	// v2 (say) gets silently dropped the moment it's written, even though
	// the webhook faithfully passed it through. Union every version's
	// declared fields into the storage version's schema so none of them
	// get pruned; every non-storage version keeps its own (narrower, only
	// for documentation/validation purposes) schema unchanged.
	if storageIdx >= 0 && len(schemas) > 1 {
		for i, schemaProps := range schemas {
			if i == storageIdx {
				continue
			}
			unionSchemaProperties(schemas[storageIdx], schemaProps)
		}
	}

	for i := range crd.Spec.Versions {
		crd.Spec.Versions[i].Schema = &extensionsv1.CustomResourceValidation{OpenAPIV3Schema: schemas[i]}
	}
	return nil
}

// unionSchemaProperties recursively merges src's declared properties into
// dst wherever dst doesn't already declare them, so dst (the storage
// version's schema) ends up a superset of every version's fields. On a
// name collision, dst's own property definition wins (its version's type
// is assumed authoritative for that field) but is still merged one level
// deeper in case nested sub-properties differ.
func unionSchemaProperties(dst, src *extensionsv1.JSONSchemaProps) {
	if dst == nil || src == nil || len(src.Properties) == 0 {
		return
	}
	if dst.Properties == nil {
		dst.Properties = map[string]extensionsv1.JSONSchemaProps{}
	}
	for name, srcProp := range src.Properties {
		dstProp, ok := dst.Properties[name]
		if !ok {
			dst.Properties[name] = srcProp
			continue
		}
		unionSchemaProperties(&dstProp, &srcProp)
		dst.Properties[name] = dstProp
	}
}

func (g *Generator) createName(group, apiVersion, name string) string {
	return strings.ToLower(fmt.Sprintf("%s/%s/%s.%s", g.namePrefix, group, apiVersion, name))
}

//func convertToCamelCase(input string) string {
//	if !strings.Contains(input, "_") {
//		return input
//	}
//	parts := strings.Split(input, "_")
//	camelCaseName := parts[0]
//	for _, p := range parts[1:] {
//		camelCaseName += strings.ToUpper(string(p[0]))
//		camelCaseName += p[1:]
//	}
//	return camelCaseName
//}
