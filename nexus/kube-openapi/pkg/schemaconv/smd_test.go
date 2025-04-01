// Copyright 2017 The Kubernetes Authors.
// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package schemaconv

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	yaml "gopkg.in/yaml.v2"

	"github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/util/proto"
	prototesting "github.com/vmware-tanzu/graph-framework-for-microservices/kube-openapi/pkg/util/proto/testing"
)

func TestToSchema(t *testing.T) {
	tests := []struct {
		name                   string
		openAPIFilename        string
		expectedSchemaFilename string
	}{
		{
			name:                   "kubernetes",
			openAPIFilename:        "swagger.json",
			expectedSchemaFilename: "new-schema.yaml",
		},
		{
			name:                   "atomics",
			openAPIFilename:        "atomic-types.json",
			expectedSchemaFilename: "atomic-types.yaml",
		},
		{
			name:                   "defaults",
			openAPIFilename:        "defaults.json",
			expectedSchemaFilename: "defaults.yaml",
		},
		{
			name:                   "preserve-unknown",
			openAPIFilename:        "preserve-unknown.json",
			expectedSchemaFilename: "preserve-unknown.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			openAPIPath := filepath.Join("testdata", tc.openAPIFilename)
			expectedNewSchemaPath := filepath.Join("testdata", tc.expectedSchemaFilename)
			testToSchema(t, openAPIPath, expectedNewSchemaPath)
		})
	}
}

func testToSchema(t *testing.T, openAPIPath, expectedNewSchemaPath string) {
	fakeSchema := prototesting.Fake{Path: openAPIPath}
	s, err := fakeSchema.OpenAPISchema()
	if err != nil {
		t.Fatalf("failed to get schema for %s: %v", openAPIPath, err)
	}
	models, err := proto.NewOpenAPIData(s)
	if err != nil {
		t.Fatal(err)
	}

	ns, err := ToSchema(models)
	if err != nil {
		t.Fatal(err)
	}
	got, err := yaml.Marshal(ns)
	if err != nil {
		t.Fatal(err)
	}

	expect, err := ioutil.ReadFile(expectedNewSchemaPath)
	if err != nil {
		t.Fatalf("Unable to read golden data file %q: %v", expectedNewSchemaPath, err)
	}
	expectWithoutHeader := removeHeader(expect)

	fmt.Printf("E: %#v\n\n", string(expectWithoutHeader))
	fmt.Printf("G: %#v\n\n", string(got))
	if string(expectWithoutHeader) != string(got) {
		t.Errorf("Computed schema did not match %q.", expectedNewSchemaPath)
		t.Logf("To recompute this file, run:\n\tgo run ./cmd/openapi2smd/openapi2smd.go < %q > %q",
			filepath.Join("pkg", "schemaconv", openAPIPath),
			filepath.Join("pkg", "schemaconv", expectedNewSchemaPath),
		)
		t.Log("You can then use `git diff` to see the changes.")
	}
}

// removeHeader removes comments (lines starting with "#") from YAML content
func removeHeader(yamlContent []byte) []byte {
	lines := bytes.Split(yamlContent, []byte("\n"))
	var filteredLines [][]byte
	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if !bytes.HasPrefix(trimmed, []byte("#")) {
			filteredLines = append(filteredLines, line)
		}
	}
	return bytes.Join(filteredLines, []byte("\n"))
}
