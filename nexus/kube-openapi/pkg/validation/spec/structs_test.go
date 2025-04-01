// Copyright 2015 go-swagger maintainers
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertSerializeJSON(t testing.TB, actual interface{}, expected string) bool {
	ser, err := json.Marshal(actual)
	if err != nil {
		return assert.Fail(t, "unable to marshal to json (%s): %#v", err, actual)
	}
	return assert.Equal(t, string(ser), expected)
}

func derefTypeOf(expected interface{}) (tpe reflect.Type) {
	tpe = reflect.TypeOf(expected)
	if tpe.Kind() == reflect.Ptr {
		tpe = tpe.Elem()
	}
	return
}

func isPointed(expected interface{}) (pointed bool) {
	tpe := reflect.TypeOf(expected)
	if tpe.Kind() == reflect.Ptr {
		pointed = true
	}
	return
}

func assertParsesJSON(t testing.TB, actual string, expected interface{}) bool {
	parsed := reflect.New(derefTypeOf(expected))
	err := json.Unmarshal([]byte(actual), parsed.Interface())
	if err != nil {
		return assert.Fail(t, "unable to unmarshal from json (%s): %s", err, actual)
	}
	act := parsed.Interface()
	if !isPointed(expected) {
		act = reflect.Indirect(parsed).Interface()
	}
	return assert.Equal(t, act, expected)
}

func TestSerialization_SerializeJSON(t *testing.T) {
	assertSerializeJSON(t, []string{"hello"}, "[\"hello\"]")
	assertSerializeJSON(t, []string{"hello", "world", "and", "stuff"}, "[\"hello\",\"world\",\"and\",\"stuff\"]")
	assertSerializeJSON(t, StringOrArray(nil), "null")
	assertSerializeJSON(t, SchemaOrArray{
		Schemas: []Schema{
			{SchemaProps: SchemaProps{Type: []string{"string"}}}},
	}, "[{\"type\":\"string\"}]")
	assertSerializeJSON(t, SchemaOrArray{
		Schemas: []Schema{
			{SchemaProps: SchemaProps{Type: []string{"string"}}},
			{SchemaProps: SchemaProps{Type: []string{"string"}}},
		}}, "[{\"type\":\"string\"},{\"type\":\"string\"}]")
	assertSerializeJSON(t, SchemaOrArray{}, "null")
}

func TestSerialization_DeserializeJSON(t *testing.T) {
	// String
	assertParsesJSON(t, "\"hello\"", StringOrArray([]string{"hello"}))
	assertParsesJSON(t, "[\"hello\",\"world\",\"and\",\"stuff\"]",
		StringOrArray([]string{"hello", "world", "and", "stuff"}))
	assertParsesJSON(t, "[\"hello\",\"world\",null,\"stuff\"]", StringOrArray([]string{"hello", "world", "", "stuff"}))
	assertParsesJSON(t, "null", StringOrArray(nil))

	// Schema
	assertParsesJSON(t, "{\"type\":\"string\"}", SchemaOrArray{Schema: &Schema{
		SchemaProps: SchemaProps{Type: []string{"string"}}},
	})
	assertParsesJSON(t, "[{\"type\":\"string\"},{\"type\":\"string\"}]", &SchemaOrArray{
		Schemas: []Schema{
			{SchemaProps: SchemaProps{Type: []string{"string"}}},
			{SchemaProps: SchemaProps{Type: []string{"string"}}},
		},
	})
	assertParsesJSON(t, "null", SchemaOrArray{})
}
