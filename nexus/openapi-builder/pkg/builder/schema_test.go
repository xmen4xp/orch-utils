// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// newComponents returns a freshly-initialised openapi3.Components with
// all maps allocated, matching what `Build()` will use in production.
func newComponents() *openapi3.Components {
	return &openapi3.Components{
		Schemas:       openapi3.Schemas{},
		RequestBodies: openapi3.RequestBodies{},
		Responses:     openapi3.ResponseBodies{},
	}
}

func TestBuildComponentsForNode_NilSchemaIsNoOp(t *testing.T) {
	c := newComponents()
	buildComponentsForNode(c, NodeInfo{Name: "orgs.Org", Schema: nil})

	if len(c.Schemas) != 0 || len(c.RequestBodies) != 0 || len(c.Responses) != 0 {
		t.Errorf("expected nil schema to be a no-op; got schemas=%d req=%d resp=%d",
			len(c.Schemas), len(c.RequestBodies), len(c.Responses))
	}
}

func TestBuildComponentsForNode_EmitsAllKeys(t *testing.T) {
	info := NodeInfo{
		Name: "orgs.Org",
		Schema: &apiextensionsv1.JSONSchemaProps{
			Type: "object",
			Properties: map[string]apiextensionsv1.JSONSchemaProps{
				"spec": {
					Type: "object",
					Properties: map[string]apiextensionsv1.JSONSchemaProps{
						"name":        {Type: "string", Description: "Organization name"},
						"description": {Type: "string"},
						"capacity":    {Type: "integer", Format: "int32"},
					},
				},
				"status": {
					Type: "object",
					Properties: map[string]apiextensionsv1.JSONSchemaProps{
						"ready": {Type: "boolean"},
						"nexus": {Type: "object"}, // must be stripped
					},
				},
			},
		},
	}

	c := newComponents()
	buildComponentsForNode(c, info)

	// All six component keys present.
	for _, key := range []string{
		"orgs.Org.Get",
		"orgs.Org.Post",
		"orgs.Org.List",
		"orgs.Org.Status",
		"orgs.Org.SingleLink",
		"orgs.Org.NamedLink",
	} {
		if _, ok := c.Schemas[key]; !ok {
			t.Errorf("missing schema component %q", key)
		}
	}

	// Request bodies and responses keyed by node name.
	if _, ok := c.RequestBodies["Createorgs.Org"]; !ok {
		t.Error("missing request body Createorgs.Org")
	}
	if _, ok := c.RequestBodies["Createorgs.Org.Status"]; !ok {
		t.Error("missing request body Createorgs.Org.Status")
	}
	for _, key := range []string{
		"Getorgs.Org",
		"Getorgs.Org.Status",
		"Listorgs.Org",
		"Getorgs.Org.SingleLink",
		"Getorgs.Org.NamedLink",
	} {
		if _, ok := c.Responses[key]; !ok {
			t.Errorf("missing response %q", key)
		}
	}

	// Status schema must NOT contain the `nexus` property — we strip it
	// explicitly. The status object should have only `ready`.
	statusSchema := c.Schemas["orgs.Org.Status"].Value
	if _, hasNexus := statusSchema.Properties["nexus"]; hasNexus {
		t.Error("status schema unexpectedly contains the `nexus` property")
	}
	if _, hasReady := statusSchema.Properties["ready"]; !hasReady {
		t.Error("status schema missing `ready` property")
	}

	// Description from spec.name propagates onto the schema.
	postSchema := c.Schemas["orgs.Org.Post"].Value
	nameProp := postSchema.Properties["name"]
	if nameProp == nil || nameProp.Value == nil {
		t.Fatal("post schema missing `name` property")
	}
	if nameProp.Value.Description != "Organization name" {
		t.Errorf("expected description propagation, got %q", nameProp.Value.Description)
	}
}

func TestBuildPropSchema_AllTypes(t *testing.T) {
	cases := []struct {
		name    string
		prop    apiextensionsv1.JSONSchemaProps
		wantNil bool
	}{
		{"string", apiextensionsv1.JSONSchemaProps{Type: "string"}, false},
		{"string-byte", apiextensionsv1.JSONSchemaProps{Type: "string", Format: "byte"}, false},
		{"string-datetime", apiextensionsv1.JSONSchemaProps{Type: "string", Format: "date-time"}, false},
		{"boolean", apiextensionsv1.JSONSchemaProps{Type: "boolean"}, false},
		{"integer", apiextensionsv1.JSONSchemaProps{Type: "integer"}, false},
		{"integer-int32", apiextensionsv1.JSONSchemaProps{Type: "integer", Format: "int32"}, false},
		{"integer-int64", apiextensionsv1.JSONSchemaProps{Type: "integer", Format: "int64"}, false},
		{"number", apiextensionsv1.JSONSchemaProps{Type: "number"}, false},
		{"object", apiextensionsv1.JSONSchemaProps{Type: "object"}, false},
		{"array", apiextensionsv1.JSONSchemaProps{Type: "array"}, false},
		{"unknown", apiextensionsv1.JSONSchemaProps{Type: "wat"}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := buildPropSchema(c.prop)
			if (got == nil) != c.wantNil {
				t.Errorf("buildPropSchema(%+v) nil=%v, want nil=%v", c.prop, got == nil, c.wantNil)
			}
		})
	}
}

func TestParseFields_SkipsGvk(t *testing.T) {
	schema := openapi3.NewObjectSchema()
	parseFields(schema, map[string]apiextensionsv1.JSONSchemaProps{
		"name":       {Type: "string"},
		"OrgGvk":     {Type: "object"}, // must be skipped (contains "Gvk")
		"ProjectGvk": {Type: "object"}, // must be skipped
	})
	if _, ok := schema.Properties["name"]; !ok {
		t.Error("expected `name` to be present")
	}
	for _, gvk := range []string{"OrgGvk", "ProjectGvk"} {
		if _, ok := schema.Properties[gvk]; ok {
			t.Errorf("expected %q to be skipped (contains `Gvk`)", gvk)
		}
	}
}

func TestParseFields_NestedObjectAndArray(t *testing.T) {
	schema := openapi3.NewObjectSchema()
	parseFields(schema, map[string]apiextensionsv1.JSONSchemaProps{
		"nested": {
			Type: "object",
			Properties: map[string]apiextensionsv1.JSONSchemaProps{
				"inner": {Type: "string"},
			},
		},
		"items": {
			Type: "array",
			Items: &apiextensionsv1.JSONSchemaPropsOrArray{
				Schema: &apiextensionsv1.JSONSchemaProps{Type: "string"},
			},
		},
	})

	nested := schema.Properties["nested"]
	if nested == nil || nested.Value == nil {
		t.Fatal("nested property missing")
	}
	if _, ok := nested.Value.Properties["inner"]; !ok {
		t.Error("nested.inner missing")
	}

	items := schema.Properties["items"]
	if items == nil || items.Value == nil {
		t.Fatal("items property missing")
	}
	if items.Value.Items == nil || items.Value.Items.Value == nil {
		t.Fatal("items.items schema missing")
	}
}
