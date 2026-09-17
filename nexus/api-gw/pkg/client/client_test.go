// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"errors"
	"fmt"
	"testing"

	"nexus-api-gw/pkg/model"

	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	testFooType      = "foos.foo.test.io"
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

func TestDeleteObjectRestrictRejectsWhenChildrenExist(t *testing.T) {
	info := recursiveNodeInfo()
	info.RestrictChildren = []string{testAISliceType}
	fakeClient := setupDeleteTest(t, map[string]string{
		testOrgType:          "org-a",
		testProjectType:      "project-a",
		"nexus/display_name": "demo-space",
	}, info)

	// A child still exists. It carries the hierarchy labels so it matches the
	// cascade selector the guard lists with.
	fakeClient.PrependReactor("list", "*", func(k8stesting.Action) (bool, runtime.Object, error) {
		list := &unstructured.UnstructuredList{}
		list.Items = []unstructured.Unstructured{{Object: map[string]interface{}{
			"apiVersion": "aislice.test.io/v1",
			"kind":       "AISlice",
			"metadata": map[string]interface{}{
				"name": "child-1",
				"labels": stringMapToInterfaceMap(map[string]string{
					testOrgType:     "org-a",
					testProjectType: "project-a",
					testSpaceType:   "demo-space",
				}),
			},
		}}}
		return true, list, nil
	})

	deleteCollectionCalled := false
	fakeClient.PrependReactor("delete-collection", "*", func(k8stesting.Action) (bool, runtime.Object, error) {
		deleteCollectionCalled = true
		return true, nil, nil
	})

	err := DeleteObject(testSpaceGVR, testSpaceType, model.CrdTypeToNodeInfo[testSpaceType], "space-hash")
	require.Error(t, err)
	require.True(t, apierrors.IsConflict(err), "expected a Conflict error, got: %v", err)
	require.False(t, deleteCollectionCalled, "no child should be deleted when restrict rejects the request")
	for _, action := range fakeClient.Actions() {
		require.False(t, action.GetVerb() == "delete" && action.GetResource() == testSpaceGVR,
			"parent must not be deleted when restrict rejects the request")
	}
}

func TestDeleteObjectRestrictAllowsWhenNoChildren(t *testing.T) {
	info := recursiveNodeInfo()
	info.RestrictChildren = []string{testAISliceType}
	fakeClient := setupDeleteTest(t, map[string]string{
		testOrgType:          "org-a",
		testProjectType:      "project-a",
		"nexus/display_name": "demo-space",
	}, info)

	// No children exist.
	fakeClient.PrependReactor("list", "*", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, &unstructured.UnstructuredList{}, nil
	})
	fakeClient.PrependReactor("delete-collection", "*", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, nil
	})

	err := DeleteObject(testSpaceGVR, testSpaceType, model.CrdTypeToNodeInfo[testSpaceType], "space-hash")
	require.NoError(t, err)

	spaceDeleted := false
	for _, action := range fakeClient.Actions() {
		if action.GetVerb() == "delete" && action.GetResource() == testSpaceGVR {
			spaceDeleted = true
		}
	}
	require.True(t, spaceDeleted, "parent should be deleted when restrict finds no children")
}

func TestDeleteObjectRestrictGatesOnlyListedChild(t *testing.T) {
	// Space restricts only on AISlice; Foo is a sibling child that must NOT gate.
	info := recursiveNodeInfo()
	info.Children = map[string]model.NodeHelperChild{
		testAISliceType: {},
		testFooType:     {},
	}
	info.RestrictChildren = []string{testAISliceType}

	t.Run("rejects when the gated child (AISlice) exists", func(t *testing.T) {
		fakeClient := setupDeleteTest(t, spaceLabels(), info)
		fakeClient.PrependReactor("list", "*", listReactor(map[string]int{testAISliceType: 1}))
		deleteCollectionCalled := false
		fakeClient.PrependReactor("delete-collection", "*", func(k8stesting.Action) (bool, runtime.Object, error) {
			deleteCollectionCalled = true
			return true, nil, nil
		})

		err := DeleteObject(testSpaceGVR, testSpaceType, model.CrdTypeToNodeInfo[testSpaceType], "space-hash")
		require.True(t, apierrors.IsConflict(err), "expected Conflict, got: %v", err)
		require.False(t, deleteCollectionCalled, "nothing should be deleted when a gated child exists")
	})

	t.Run("allows when only a non-gated sibling (Foo) exists", func(t *testing.T) {
		fakeClient := setupDeleteTest(t, spaceLabels(), info)
		fakeClient.PrependReactor("list", "*", listReactor(map[string]int{testFooType: 1}))
		fakeClient.PrependReactor("delete-collection", "*", func(k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, nil
		})

		err := DeleteObject(testSpaceGVR, testSpaceType, model.CrdTypeToNodeInfo[testSpaceType], "space-hash")
		require.NoError(t, err)
		spaceDeleted := false
		for _, action := range fakeClient.Actions() {
			if action.GetVerb() == "delete" && action.GetResource() == testSpaceGVR {
				spaceDeleted = true
			}
		}
		require.True(t, spaceDeleted, "a non-gated sibling must not block deletion")
	})
}

func TestDeleteObjectRestrictWithMultipleGatedChildren(t *testing.T) {
	// Space has 4 children; 3 are gated (AISlice, Workload, App), Foo is not.
	info := recursiveNodeInfo()
	info.Children = map[string]model.NodeHelperChild{
		testAISliceType:  {},
		testWorkloadType: {},
		testAppType:      {},
		testFooType:      {},
	}
	info.RestrictChildren = []string{testAISliceType, testWorkloadType, testAppType}

	t.Run("rejects when a later gated child exists (checks all gated, not just the first)", func(t *testing.T) {
		fakeClient := setupDeleteTest(t, spaceLabels(), info)
		// Only the 3rd gated child (App) exists; the first two are empty.
		fakeClient.PrependReactor("list", "*", listReactor(map[string]int{testAppType: 1}))
		deleteCollectionCalled := false
		fakeClient.PrependReactor("delete-collection", "*", func(k8stesting.Action) (bool, runtime.Object, error) {
			deleteCollectionCalled = true
			return true, nil, nil
		})

		err := DeleteObject(testSpaceGVR, testSpaceType, model.CrdTypeToNodeInfo[testSpaceType], "space-hash")
		require.True(t, apierrors.IsConflict(err), "expected Conflict, got: %v", err)
		require.False(t, deleteCollectionCalled, "nothing should be deleted when any gated child exists")
	})

	t.Run("allows when all gated children are absent even if a non-gated sibling exists", func(t *testing.T) {
		fakeClient := setupDeleteTest(t, spaceLabels(), info)
		// Only the non-gated Foo exists.
		fakeClient.PrependReactor("list", "*", listReactor(map[string]int{testFooType: 1}))
		fakeClient.PrependReactor("delete-collection", "*", func(k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, nil
		})

		err := DeleteObject(testSpaceGVR, testSpaceType, model.CrdTypeToNodeInfo[testSpaceType], "space-hash")
		require.NoError(t, err)
		spaceDeleted := false
		for _, action := range fakeClient.Actions() {
			if action.GetVerb() == "delete" && action.GetResource() == testSpaceGVR {
				spaceDeleted = true
			}
		}
		require.True(t, spaceDeleted, "non-gated siblings must not block deletion")
	})
}

func TestDeleteObjectRestrictBlocksWhenChildListFails(t *testing.T) {
	info := recursiveNodeInfo()
	info.RestrictChildren = []string{testAISliceType}
	fakeClient := setupDeleteTest(t, spaceLabels(), info)

	listErr := errors.New("api server unavailable")
	fakeClient.PrependReactor("list", "*", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, listErr
	})
	deleteCollectionCalled := false
	fakeClient.PrependReactor("delete-collection", "*", func(k8stesting.Action) (bool, runtime.Object, error) {
		deleteCollectionCalled = true
		return true, nil, nil
	})

	err := DeleteObject(testSpaceGVR, testSpaceType, model.CrdTypeToNodeInfo[testSpaceType], "space-hash")
	require.Error(t, err)
	require.False(t, apierrors.IsConflict(err), "a List failure must surface as an error, not Conflict")
	require.False(t, deleteCollectionCalled, "nothing should be deleted when the restrict check cannot be evaluated")
	for _, action := range fakeClient.Actions() {
		require.False(t, action.GetVerb() == "delete" && action.GetResource() == testSpaceGVR,
			"parent must not be deleted when the restrict check errors")
	}
}

func TestDeleteObjectRestrictBlocksAncestorDelete(t *testing.T) {
	// Config -> Space -> AISlice, where Space restricts on AISlice. Deleting the
	// Config (an ancestor) must be blocked while an AISlice exists, because the
	// cascade would otherwise remove it.
	const testConfigType = "configs.config.test.io"
	configGVR := gvrFromCrdType(testConfigType)

	originalClient := Client
	originalNodeInfo := model.CrdTypeToNodeInfo
	t.Cleanup(func() {
		Client = originalClient
		model.CrdTypeToNodeInfo = originalNodeInfo
	})

	model.CrdTypeToNodeInfo = map[string]model.NodeInfo{
		testConfigType: {
			ParentHierarchy: []string{testOrgType, testProjectType},
			Children:        map[string]model.NodeHelperChild{testSpaceType: {}},
		},
		testSpaceType: {
			ParentHierarchy:  []string{testOrgType, testProjectType},
			Children:         map[string]model.NodeHelperChild{testAISliceType: {}},
			RestrictChildren: []string{testAISliceType},
		},
		testAISliceType: {},
	}

	configObj := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": configGVR.GroupVersion().String(),
		"kind":       "Config",
		"metadata": map[string]interface{}{
			"name": "config-hash",
			"labels": stringMapToInterfaceMap(map[string]string{
				testOrgType:          "org-a",
				testProjectType:      "project-a",
				"nexus/display_name": "default",
			}),
		},
	}}
	gvrToListKind := map[schema.GroupVersionResource]string{
		configGVR:                       "ConfigList",
		gvrFromCrdType(testSpaceType):   "SpaceList",
		gvrFromCrdType(testAISliceType): "AISliceList",
	}

	t.Run("rejects ancestor delete while a restricted descendant exists", func(t *testing.T) {
		fakeClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), gvrToListKind, configObj)
		Client = fakeClient
		// The AISlice carries the config-scoped hierarchy labels so it matches the
		// selector the guard lists with when deleting the Config.
		fakeClient.PrependReactor("list", "*", func(k8stesting.Action) (bool, runtime.Object, error) {
			list := &unstructured.UnstructuredList{}
			list.Items = []unstructured.Unstructured{{Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "child-1",
					"labels": stringMapToInterfaceMap(map[string]string{
						testOrgType:     "org-a",
						testProjectType: "project-a",
						testConfigType:  "default",
					}),
				},
			}}}
			return true, list, nil
		})
		deleteCollectionCalled := false
		fakeClient.PrependReactor("delete-collection", "*", func(k8stesting.Action) (bool, runtime.Object, error) {
			deleteCollectionCalled = true
			return true, nil, nil
		})

		err := DeleteObject(configGVR, testConfigType, model.CrdTypeToNodeInfo[testConfigType], "config-hash")
		require.True(t, apierrors.IsConflict(err), "expected Conflict, got: %v", err)
		require.False(t, deleteCollectionCalled, "ancestor delete must remove nothing while a restricted descendant exists")
	})

	t.Run("allows ancestor delete when no restricted descendant exists", func(t *testing.T) {
		fakeClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), gvrToListKind, configObj)
		Client = fakeClient
		fakeClient.PrependReactor("list", "*", func(k8stesting.Action) (bool, runtime.Object, error) {
			return true, &unstructured.UnstructuredList{}, nil
		})
		fakeClient.PrependReactor("delete-collection", "*", func(k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, nil
		})

		err := DeleteObject(configGVR, testConfigType, model.CrdTypeToNodeInfo[testConfigType], "config-hash")
		require.NoError(t, err)
	})
}

func spaceLabels() map[string]string {
	return map[string]string{
		testOrgType:          "org-a",
		testProjectType:      "project-a",
		"nexus/display_name": "demo-space",
	}
}

// listReactor returns a fake LIST reaction that yields the requested number of
// items (with matching hierarchy labels) per child crd type.
func listReactor(countsByCrdType map[string]int) k8stesting.ReactionFunc {
	byResource := map[string]int{}
	for crdType, n := range countsByCrdType {
		byResource[gvrFromCrdType(crdType).Resource] = n
	}
	return func(action k8stesting.Action) (bool, runtime.Object, error) {
		list := &unstructured.UnstructuredList{}
		for i := 0; i < byResource[action.GetResource().Resource]; i++ {
			list.Items = append(list.Items, unstructured.Unstructured{Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": fmt.Sprintf("child-%d", i),
					"labels": stringMapToInterfaceMap(map[string]string{
						testOrgType:     "org-a",
						testProjectType: "project-a",
						testSpaceType:   "demo-space",
					}),
				},
			}})
		}
		return true, list, nil
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
		testFooType: {},
	}

	object := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": testSpaceGVR.GroupVersion().String(),
		"kind":       "Space",
		"metadata": map[string]interface{}{
			"name":   "space-hash",
			"labels": stringMapToInterfaceMap(objectLabels),
		},
	}}
	// Register list kinds so the fake client can serve LIST calls (used by the
	// restrict deletion-policy guard). Harmless for tests that only delete.
	gvrToListKind := map[schema.GroupVersionResource]string{
		testSpaceGVR:                     "SpaceList",
		gvrFromCrdType(testAISliceType):  "AISliceList",
		gvrFromCrdType(testWorkloadType): "WorkloadList",
		gvrFromCrdType(testAppType):      "AppList",
		gvrFromCrdType(testFooType):      "FooList",
	}
	fakeClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), gvrToListKind, object)
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
