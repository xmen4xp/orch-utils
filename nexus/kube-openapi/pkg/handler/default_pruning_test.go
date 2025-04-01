// Copyright 2020 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package handler_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/handler"

	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/validation/spec"
)

func TestDefaultPruning(t *testing.T) {
	def := spec.Definitions{
		"foo": spec.Schema{
			SchemaProps: spec.SchemaProps{
				Default: 0,
				AllOf:   []spec.Schema{spec.Schema{SchemaProps: spec.SchemaProps{Default: "default-string", Title: "Field"}}},
				AnyOf:   []spec.Schema{spec.Schema{SchemaProps: spec.SchemaProps{Default: "default-string", Title: "Field"}}},
				OneOf:   []spec.Schema{spec.Schema{SchemaProps: spec.SchemaProps{Default: "default-string", Title: "Field"}}},
				Not:     &spec.Schema{SchemaProps: spec.SchemaProps{Default: "default-string", Title: "Field"}},
				Properties: map[string]spec.Schema{
					"foo": spec.Schema{SchemaProps: spec.SchemaProps{Default: "default-string", Title: "Field"}},
				},
				AdditionalProperties: &spec.SchemaOrBool{Schema: &spec.Schema{SchemaProps: spec.SchemaProps{Default: "default-string", Title: "Field"}}},
				PatternProperties: map[string]spec.Schema{
					"foo": spec.Schema{SchemaProps: spec.SchemaProps{Default: "default-string", Title: "Field"}},
				},
				Dependencies: spec.Dependencies{
					"foo": spec.SchemaOrStringArray{Schema: &spec.Schema{SchemaProps: spec.SchemaProps{Default: "default-string", Title: "Field"}}},
				},
				AdditionalItems: &spec.SchemaOrBool{
					Schema: &spec.Schema{SchemaProps: spec.SchemaProps{Default: "default-string", Title: "Field"}},
				},
				Definitions: spec.Definitions{
					"bar": spec.Schema{SchemaProps: spec.SchemaProps{Default: "default-string", Title: "Field"}},
				},
			},
		},
	}
	jsonDef, err := json.Marshal(def)
	if err != nil {
		t.Fatalf("Failed to marshal definition: %v", err)
	}
	wanted := spec.Definitions{
		"foo": spec.Schema{
			SchemaProps: spec.SchemaProps{
				AllOf: []spec.Schema{spec.Schema{SchemaProps: spec.SchemaProps{Title: "Field"}}},
				AnyOf: []spec.Schema{spec.Schema{SchemaProps: spec.SchemaProps{Title: "Field"}}},
				OneOf: []spec.Schema{spec.Schema{SchemaProps: spec.SchemaProps{Title: "Field"}}},
				Not:   &spec.Schema{SchemaProps: spec.SchemaProps{Title: "Field"}},
				Properties: map[string]spec.Schema{
					"foo": spec.Schema{SchemaProps: spec.SchemaProps{Title: "Field"}},
				},
				AdditionalProperties: &spec.SchemaOrBool{Schema: &spec.Schema{SchemaProps: spec.SchemaProps{Title: "Field"}}},
				PatternProperties: map[string]spec.Schema{
					"foo": spec.Schema{SchemaProps: spec.SchemaProps{Title: "Field"}},
				},
				Dependencies: spec.Dependencies{
					"foo": spec.SchemaOrStringArray{Schema: &spec.Schema{SchemaProps: spec.SchemaProps{Title: "Field"}}},
				},
				AdditionalItems: &spec.SchemaOrBool{
					Schema: &spec.Schema{SchemaProps: spec.SchemaProps{Title: "Field"}},
				},
				Definitions: spec.Definitions{
					"bar": spec.Schema{SchemaProps: spec.SchemaProps{Title: "Field"}},
				},
			},
		},
	}

	got := handler.PruneDefaults(def)
	if !reflect.DeepEqual(got, wanted) {
		gotJSON, _ := json.Marshal(got)
		wantedJSON, _ := json.Marshal(wanted)
		t.Fatalf("got: %v\nwanted %v", string(gotJSON), string(wantedJSON))
	}
	// Make sure that def hasn't been changed.
	newDef, _ := json.Marshal(def)
	if string(newDef) != string(jsonDef) {
		t.Fatalf("prune removed defaults from initial config:\nBefore: %v\nAfter: %v", string(jsonDef), string(newDef))
	}
	// Make sure that no-op doesn't change the object.
	if reflect.ValueOf(handler.PruneDefaults(got)).Pointer() != reflect.ValueOf(got).Pointer() {
		t.Fatal("no-op prune returned new object")
	}
}
