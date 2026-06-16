// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"encoding/json"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// buildComponentsForNode populates `components` with the schemas,
// request bodies, and responses derived from a single node's CRD
// schema. The component-key scheme matches today's emission exactly:
//
//	<name>.Get        — spec + status object
//	<name>.Post       — spec object (request body)
//	<name>.List       — array of {name, spec, status}
//	<name>.Status     — status object only
//	<name>.SingleLink — empty object (placeholder; today's behavior)
//	<name>.NamedLink  — array of empty objects (placeholder; today's behavior)
//
// Where <name> is the NodeInfo.Name (e.g. "users.User") — NOT the
// qualifiedKind. Component keys never collide because they include the
// package segment.
//
// If info.Schema is nil this function is a no-op. Adapter code is
// responsible for providing the schema; missing schemas are tolerated
// silently because some synthetic nodes legitimately have none.
func buildComponentsForNode(components *openapi3.Components, info NodeInfo) {
	if info.Schema == nil {
		return
	}

	getKey := componentKey(info.Name, "Get")
	postKey := componentKey(info.Name, "Post")
	listKey := componentKey(info.Name, "List")
	statusKey := componentKey(info.Name, "Status")
	singleLinkKey := componentKey(info.Name, "SingleLink")
	namedLinkKey := componentKey(info.Name, "NamedLink")

	specProps := info.Schema.Properties["spec"].Properties
	jsonSpecSchema := openapi3.NewObjectSchema()
	parseFields(jsonSpecSchema, specProps)

	statusProps := info.Schema.Properties["status"].Properties
	delete(statusProps, "nexus")
	jsonStatusSchema := openapi3.NewObjectSchema()
	parseFields(jsonStatusSchema, statusProps)

	components.Schemas[statusKey] = openapi3.NewSchemaRef("", jsonStatusSchema)

	jsonSpecAndStatusSchema := openapi3.NewObjectSchema()
	jsonSpecAndStatusSchema.WithProperty("spec", jsonSpecSchema)
	jsonSpecAndStatusSchema.WithProperty("status", jsonStatusSchema)

	components.Schemas[postKey] = openapi3.NewSchemaRef("", jsonSpecSchema)
	components.Schemas[getKey] = openapi3.NewSchemaRef("", jsonSpecAndStatusSchema)

	jsonListObjectSchema := openapi3.NewObjectSchema()
	jsonListObjectSchema.WithProperty("name", openapi3.NewStringSchema())
	jsonListObjectSchema.WithProperty("spec", jsonSpecSchema)
	jsonListObjectSchema.WithProperty("status", jsonStatusSchema)
	jsonListSchema := openapi3.NewArraySchema().WithItems(jsonListObjectSchema)

	components.Schemas[listKey] = openapi3.NewSchemaRef("", jsonListSchema)

	// TODO: real schemas for single-link and named-link traversal.
	// Today both consumers emit empty placeholders; we preserve that.
	jsonSingleLinkSchema := openapi3.NewObjectSchema()
	jsonNamedLinkSchema := openapi3.NewArraySchema().WithItems(jsonSingleLinkSchema)
	components.Schemas[singleLinkKey] = openapi3.NewSchemaRef("", jsonSingleLinkSchema)
	components.Schemas[namedLinkKey] = openapi3.NewSchemaRef("", jsonNamedLinkSchema)

	components.RequestBodies["Create"+info.Name] = &openapi3.RequestBodyRef{
		Value: openapi3.NewRequestBody().
			WithDescription("Request used to create " + info.Name).
			WithRequired(true).
			WithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + postKey}),
	}

	components.Responses["Get"+info.Name] = &openapi3.ResponseRef{
		Value: openapi3.NewResponse().
			WithDescription("Response returned back after getting " + info.Name + " object").
			WithContent(
				openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + getKey}),
			),
	}

	components.RequestBodies["Create"+statusKey] = &openapi3.RequestBodyRef{
		Value: openapi3.NewRequestBody().
			WithDescription("Request used to create Status subresource of " + info.Name).
			WithRequired(false).
			WithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + statusKey}),
	}

	components.Responses["Get"+statusKey] = &openapi3.ResponseRef{
		Value: openapi3.NewResponse().
			WithDescription("Response returned back after getting status subresource of " + info.Name + " object").
			WithContent(
				openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + statusKey}),
			),
	}

	components.Responses["List"+info.Name] = &openapi3.ResponseRef{
		Value: openapi3.NewResponse().
			WithDescription("Response returned back after getting " + info.Name + " objects").
			WithContent(
				openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + listKey}),
			),
	}

	components.Responses["Get"+singleLinkKey] = &openapi3.ResponseRef{
		Value: openapi3.NewResponse().
			WithDescription("Response returned back after getting " + info.Name + " objects").
			WithContent(
				openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + singleLinkKey}),
			),
	}

	components.Responses["Get"+namedLinkKey] = &openapi3.ResponseRef{
		Value: openapi3.NewResponse().
			WithDescription("Response returned back after getting " + info.Name + " objects").
			WithContent(
				openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + namedLinkKey}),
			),
	}
}

// componentKey returns the canonical "<name>.<keyType>" used throughout
// the spec components map and the corresponding $ref strings.
func componentKey(name, keyType string) string {
	return name + "." + keyType
}

// parseFields walks a JSONSchemaProps property map and attaches each
// non-Gvk field to `jsonSchema`, propagating description and example
// metadata where present.
func parseFields(jsonSchema *openapi3.Schema, props map[string]apiextensionsv1.JSONSchemaProps) {
	for name, prop := range props {
		if strings.Contains(name, "Gvk") {
			continue
		}
		schema := buildPropSchema(prop)
		if schema == nil {
			continue
		}
		if prop.Description != "" {
			schema.Description = prop.Description
		}
		if prop.Example != nil {
			var ex interface{}
			if err := json.Unmarshal(prop.Example.Raw, &ex); err == nil {
				schema.Example = ex
			}
		}
		jsonSchema.WithProperty(name, schema)
	}
}

// buildPropSchema converts a single JSONSchemaProps into an
// openapi3.Schema. Handles primitives, objects (including maps via
// additionalProperties), and arrays. Description/example are attached
// by the caller (parseFields) so the same helper can recurse into
// nested items and additionalProperties.
//
// Returns nil for unrecognised types; callers skip nil schemas. We do
// not log here because the builder is logger-agnostic; adapters log
// via their own loggers if they wish to observe missing types.
func buildPropSchema(prop apiextensionsv1.JSONSchemaProps) *openapi3.Schema {
	switch prop.Type {
	case "string":
		switch prop.Format {
		case "byte":
			return openapi3.NewBytesSchema()
		case "date-time":
			return openapi3.NewDateTimeSchema()
		default:
			return openapi3.NewStringSchema()
		}
	case "boolean":
		return openapi3.NewBoolSchema()
	case "integer":
		switch prop.Format {
		case "int32":
			return openapi3.NewInt32Schema()
		case "int64":
			return openapi3.NewInt64Schema()
		default:
			return openapi3.NewIntegerSchema()
		}
	case "number":
		return openapi3.NewFloat64Schema()
	case "object":
		schema := openapi3.NewObjectSchema()
		if prop.AdditionalProperties != nil && prop.AdditionalProperties.Schema != nil {
			schema.WithAdditionalProperties(buildPropSchema(*prop.AdditionalProperties.Schema))
		} else {
			parseFields(schema, prop.Properties)
		}
		return schema
	case "array":
		if prop.Items != nil && prop.Items.Schema != nil {
			return openapi3.NewArraySchema().WithItems(buildPropSchema(*prop.Items.Schema))
		}
		return openapi3.NewArraySchema()
	default:
		return nil
	}
}
