// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package mermaid

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/pkg/parser"
)

func TestGenerateMermaidGraph(t *testing.T) {
	// Create test nodes
	rootNode := parser.Node{
		Name:    "Root",
		CrdName: "roots.example.com",
		SingleChildren: map[string]parser.Node{
			"Config": {
				Name:    "Config",
				CrdName: "configs.example.com",
				MultipleChildren: map[string]parser.Node{
					"Services": {
						Name:    "Service",
						CrdName: "services.example.com",
					},
				},
				SingleLink: map[string]parser.Node{
					"Policy": {
						Name:    "Policy",
						CrdName: "policies.example.com",
					},
				},
			},
		},
	}

	// Create graph
	graph := map[string]parser.Node{
		"Root": rootNode,
	}

	// Generate mermaid
	result := GenerateMermaidGraph(graph)

	// Verify basic structure
	if !strings.Contains(result, "graph LR") {
		t.Error("Expected mermaid graph to start with 'graph LR'")
	}

	// Verify nodes are declared
	if !strings.Contains(result, `roots_example_com["roots.example.com"]`) {
		t.Error("Expected root node declaration")
	}

	if !strings.Contains(result, `configs_example_com["configs.example.com"]`) {
		t.Error("Expected config node declaration")
	}

	// Verify relationships
	if !strings.Contains(result, `roots_example_com -->|"HAS CHILD Config"| configs_example_com`) {
		t.Error("Expected child relationship")
	}

	if !strings.Contains(result, `configs_example_com -->|"HAS CHILDREN Services"| services_example_com`) {
		t.Error("Expected children relationship")
	}

	if !strings.Contains(result, `configs_example_com -.->|"LINKS TO Policy"| policies_example_com`) {
		t.Error("Expected link relationship")
	}
}

func TestSanitizeNodeId(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"roots.example.com", "roots_example_com"},
		{"my-service.test.com", "my_service_test_com"},
		{"service with spaces", "service_with_spaces"},
	}

	for _, test := range tests {
		result := sanitizeNodeId(test.input)
		if result != test.expected {
			t.Errorf("sanitizeNodeId(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestGenerateMermaidFile(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "mermaid_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create simple test graph
	graph := map[string]parser.Node{
		"Root": {
			Name:    "Root",
			CrdName: "roots.test.com",
		},
	}

	// Generate file
	err = GenerateMermaidFile(graph, tempDir)
	if err != nil {
		t.Fatalf("GenerateMermaidFile failed: %v", err)
	}

	// Verify file exists
	filePath := filepath.Join(tempDir, "datamodel-graph.md")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Expected mermaid file to be created")
	}

	// Verify file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "# Datamodel Graph") {
		t.Error("Expected file to contain title")
	}

	if !strings.Contains(contentStr, "```mermaid") {
		t.Error("Expected file to contain mermaid code block")
	}

	if !strings.Contains(contentStr, "graph LR") {
		t.Error("Expected file to contain mermaid graph")
	}
}
