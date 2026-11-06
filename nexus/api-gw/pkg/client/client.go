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

// DeleteObject deletes hashedName of kind crdType. When force is true, any guard normally
// imposed by DeferredDelete/BlockDeleteOnChildren (on this object or anywhere in its subtree)
// is bypassed entirely and the object is cascade-deleted the way a non-guarded node always is --
// this is the explicit, caller-opted-in escape hatch for cases where bottom-up deletion isn't
// practical (e.g. cleaning up a failed/abandoned hierarchy).
func DeleteObject(gvr schema.GroupVersionResource, crdType string, crdInfo model.NodeInfo, hashedName string, force bool) error {
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

		if !force && (crdInfo.DeferredDelete || crdInfo.BlockDeleteOnChildren) {
			// Guarded nodes are never cascaded: refuse to delete this object while it still
			// has any child, so the caller must delete bottom-up explicitly (or pass force).
			// This is a synchronous, framework-owned check (no admission webhook involved) --
			// the fixed rule "any child blocks delete", applied the same way regardless of
			// which write path (api-gw, generated client) issues the request.
			if blocked, err := hasChildren(crdType, listOpts); err != nil {
				return err
			} else if blocked {
				return fmt.Errorf("cannot delete %s %q: it still has children", crdType, hashedName)
			}
		} else {
			if !force {
				// All-or-nothing pre-check: refuse the whole delete upfront if any guarded
				// descendant anywhere in the subtree currently has children, before touching
				// any sibling branch. Without this, whichever branch the cascade happens to
				// reach first (Go map iteration order) determines whether an unrelated,
				// already-safe-to-delete branch completes before a blocked branch aborts the
				// call -- this makes the whole delete atomic instead.
				if blocker, err := blockingDeferredDescendant(crdType, listOpts); err != nil {
					return err
				} else if blocker != "" {
					return fmt.Errorf("cannot delete %s %q: descendant %q still has children", crdType, hashedName, blocker)
				}
			}
			for k := range crdInfo.Children {
				if err = DeleteChildren(k, listOpts, force); err != nil {
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

// hasChildren returns true if any instance of any of crdType's declared immediate child
// kinds currently exists under listOpts (the same selector cascade itself uses to scope a
// child List/DeleteCollection to one specific parent instance).
func hasChildren(crdType string, listOpts metav1.ListOptions) (bool, error) {
	crdInfo := model.CrdTypeToNodeInfo[crdType]
	for childType := range crdInfo.Children {
		gvr := gvrForCrdType(childType)
		list, err := Client.Resource(gvr).List(context.TODO(), listOpts)
		if err != nil {
			return false, err
		}
		if len(list.Items) > 0 {
			return true, nil
		}
	}
	return false, nil
}

// blockingDeferredDescendant recursively walks the descendant kinds of crdType (scoped by
// listOpts) and returns "<kind>/<name>" for the first deferred-delete descendant instance found
// that currently has children of its own, or "" if none. It's a read-only, all-or-nothing
// pre-check -- run once, before any cascade begins -- so a blocking descendant anywhere in the
// subtree is found before any sibling branch is touched.
func blockingDeferredDescendant(crdType string, listOpts metav1.ListOptions) (string, error) {
	crdInfo := model.CrdTypeToNodeInfo[crdType]
	for childType := range crdInfo.Children {
		childInfo := model.CrdTypeToNodeInfo[childType]
		list, err := Client.Resource(gvrForCrdType(childType)).List(context.TODO(), listOpts)
		if err != nil {
			return "", err
		}
		for _, item := range list.Items {
			itemListOpts, err := cascadeListOptions(item.GetLabels(), childType, childInfo.ParentHierarchy)
			if err != nil {
				return "", err
			}
			if childInfo.DeferredDelete || childInfo.BlockDeleteOnChildren {
				if blocked, err := hasChildren(childType, itemListOpts); err != nil {
					return "", err
				} else if blocked {
					return childType + "/" + item.GetName(), nil
				}
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

func DeleteChildren(crdType string, listOpts metav1.ListOptions, force bool) error {
	crdInfo := model.CrdTypeToNodeInfo[crdType]
	gvr := gvrForCrdType(crdType)

	if !force && (crdInfo.DeferredDelete || crdInfo.BlockDeleteOnChildren) {
		// Guarded kinds are never cascaded into, whether reached directly or via an ancestor's
		// cascade: list the live instances under this selector and refuse to delete any of
		// them while they still have children of their own. This is the same synchronous,
		// framework-owned check as DeleteObject's, applied here so an ancestor's cascade
		// cannot bypass it by reaching this kind through DeleteCollection instead of a direct
		// delete.
		list, err := Client.Resource(gvr).List(context.TODO(), listOpts)
		if err != nil {
			return err
		}
		for _, item := range list.Items {
			childListOpts, err := cascadeListOptions(item.GetLabels(), crdType, crdInfo.ParentHierarchy)
			if err != nil {
				return err
			}
			if blocked, err := hasChildren(crdType, childListOpts); err != nil {
				return err
			} else if blocked {
				return fmt.Errorf("cannot delete %s %q: it still has children", crdType, item.GetName())
			}
		}
		return Client.Resource(gvr).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, listOpts)
	}

	for k := range crdInfo.Children {
		if err := DeleteChildren(k, listOpts, force); err != nil {
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
