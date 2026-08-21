// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"errors"
	"testing"

	"nexus-api-gw/pkg/model"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

const (
	testOrgType      = "orgs.org.test.io"
	testProjectType  = "projects.project.test.io"
	testSpaceType    = "spaces.space.test.io"
	testAISliceType  = "aislices.aislice.test.io"
	testWorkloadType = "workloads.workload.test.io"
	testAppType      = "apps.app.test.io"
)

var testSpaceGVR = schema.GroupVersionResource{
	Group:    "space.test.io",
	Version:  "v1",
	Resource: "spaces",
}

type deleteCollectionCall struct {
	resource string
	selector string
}

func TestDeleteObjectScopesRecursiveDeletionToRootHierarchy(t *testing.T) {
	fakeClient := setupDeleteTest(t, map[string]string{
		testOrgType:                 "org-a",
		testProjectType:             "project-a",
		"nexus/display_name":        "demo-space",
		"app.kubernetes.io/managed": "extra-label",
	}, recursiveNodeInfo())

	var calls []deleteCollectionCall
	fakeClient.PrependReactor("delete-collection", "*", func(action k8stesting.Action) (bool, runtime.Object, error) {
		deleteAction, ok := action.(k8stesting.DeleteCollectionAction)
		require.True(t, ok)
		calls = append(calls, deleteCollectionCall{
			resource: action.GetResource().Resource,
			selector: deleteAction.GetListRestrictions().Labels.String(),
		})
		return true, nil, nil
	})

	err := DeleteObject(testSpaceGVR, testSpaceType, model.CrdTypeToNodeInfo[testSpaceType], "space-hash")
	require.NoError(t, err)
	require.Len(t, calls, 3)
	require.Equal(t, []string{"apps", "workloads", "aislices"}, []string{
		calls[0].resource,
		calls[1].resource,
		calls[2].resource,
	})

	expectedSelector := k8slabels.SelectorFromSet(k8slabels.Set{
		testOrgType:     "org-a",
		testProjectType: "project-a",
		testSpaceType:   "demo-space",
	}).String()
	for _, call := range calls {
		require.Equal(t, expectedSelector, call.selector)
		require.NotContains(t, call.selector, "app.kubernetes.io/managed")
	}

	selector, err := k8slabels.Parse(calls[0].selector)
	require.NoError(t, err)
	require.True(t, selector.Matches(k8slabels.Set{
		testOrgType:     "org-a",
		testProjectType: "project-a",
		testSpaceType:   "demo-space",
	}))
	require.False(t, selector.Matches(k8slabels.Set{
		testOrgType:     "org-b",
		testProjectType: "project-a",
		testSpaceType:   "demo-space",
	}))
	require.False(t, selector.Matches(k8slabels.Set{
		testOrgType:     "org-a",
		testProjectType: "project-b",
		testSpaceType:   "demo-space",
	}))

	for _, action := range fakeClient.Actions() {
		require.NotEqual(t, "list", action.GetVerb())
	}
}

func TestDeleteObjectPreservesDefaultHierarchyValues(t *testing.T) {
	fakeClient := setupDeleteTest(t, map[string]string{
		testOrgType:          "default",
		testProjectType:      "default",
		"nexus/display_name": "demo-space",
	}, recursiveNodeInfo())

	var selector string
	fakeClient.PrependReactor("delete-collection", "*", func(action k8stesting.Action) (bool, runtime.Object, error) {
		deleteAction, ok := action.(k8stesting.DeleteCollectionAction)
		require.True(t, ok)
		selector = deleteAction.GetListRestrictions().Labels.String()
		return true, nil, nil
	})

	err := DeleteObject(testSpaceGVR, testSpaceType, model.CrdTypeToNodeInfo[testSpaceType], "space-hash")
	require.NoError(t, err)
	require.Equal(t, k8slabels.SelectorFromSet(k8slabels.Set{
		testOrgType:     "default",
		testProjectType: "default",
		testSpaceType:   "demo-space",
	}).String(), selector)
}

func TestDeleteObjectFailsBeforeDeletionWhenHierarchyLabelsAreMissing(t *testing.T) {
	tests := map[string]map[string]string{
		"missing organization": {
			testProjectType:      "project-a",
			"nexus/display_name": "demo-space",
		},
		"missing project": {
			testOrgType:          "org-a",
			"nexus/display_name": "demo-space",
		},
		"missing display name": {
			testOrgType:     "org-a",
			testProjectType: "project-a",
		},
	}

	for name, objectLabels := range tests {
		t.Run(name, func(t *testing.T) {
			fakeClient := setupDeleteTest(t, objectLabels, recursiveNodeInfo())

			err := DeleteObject(testSpaceGVR, testSpaceType, model.CrdTypeToNodeInfo[testSpaceType], "space-hash")
			require.ErrorContains(t, err, "cannot safely delete children")
			require.Len(t, fakeClient.Actions(), 1)
			require.Equal(t, "get", fakeClient.Actions()[0].GetVerb())
		})
	}
}

func TestDeleteObjectDeletesLeafWithoutCascadeLabels(t *testing.T) {
	fakeClient := setupDeleteTest(t, nil, model.NodeInfo{})

	err := DeleteObject(testSpaceGVR, testSpaceType, model.CrdTypeToNodeInfo[testSpaceType], "space-hash")
	require.NoError(t, err)
	require.Len(t, fakeClient.Actions(), 2)
	require.Equal(t, "get", fakeClient.Actions()[0].GetVerb())
	require.Equal(t, "delete", fakeClient.Actions()[1].GetVerb())
}

func TestDeleteObjectDoesNotDeleteRootWhenChildDeletionFails(t *testing.T) {
	fakeClient := setupDeleteTest(t, map[string]string{
		testOrgType:          "org-a",
		testProjectType:      "project-a",
		"nexus/display_name": "demo-space",
	}, recursiveNodeInfo())
	deleteErr := errors.New("delete collection failed")
	fakeClient.PrependReactor("delete-collection", "*", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, deleteErr
	})

	err := DeleteObject(testSpaceGVR, testSpaceType, model.CrdTypeToNodeInfo[testSpaceType], "space-hash")
	require.ErrorIs(t, err, deleteErr)
	for _, action := range fakeClient.Actions() {
		require.False(t, action.GetVerb() == "delete" && action.GetResource() == testSpaceGVR)
	}
}

func setupDeleteTest(t *testing.T, objectLabels map[string]string, rootInfo model.NodeInfo) *dynamicfake.FakeDynamicClient {
	t.Helper()

	originalClient := Client
	originalNodeInfo := model.CrdTypeToNodeInfo
	t.Cleanup(func() {
		Client = originalClient
		model.CrdTypeToNodeInfo = originalNodeInfo
	})

	model.CrdTypeToNodeInfo = map[string]model.NodeInfo{
		testSpaceType: rootInfo,
		testAISliceType: {
			Children: map[string]model.NodeHelperChild{
				testWorkloadType: {},
			},
		},
		testWorkloadType: {
			Children: map[string]model.NodeHelperChild{
				testAppType: {},
			},
		},
		testAppType: {},
	}

	object := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": testSpaceGVR.GroupVersion().String(),
		"kind":       "Space",
		"metadata": map[string]interface{}{
			"name":   "space-hash",
			"labels": stringMapToInterfaceMap(objectLabels),
		},
	}}
	fakeClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), object)
	Client = fakeClient
	return fakeClient
}

func recursiveNodeInfo() model.NodeInfo {
	return model.NodeInfo{
		ParentHierarchy: []string{testOrgType, testProjectType},
		Children: map[string]model.NodeHelperChild{
			testAISliceType: {},
		},
	}
}

func stringMapToInterfaceMap(values map[string]string) map[string]interface{} {
	result := make(map[string]interface{}, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}
