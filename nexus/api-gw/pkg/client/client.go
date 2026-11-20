// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"nexus-api-gw/pkg/model"

	"github.com/vmware-tanzu/graph-framework-for-microservices/common-library/pkg/nexus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	nexus_client "nexus/admin/api/build/nexus-client"

	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
)

var (
	Client              dynamic.Interface
	CoreClient          kubernetes.Interface
	Host                string
	HostScheme          = "http"
	HostTLSClientConfig rest.TLSClientConfig
	appName             = "nexus-api-gw-client"
	log                 = logging.GetLogger(appName)
	NexusClient         *nexus_client.Clientset
)

func New(config *rest.Config) (err error) {
	if config != nil {
		HostTLSClientConfig = config.TLSClientConfig
	}

	if kubeAPIHostPort, isSpecified := os.LookupEnv("KUBEAPI_ENDPOINT"); isSpecified {
		parsedURI, err := url.Parse(kubeAPIHostPort)
		if err != nil {
			return fmt.Errorf("parsing URI %v failed with error %w", kubeAPIHostPort, err)
		}

		if parsedURI.Scheme != "" {
			HostScheme = parsedURI.Scheme
		}
		Host = fmt.Sprintf("%s://%s", HostScheme, parsedURI.Host)

		log.Debug().Msgf("kubeApiHostPort: %+v", kubeAPIHostPort)
		log.Debug().Msgf("parsedURI: %+v", parsedURI)
		log.Debug().Msgf("Host: %s", Host)
	} else {
		Host = config.Host
	}

	config.Burst = 1500
	config.QPS = 1000
	Client, err = dynamic.NewForConfig(config)
	if err != nil {
		log.InfraErr(err).Msg("error building dynamic client")
		return err
	}
	CoreClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	return nil
}

func NewNexusClient(config *rest.Config) error {
	// Create a datamodel client handle.
	var err error
	NexusClient, err = nexus_client.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create nexus client: %s", err)
	}
	return nil
}

func CreateObject(gvr schema.GroupVersionResource, kind, hashedName string, labels map[string]string,
	annotations map[string]string, body map[string]interface{}, finalizers []string,
) error {
	labelsUnstructured := map[string]interface{}{}
	for k, v := range labels {
		labelsUnstructured[k] = v
	}
	finalizersUnstructured := make([]interface{}, len(finalizers))
	for i, f := range finalizers {
		finalizersUnstructured[i] = f
	}

	metadata := map[string]interface{}{
		"name":       hashedName,
		"labels":     labelsUnstructured,
		"finalizers": finalizersUnstructured,
	}
	if len(annotations) > 0 {
		annotationsUnstructured := map[string]interface{}{}
		for k, v := range annotations {
			annotationsUnstructured[k] = v
		}
		metadata["annotations"] = annotationsUnstructured
	}

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": gvr.GroupVersion().String(),
			"kind":       kind,
			"metadata":   metadata,
			"spec":       body,
		},
	}

	// Create resource
	_, err := Client.Resource(gvr).Create(context.TODO(), obj, metav1.CreateOptions{})
	return err
}

func GetObject(gvr schema.GroupVersionResource, hashedName string, opts metav1.GetOptions) (*unstructured.Unstructured, error) {
	obj, err := Client.Resource(gvr).Get(context.TODO(), hashedName, opts)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// DeleteObject deletes hashedName of kind crdType.
func DeleteObject(gvr schema.GroupVersionResource, crdType string, crdInfo model.NodeInfo, hashedName string) error {
	// Get object
	obj, err := Client.Resource(gvr).Get(context.TODO(), hashedName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if len(crdInfo.Children) > 0 {
		listOpts, err := cascadeListOptions(obj.GetLabels(), crdType, crdInfo.ParentHierarchy)
		if err != nil {
			return err
		}

		// All-or-nothing pre-check: refuse the whole delete upfront if a live instance of any
		// deferred-delete kind exists anywhere in the subtree, before touching any sibling
		// branch -- regardless of whether this node itself is deferred-delete. A deferred-delete
		// node is therefore only ever accepted once its own subtree is clear too: the caller
		// must delete that instance explicitly first (always accepted, once it clears its own
		// subtree check) and retry. Without this upfront check, whichever branch cascade happens
		// to reach first (Go map iteration order) determines whether an unrelated,
		// already-safe-to-delete branch completes before the blocker is found -- this makes the
		// whole delete atomic instead.
		if blocker, err := blockingDeferredDescendant(crdType, listOpts); err != nil {
			return err
		} else if blocker != "" {
			return fmt.Errorf("cannot be deleted as %s exists", blocker)
		}

		if crdInfo.DeferredDelete {
			// Deferred-delete nodes are never cascaded: once the check above passes, this
			// delete is accepted, the finalizer holds the object in Terminating, and it's
			// entirely the owning controller's job to tear down children and remove the
			// finalizer (see FinalizeDelete<Node>ByName in the generated client).
		} else {
			for k := range crdInfo.Children {
				if err = DeleteChildren(k, listOpts); err != nil {
					return err
				}
			}
		}
	}

	// Delete object
	if err = Client.Resource(gvr).Delete(context.TODO(), hashedName, metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

// blockingDeferredDescendant recursively walks the descendant kinds of crdType (scoped by
// listOpts) and returns "<Kind> <display name>" for the first live instance found of any
// deferred-delete kind anywhere below crdType, or "" if none exists. The kind here is the
// object's own Kind (e.g. "Space"), not the internal crdType/group-qualified type string or
// hierarchy path, so a caller-facing error message can name what's blocking the delete simply,
// without exposing internal routing details the caller has no reason to know. It's a read-only,
// all-or-nothing pre-check -- run once, before any cascade begins -- so a blocking descendant
// anywhere in the subtree is found before any sibling branch is touched. Presence alone blocks,
// regardless of whether that instance itself currently has children: deleting a deferred-delete
// node is meant to be a deliberate, explicit action, not an incidental side effect of deleting an
// ancestor.
func blockingDeferredDescendant(crdType string, listOpts metav1.ListOptions) (string, error) {
	crdInfo := model.CrdTypeToNodeInfo[crdType]
	for childType := range crdInfo.Children {
		childInfo := model.CrdTypeToNodeInfo[childType]
		list, err := Client.Resource(gvrForCrdType(childType)).List(context.TODO(), listOpts)
		if err != nil {
			return "", err
		}
		if childInfo.DeferredDelete && len(list.Items) > 0 {
			return describeBlocker(list.Items[0]), nil
		}
		for _, item := range list.Items {
			itemListOpts, err := cascadeListOptions(item.GetLabels(), childType, childInfo.ParentHierarchy)
			if err != nil {
				return "", err
			}
			if blocker, err := blockingDeferredDescendant(childType, itemListOpts); err != nil {
				return "", err
			} else if blocker != "" {
				return blocker, nil
			}
		}
	}
	return "", nil
}

// describeBlocker returns "<Kind> <display name>" for obj, e.g. "Space spaceA" -- a short,
// caller-facing identifier naming what's blocking a delete without hierarchy/routing details.
// Falls back to obj's (hashed) object name if the display name label is somehow missing.
func describeBlocker(obj unstructured.Unstructured) string {
	return obj.GetKind() + " " + displayName(obj)
}

// displayName returns obj's human-readable nexus display name, falling back to its (hashed)
// object name if the label is somehow missing.
func displayName(obj unstructured.Unstructured) string {
	if name := obj.GetLabels()["nexus/display_name"]; name != "" {
		return name
	}
	return obj.GetName()
}

func gvrForCrdType(crdType string) schema.GroupVersionResource {
	parts := strings.Split(crdType, ".")
	return schema.GroupVersionResource{
		Group:    strings.Join(parts[1:], "."),
		Version:  "v1",
		Resource: parts[0],
	}
}

func cascadeListOptions(objectLabels map[string]string, crdType string, parentHierarchy []string) (metav1.ListOptions, error) {
	selectorLabels := k8slabels.Set{}
	for _, parent := range parentHierarchy {
		value := objectLabels[parent]
		if value == "" {
			return metav1.ListOptions{}, fmt.Errorf("cannot safely delete children: missing hierarchy label %q", parent)
		}
		selectorLabels[parent] = value
	}

	displayName := objectLabels["nexus/display_name"]
	if displayName == "" {
		return metav1.ListOptions{}, fmt.Errorf("cannot safely delete children: missing hierarchy label %q", "nexus/display_name")
	}
	selectorLabels[crdType] = displayName

	return metav1.ListOptions{LabelSelector: k8slabels.SelectorFromSet(selectorLabels).String()}, nil
}

func DeleteChildren(crdType string, listOpts metav1.ListOptions) error {
	crdInfo := model.CrdTypeToNodeInfo[crdType]
	gvr := gvrForCrdType(crdType)

	if crdInfo.DeferredDelete {
		// Deferred-delete kinds are never cascaded into: the DeleteCollection below still sets
		// deletionTimestamp on each instance, which the finalizer turns into Terminating
		// rather than actually removing it -- the owning controller handles the rest of the
		// teardown. In practice, an ancestor's own upfront pre-scan (blockingDeferredDescendant) already
		// refuses the whole delete before cascade ever reaches a kind with a live instance
		// here, so this branch mainly guards direct children-only cascade calls.
		return Client.Resource(gvr).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, listOpts)
	}

	for k := range crdInfo.Children {
		if err := DeleteChildren(k, listOpts); err != nil {
			return err
		}
	}

	return Client.Resource(gvr).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, listOpts)
}

type PatchOp struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

type Patch []PatchOp

func (p Patch) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func GetParent(parentCrdType string, parentCrdInfo model.NodeInfo, labels map[string]string) (*unstructured.Unstructured, error) {
	parentParts := strings.Split(parentCrdType, ".")
	gvr := schema.GroupVersionResource{
		Group:    strings.Join(parentParts[1:], "."),
		Version:  "v1",
		Resource: parentParts[0],
	}

	parentName := labels[parentCrdType]
	hashedParentName := nexus.GetHashedName(parentCrdType, parentCrdInfo.ParentHierarchy, labels, parentName)
	return GetObject(gvr, hashedParentName, metav1.GetOptions{})
}

// TODO: build PatchOP in common-library.
func UpdateParentWithAddedChild(parentCrdType string, parentCrdInfo model.NodeInfo,
	labels map[string]string, childCrdInfo model.NodeInfo, childCrdType, childName, childHashedName string,
) error {
	var (
		patchType types.PatchType
		marshaled []byte
	)

	parentParts := strings.Split(parentCrdType, ".")
	gvr := schema.GroupVersionResource{
		Group:    strings.Join(parentParts[1:], "."),
		Version:  "v1",
		Resource: parentParts[0],
	}

	parentName := labels[parentCrdType]
	hashedParentName := nexus.GetHashedName(parentCrdType, parentCrdInfo.ParentHierarchy, labels, parentName)
	childGvk := parentCrdInfo.Children[childCrdType]

	childParts := strings.Split(childCrdType, ".")
	group := strings.Join(childParts[1:], ".")
	childNameParts := strings.Split(childCrdInfo.Name, ".")

	if childGvk.IsNamed {
		payload := fmt.Sprintf(`{
			"spec": {
				"%s": {
					"%s": {
						"name": "%s",
						"kind": "%s",
						"group": "%s"
					}
				}
			}
		}`, childGvk.FieldNameGvk, childName, childHashedName, childNameParts[1], group)

		patchType = types.MergePatchType
		marshaled = []byte(payload)
	} else {
		var patch Patch
		patchOp := PatchOp{
			Op:   "replace",
			Path: "/spec/" + childGvk.FieldNameGvk,
			Value: map[string]interface{}{
				"name":  childHashedName,
				"group": group,
				"kind":  childNameParts[1],
			},
		}
		patch = append(patch, patchOp)
		patchBytes, err := patch.Marshal()
		if err != nil {
			return err
		}
		marshaled = patchBytes
		patchType = types.JSONPatchType
	}

	_, err := Client.Resource(gvr).Patch(context.TODO(),
		hashedParentName, patchType, marshaled, metav1.PatchOptions{},
	)
	if err != nil {
		log.InfraErr(err).Msgf("UpdateParentWithAddedChild: failed to patch %v", hashedParentName)
		return err
	}

	return nil
}

func UpdateParentWithRemovedChild(parentCrdType string, parentCrdInfo model.NodeInfo,
	labels map[string]string, childCrdType, childName string,
) error {
	parentParts := strings.Split(parentCrdType, ".")
	gvr := schema.GroupVersionResource{
		Group:    strings.Join(parentParts[1:], "."),
		Version:  "v1",
		Resource: parentParts[0],
	}

	parentName := labels[parentCrdType]
	hashedParentName := nexus.GetHashedName(parentCrdType, parentCrdInfo.ParentHierarchy, labels, parentName)
	childGvk := parentCrdInfo.Children[childCrdType]

	var patchOp PatchOp
	if childGvk.IsNamed {
		patchOp = PatchOp{
			Op:   "remove",
			Path: "/spec/" + childGvk.FieldNameGvk + "/" + childName,
		}
	} else {
		patchOp = PatchOp{
			Op:   "remove",
			Path: "/spec/" + childGvk.FieldNameGvk,
		}
	}

	var patch Patch
	patch = append(patch, patchOp)

	marshaled, err := patch.Marshal()
	if err != nil {
		return err
	}

	_, err = Client.Resource(gvr).Patch(context.TODO(), hashedParentName, types.JSONPatchType, marshaled, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil
}
