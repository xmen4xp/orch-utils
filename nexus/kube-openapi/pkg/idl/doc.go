// Copyright 2018 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// The IDL package describes comment directives that may be applied to
// API types and fields.
package idl

// ListType annotates a list to further describe its topology. It may
// have 3 possible values: "atomic", "map", or "set". Note that there is
// no default, and the generation step will fail if a list is found that
// is missing the tag.
//
// This tag MUST only be used on lists, or the generation step will
// fail.
//
// # Atomic
//
// Example:
//
//	+listType=atomic
//
// Atomic lists will be entirely replaced when updated. This tag may be
// used on any type of list (struct, scalar, ...).
//
// Using this tag will generate the following OpenAPI extension:
//
//	"x-kubernetes-list-type": "atomic"
//
// # Map
//
// Example:
//
//	+listType=map
//
// These lists are like maps in that their elements have a non-index key
// used to identify them. Order is preserved upon merge. Using the map
// tag on a list with non-struct elements will result in an error during
// the generation step.
//
// Using this tag will generate the following OpenAPI extension:
//
//	"x-kubernetes-list-type": "map"
//
// # Set
//
// Example:
//
//	+listType=set
//
// Sets are lists that must not have multiple times the same value. Each
// value must be a scalar (or another atomic type).
//
// Using this tag will generate the following OpenAPI extension:
//
//	"x-kubernetes-list-type": "set"
type ListType string

// ListMapKey annotates map lists by specifying the key used as the index of the map.
//
// This tag MUST only be used on lists that have the listType=map
// attribute, or the generation step will fail. Also, the value
// specified for this attribute must be a scalar typed field of the
// child structure (no nesting is supported).
//
// An example of how this can be used is shown in the ListType (map) example.
//
// Example:
//
//	+listMapKey=name
//
// Using this tag will generate the following OpenAPI extension:
//
//	"x-kubernetes-list-map-key": "name"
type ListMapKey string

// MapType annotates a map to further describe its topology. It may
// have one of two values: `atomic` or `granular`. `atomic` means that the entire map is
// considered as a whole; actors that wish to update the map can only
// entirely replace it. `granular` means that specific values in the map can be
// updated separately from other fields.
//
// By default, a map will be considered as a set of distinct values that
// can be updated individually (i.e. the equivalent of `granular`).
// This default will still generate an OpenAPI extension with key: "x-kubernetes-map-type".
//
// This tag MUST only be used on maps, or the generation step will fail.
//
// # Atomic
//
// Example:
//
//	+mapType=atomic
//
// Atomic maps will be entirely replaced when updated. This tag may be
// used on any map.
//
// Using this tag will generate the following OpenAPI extension:
//
//	"x-kubernetes-map-type": "atomic"
type MapType string

// OpenAPIGen needs to be described.
type OpenAPIGen string

// Optional annotates a field to specify it may be omitted.
// By default, fields will be marked as required if not otherwise specified.
//
// Example:
//
//	+optional
//
// Additionally, the json struct tag directive "omitempty" can be used to imply
// the same.
//
// Example:
//
//	OptionalField `json:"optionalField,omitempty"`
type Optional string

// PatchMergeKey needs to be described.
type PatchMergeKey string

// PatchStrategy needs to be described.
type PatchStrategy string

// StructType annotates a struct to further describe its topology. It may
// have one of two values: `atomic` or `granular`. `atomic` means that the entire struct is
// considered as a whole; actors that wish to update the struct can only
// entirely replace it. `granular` means that specific fields in the struct can be
// updated separately from other fields.
//
// By default, a struct will be considered as a set of distinct values that
// can be updated individually (`granular`).
// This default will still generate an OpenAPI extension with key: "x-kubernetes-map-type".
//
// This tag MUST only be used on structs, or the generation step will fail.
//
// # Atomic
//
// Example:
//
//	+structType=atomic
//
// Atomic structs will be entirely replaced when updated. This tag may be
// used on any struct.
//
// Using this tag will generate the following OpenAPI extension:
//
//	"x-kubernetes-map-type": "atomic"
type StructType string

// Union is TBD.
type Union string
