// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package mermaid

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/pkg/parser"
)

// GenerateMermaidGraph creates a mermaid graph representation of the datamodel
func GenerateMermaidGraph(graph map[string]parser.Node) string {
	var mermaid strings.Builder
	mermaid.WriteString("graph TB\n")

	// Track visited nodes to avoid duplicates
	visited := make(map[string]bool)

	// Generate node declarations and relationships
	for _, rootNode := range graph {
		rootNode.Walk(func(node *parser.Node) {
			nodeId := sanitizeNodeId(node.CrdName)

			// Declare node if not already visited
			if !visited[nodeId] {
				mermaid.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", nodeId, node.CrdName))
				visited[nodeId] = true
			}

			// Generate child relationships
			for _, child := range node.SingleChildren {
				childId := sanitizeNodeId(child.CrdName)
				if !visited[childId] {
					mermaid.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", childId, child.CrdName))
					visited[childId] = true
				}
				mermaid.WriteString(fmt.Sprintf("  %s -->|\"HAS CHILD\"| %s\n", nodeId, childId))
			}

			for _, child := range node.MultipleChildren {
				childId := sanitizeNodeId(child.CrdName)
				if !visited[childId] {
					mermaid.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", childId, child.CrdName))
					visited[childId] = true
				}
				mermaid.WriteString(fmt.Sprintf("  %s -->|\"HAS CHILDREN\"| %s\n", nodeId, childId))
			}

			// Generate link relationships (using dotted lines to distinguish from children)
			for _, link := range node.SingleLink {
				linkId := sanitizeNodeId(link.CrdName)
				if !visited[linkId] {
					mermaid.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", linkId, link.CrdName))
					visited[linkId] = true
				}
				mermaid.WriteString(fmt.Sprintf("  %s -.->|\"LINKS TO\"| %s\n", nodeId, linkId))
			}

			for _, link := range node.MultipleLink {
				linkId := sanitizeNodeId(link.CrdName)
				if !visited[linkId] {
					mermaid.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", linkId, link.CrdName))
					visited[linkId] = true
				}
				mermaid.WriteString(fmt.Sprintf("  %s -.->|\"LINKS TO\"| %s\n", nodeId, linkId))
			}
		})
	}

	return mermaid.String()
}

// sanitizeNodeId removes characters that might cause issues in mermaid syntax
func sanitizeNodeId(nodeId string) string {
	// Replace dots and other special characters with underscores
	sanitized := strings.ReplaceAll(nodeId, ".", "_")
	sanitized = strings.ReplaceAll(sanitized, "-", "_")
	sanitized = strings.ReplaceAll(sanitized, " ", "_")
	return sanitized
}

// GenerateMermaidFile creates a mermaid graph file and saves it to the specified directory
func GenerateMermaidFile(graph map[string]parser.Node, outputDir string) error {
	// Generate the mermaid graph content
	mermaidContent := GenerateMermaidGraph(graph)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %v", outputDir, err)
	}

	// Write to file
	filePath := filepath.Join(outputDir, "datamodel-graph-mermaid.md")
	content := fmt.Sprintf("# Datamodel Graph\n\nThis graph shows the relationships between nodes in the datamodel.\n\n```mermaid\n%s```\n", mermaidContent)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write mermaid file to %s: %v", filePath, err)
	}

	log.Infof("Generated mermaid graph file: %s", filePath)
	return nil
}
