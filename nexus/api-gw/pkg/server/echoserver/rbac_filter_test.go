// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package echoserver

import (
	"testing"

	"nexus-api-gw/pkg/model"
	"nexus-api-gw/pkg/utils"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestParseRBACScope_MissingHeader(t *testing.T) {
	allowed, shouldFilter := parseRBACScope("")
	assert.Nil(t, allowed)
	assert.False(t, shouldFilter)
}

func TestParseRBACScope_Wildcard(t *testing.T) {
	allowed, shouldFilter := parseRBACScope("*")
	assert.Nil(t, allowed)
	assert.False(t, shouldFilter)
}

func TestParseRBACScope_ValidArray(t *testing.T) {
	allowed, shouldFilter := parseRBACScope(`["org-a","org-b"]`)
	assert.True(t, shouldFilter)
	assert.Equal(t, map[string]bool{"org-a": true, "org-b": true}, allowed)
}

func TestParseRBACScope_SingleItem(t *testing.T) {
	allowed, shouldFilter := parseRBACScope(`["coke"]`)
	assert.True(t, shouldFilter)
	assert.Equal(t, map[string]bool{"coke": true}, allowed)
}

func TestParseRBACScope_EmptyArray(t *testing.T) {
	allowed, shouldFilter := parseRBACScope(`[]`)
	assert.True(t, shouldFilter)
	assert.Empty(t, allowed)
}

func TestParseRBACScope_InvalidJSON(t *testing.T) {
	allowed, shouldFilter := parseRBACScope("not-json")
	assert.Nil(t, allowed)
	assert.False(t, shouldFilter)
}

func makeUnstructuredItem(displayName, k8sName string) unstructured.Unstructured {
	item := unstructured.Unstructured{}
	item.SetName(k8sName)
	if displayName != "" {
		item.SetLabels(map[string]string{
			utils.DisplayNameLabelConst: displayName,
		})
	}
	item.Object["spec"] = map[string]interface{}{}
	item.Object["status"] = map[string]interface{}{}
	return item
}

func TestProcessListResponse_NoFilter(t *testing.T) {
	objs := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			makeUnstructuredItem("coke", "hashed-coke"),
			makeUnstructuredItem("pepsi", "hashed-pepsi"),
		},
	}
	crdInfo := model.NodeInfo{
		Children: map[string]model.NodeHelperChild{},
		Links:    map[string]model.NodeHelperChild{},
	}

	result := processListResponse(objs, crdInfo, nil)

	assert.Len(t, result, 2)
	assert.Equal(t, "coke", result[0]["name"])
	assert.Equal(t, "pepsi", result[1]["name"])
}

func TestProcessListResponse_WithFilter(t *testing.T) {
	objs := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			makeUnstructuredItem("coke", "hashed-coke"),
			makeUnstructuredItem("pepsi", "hashed-pepsi"),
			makeUnstructuredItem("sprite", "hashed-sprite"),
		},
	}
	crdInfo := model.NodeInfo{
		Children: map[string]model.NodeHelperChild{},
		Links:    map[string]model.NodeHelperChild{},
	}
	allowedNames := map[string]bool{"coke": true, "sprite": true}

	result := processListResponse(objs, crdInfo, allowedNames)

	assert.Len(t, result, 2)
	names := []string{result[0]["name"].(string), result[1]["name"].(string)}
	assert.ElementsMatch(t, []string{"coke", "sprite"}, names)
}

func TestProcessListResponse_FilterNoMatch(t *testing.T) {
	objs := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			makeUnstructuredItem("coke", "hashed-coke"),
			makeUnstructuredItem("pepsi", "hashed-pepsi"),
		},
	}
	crdInfo := model.NodeInfo{
		Children: map[string]model.NodeHelperChild{},
		Links:    map[string]model.NodeHelperChild{},
	}
	allowedNames := map[string]bool{"sprite": true}

	result := processListResponse(objs, crdInfo, allowedNames)

	assert.Empty(t, result)
}

func TestProcessListResponse_FilterItemWithNoDisplayName(t *testing.T) {
	objs := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			makeUnstructuredItem("", "hashed-no-display"),
			makeUnstructuredItem("pepsi", "hashed-pepsi"),
		},
	}
	crdInfo := model.NodeInfo{
		Children: map[string]model.NodeHelperChild{},
		Links:    map[string]model.NodeHelperChild{},
	}
	allowedNames := map[string]bool{"pepsi": true}

	result := processListResponse(objs, crdInfo, allowedNames)

	assert.Len(t, result, 1)
	assert.Equal(t, "pepsi", result[0]["name"])
}
