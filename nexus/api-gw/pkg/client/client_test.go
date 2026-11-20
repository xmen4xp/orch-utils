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
	testOrgType       = "orgs.org.test.io"
	testProjectType   = "projects.project.test.io"
	testSpaceType     = "spaces.space.test.io"
	testAISliceType   = "aislices.aislice.test.io"
	testWorkloadType  = "workloads.workload.test.io"
	testAppType       = "apps.app.test.io"
	testConfigType    = "configs.config.test.io"
	testAlertRuleType = "alertrules.alertrule.test.io"
)

var testConfigGVR = schema.GroupVersionResource{
	Group:    "config.test.io",
	Version:  "v1",
	Resource: "configs",
}

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

func TestDeleteObjectAcceptsDeferredDeleteNodeWithNonDeferredChildren(t *testing.T) {
	// A deferred-delete node's own delete is accepted once its universal pre-check passes --
	// here, aislices exist but aren't themselves deferred-delete, so the check finds nothing
	// blocking. Cascade into its own children is still skipped regardless; the finalizer (not
	// exercised by this fake client) is what actually holds it in Terminating.
	info := recursiveNodeInfo()
	info.DeferredDelete = true
	fakeClient := setupDeleteTest(t, map[string]string{
		testOrgType:          "org-a",
		testProjectType:      "project-a",
		"nexus/display_name": "demo-space",
	}, info)

	fakeClient.PrependReactor("list", "aislices", func(k8stesting.Action) (bool, runtime.Object, error) {
		list := &unstructured.UnstructuredList{Items: []unstructured.Unstructured{{Object: map[string]interface{}{
			"apiVersion": "aislice.test.io/v1",
			"kind":       "AISlice",
			"metadata":   map[string]interface{}{"name": "slice-1"},
		}}}}
		return true, list, nil
	})
	cascadeCalls := 0
	fakeClient.PrependReactor("delete-collection", "*", func(k8stesting.Action) (bool, runtime.Object, error) {
		cascadeCalls++
		return true, nil, nil
	})
	deleteCalled := false
	fakeClient.PrependReactor("delete", "spaces", func(k8stesting.Action) (bool, runtime.Object, error) {
		deleteCalled = true
		return true, nil, nil
	})

	err := DeleteObject(testSpaceGVR, testSpaceType, model.CrdTypeToNodeInfo[testSpaceType], "space-hash")
	require.NoError(t, err)
	require.Zero(t, cascadeCalls, "deferred-delete node must not cascade to children")
	require.True(t, deleteCalled, "deferred-delete node's own delete must be accepted once its pre-check passes")
}

func TestDeleteObjectRejectsDeferredDeleteNodeWithNestedDeferredDeleteDescendant(t *testing.T) {
	// space is deferred-delete, but has its own nested deferred-delete descendant (aislice,
	// several levels down via workload). The universal pre-check now applies regardless of
	// space's own DeferredDelete status: space's delete must be refused outright too, exactly
	// like a non-deferred ancestor would be -- a deferred-delete node is only ever accepted once
	// its own subtree is clear of other deferred-delete instances.
	originalClient := Client
	originalNodeInfo := model.CrdTypeToNodeInfo
	t.Cleanup(func() {
		Client = originalClient
		model.CrdTypeToNodeInfo = originalNodeInfo
	})

	model.CrdTypeToNodeInfo = map[string]model.NodeInfo{
		testSpaceType: {
			DeferredDelete:  true,
			ParentHierarchy: []string{testOrgType, testProjectType},
			Children:        map[string]model.NodeHelperChild{testWorkloadType: {}},
		},
		testWorkloadType: {
			ParentHierarchy: []string{testOrgType, testProjectType, testSpaceType},
			Children:        map[string]model.NodeHelperChild{testAISliceType: {}},
		},
		testAISliceType: {
			DeferredDelete:  true,
			ParentHierarchy: []string{testOrgType, testProjectType, testSpaceType, testWorkloadType},
		},
	}

	space := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": testSpaceGVR.GroupVersion().String(),
		"kind":       "Space",
		"metadata": map[string]interface{}{
			"name": "space-hash",
			"labels": map[string]interface{}{
				testOrgType:          "org-a",
				testProjectType:      "project-a",
				"nexus/display_name": "demo-space",
			},
		},
	}}
	workload := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "workload.test.io/v1",
		"kind":       "Workload",
		"metadata": map[string]interface{}{
			"name": "workload-hash",
			"labels": map[string]interface{}{
				testOrgType:          "org-a",
				testProjectType:      "project-a",
				testSpaceType:        "demo-space",
				"nexus/display_name": "demo-workload",
			},
		},
	}}
	aislice := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "aislice.test.io/v1",
		"kind":       "AISlice",
		"metadata": map[string]interface{}{
			"name": "aislice-hash",
			"labels": map[string]interface{}{
				testOrgType:          "org-a",
				testProjectType:      "project-a",
				testSpaceType:        "demo-space",
				testWorkloadType:     "demo-workload",
				"nexus/display_name": "demo-aislice",
			},
		},
	}}
	fakeClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), space, workload, aislice)
	Client = fakeClient

	spaceDeleteCalled := false
	fakeClient.PrependReactor("delete", "spaces", func(k8stesting.Action) (bool, runtime.Object, error) {
		spaceDeleteCalled = true
		return true, nil, nil
	})

	err := DeleteObject(testSpaceGVR, testSpaceType, model.CrdTypeToNodeInfo[testSpaceType], "space-hash")
	require.Error(t, err)
	require.ErrorContains(t, err, "exists")
	require.False(t, spaceDeleteCalled, "a deferred-delete node must not be deleted while its own subtree still has a deferred-delete descendant")
}

func TestDeleteChildrenSkipsCascadeIntoDeferredDeleteDescendant(t *testing.T) {
	// space (non-deferred) has an aislice descendant that is deferred-delete. Cascading from the
	// ancestor must not recurse into aislice's own children, and must issue DeleteCollection
	// directly on aislices (which the finalizer turns into Terminating rather than removal).
	originalClient := Client
	originalNodeInfo := model.CrdTypeToNodeInfo
	t.Cleanup(func() {
		Client = originalClient
		model.CrdTypeToNodeInfo = originalNodeInfo
	})

	model.CrdTypeToNodeInfo = map[string]model.NodeInfo{
		testSpaceType: {
			ParentHierarchy: []string{testOrgType, testProjectType},
			Children:        map[string]model.NodeHelperChild{testAISliceType: {}},
		},
		testAISliceType: {
			DeferredDelete:  true,
			ParentHierarchy: []string{testOrgType, testProjectType, testSpaceType},
			Children:        map[string]model.NodeHelperChild{testWorkloadType: {}},
		},
		testWorkloadType: {},
	}

	space := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": testSpaceGVR.GroupVersion().String(),
		"kind":       "Space",
		"metadata": map[string]interface{}{
			"name": "space-hash",
			"labels": map[string]interface{}{
				testOrgType:          "org-a",
				testProjectType:      "project-a",
				"nexus/display_name": "demo-space",
			},
		},
	}}
	fakeClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), space)
	Client = fakeClient

	var deletedCollections []string
	fakeClient.PrependReactor("delete-collection", "*", func(action k8stesting.Action) (bool, runtime.Object, error) {
		deletedCollections = append(deletedCollections, action.GetResource().Resource)
		return true, nil, nil
	})

	err := DeleteObject(testSpaceGVR, testSpaceType, model.CrdTypeToNodeInfo[testSpaceType], "space-hash")
	require.NoError(t, err)
	// aislices is cascaded into directly (DeleteCollection, not removed); workloads under it is
	// never cascaded into.
	require.Equal(t, []string{"aislices"}, deletedCollections)
}

func TestDeleteObjectRefusesAncestorWhenDeferredDeleteDescendantExists(t *testing.T) {
	// space (non-deferred) has an aislice descendant that is deferred-delete. Deleting space
	// must be refused outright because aislice exists at all -- regardless of whether aislice
	// itself currently has any workload of its own.
	originalClient := Client
	originalNodeInfo := model.CrdTypeToNodeInfo
	t.Cleanup(func() {
		Client = originalClient
		model.CrdTypeToNodeInfo = originalNodeInfo
	})

	model.CrdTypeToNodeInfo = map[string]model.NodeInfo{
		testSpaceType: {
			ParentHierarchy: []string{testOrgType, testProjectType},
			Children:        map[string]model.NodeHelperChild{testAISliceType: {}},
		},
		testAISliceType: {
			DeferredDelete:  true,
			ParentHierarchy: []string{testOrgType, testProjectType, testSpaceType},
		},
	}

	space := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": testSpaceGVR.GroupVersion().String(),
		"kind":       "Space",
		"metadata": map[string]interface{}{
			"name": "space-hash",
			"labels": map[string]interface{}{
				testOrgType:          "org-a",
				testProjectType:      "project-a",
				"nexus/display_name": "demo-space",
			},
		},
	}}
	aislice := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "aislice.test.io/v1",
		"kind":       "AISlice",
		"metadata": map[string]interface{}{
			"name": "aislice-hash",
			"labels": map[string]interface{}{
				testOrgType:          "org-a",
				testProjectType:      "project-a",
				testSpaceType:        "demo-space",
				"nexus/display_name": "demo-aislice",
			},
		},
	}}
	fakeClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), space, aislice)
	Client = fakeClient

	spaceDeleteCalled := false
	fakeClient.PrependReactor("delete", "spaces", func(k8stesting.Action) (bool, runtime.Object, error) {
		spaceDeleteCalled = true
		return true, nil, nil
	})
	aisliceTouched := false
	fakeClient.PrependReactor("delete-collection", "aislices", func(k8stesting.Action) (bool, runtime.Object, error) {
		aisliceTouched = true
		return true, nil, nil
	})

	err := DeleteObject(testSpaceGVR, testSpaceType, model.CrdTypeToNodeInfo[testSpaceType], "space-hash")
	require.Error(t, err)
	require.ErrorContains(t, err, "exists")
	require.False(t, aisliceTouched, "aislice must not be touched: the pre-scan is read-only")
	require.False(t, spaceDeleteCalled, "space must not be deleted while a deferred-delete descendant exists")
}

func TestDeleteObjectIsAtomicAcrossSiblingBranches(t *testing.T) {
	// config (non-deferred) has two independent children: space (non-deferred, has a live
	// deferred-delete aislice descendant -- blocking) and alertrule (non-deferred, no children
	// -- otherwise deletable). The upfront subtree pre-check must find the blocking aislice
	// before cascading into either branch, so alertrule is never touched regardless of which
	// branch would be processed first.
	originalClient := Client
	originalNodeInfo := model.CrdTypeToNodeInfo
	t.Cleanup(func() {
		Client = originalClient
		model.CrdTypeToNodeInfo = originalNodeInfo
	})

	model.CrdTypeToNodeInfo = map[string]model.NodeInfo{
		testConfigType: {
			ParentHierarchy: []string{testOrgType, testProjectType},
			Children: map[string]model.NodeHelperChild{
				testSpaceType:     {},
				testAlertRuleType: {},
			},
		},
		testSpaceType: {
			ParentHierarchy: []string{testOrgType, testProjectType, testConfigType},
			Children:        map[string]model.NodeHelperChild{testAISliceType: {}},
		},
		testAISliceType: {
			DeferredDelete:  true,
			ParentHierarchy: []string{testOrgType, testProjectType, testConfigType, testSpaceType},
		},
		testAlertRuleType: {ParentHierarchy: []string{testOrgType, testProjectType, testConfigType}},
	}

	config := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": testConfigGVR.GroupVersion().String(),
		"kind":       "Config",
		"metadata": map[string]interface{}{
			"name": "config-hash",
			"labels": map[string]interface{}{
				testOrgType:          "org-a",
				testProjectType:      "project-a",
				"nexus/display_name": "demo-config",
			},
		},
	}}
	space := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "space.test.io/v1",
		"kind":       "Space",
		"metadata": map[string]interface{}{
			"name": "space-hash",
			"labels": map[string]interface{}{
				testOrgType:          "org-a",
				testProjectType:      "project-a",
				testConfigType:       "demo-config",
				"nexus/display_name": "demo-space",
			},
		},
	}}
	aislice := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "aislice.test.io/v1",
		"kind":       "AISlice",
		"metadata": map[string]interface{}{
			"name": "aislice-hash",
			"labels": map[string]interface{}{
				testOrgType:          "org-a",
				testProjectType:      "project-a",
				testConfigType:       "demo-config",
				testSpaceType:        "demo-space",
				"nexus/display_name": "demo-aislice",
			},
		},
	}}
	alertRule := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "alertrule.test.io/v1",
		"kind":       "AlertRule",
		"metadata": map[string]interface{}{
			"name": "alertrule-hash",
			"labels": map[string]interface{}{
				testOrgType:          "org-a",
				testProjectType:      "project-a",
				testConfigType:       "demo-config",
				"nexus/display_name": "demo-alertrule",
			},
		},
	}}

	fakeClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), config, space, aislice, alertRule)
	Client = fakeClient

	alertRuleDeleteCollectionCalled := false
	fakeClient.PrependReactor("delete-collection", "alertrules", func(k8stesting.Action) (bool, runtime.Object, error) {
		alertRuleDeleteCollectionCalled = true
		return true, nil, nil
	})
	spaceTouched := false
	fakeClient.PrependReactor("delete-collection", "spaces", func(k8stesting.Action) (bool, runtime.Object, error) {
		spaceTouched = true
		return true, nil, nil
	})
	configDeleteCalled := false
	fakeClient.PrependReactor("delete", "configs", func(k8stesting.Action) (bool, runtime.Object, error) {
		configDeleteCalled = true
		return true, nil, nil
	})

	err := DeleteObject(testConfigGVR, testConfigType, model.CrdTypeToNodeInfo[testConfigType], "config-hash")
	require.Error(t, err)
	require.ErrorContains(t, err, "exists")
	require.False(t, alertRuleDeleteCollectionCalled, "alertrule must not be touched: the whole delete is refused upfront")
	require.False(t, spaceTouched, "space must not be touched either: the pre-check is read-only")
	require.False(t, configDeleteCalled, "config must not be deleted while a descendant branch is blocked")
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
