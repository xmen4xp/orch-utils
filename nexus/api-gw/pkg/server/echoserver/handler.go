// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package echoserver

import (
	"context"
	"encoding/json"
	baseErrors "errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"nexus-api-gw/pkg/client"
	"nexus-api-gw/pkg/config"
	"nexus-api-gw/pkg/model"
	"nexus-api-gw/pkg/openapi/declarative"
	"nexus-api-gw/pkg/utils"

	"github.com/labstack/echo/v4"
	"github.com/vmware-tanzu/graph-framework-for-microservices/common-library/pkg/nexus"
	nn "github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

const (
	minPathLength           = 2
	rbacResourceHeader      = "x-rbac-resource"
	rbacScopeHeader         = "x-rbac-scope"
	rbacLastUpdatedByHeader = "x-rbac-last-updated-by"
	lastUpdatedByAnnotation = "hdai/last-updated-by"
)

type DefaultResponse struct {
	Message string `json:"message"`
}

// checkAndProxyExtensionAPI checks if the request is for an Extension REST API
// and proxies it to the backend if so. Returns (true, error) if handled as extension,
// (false, nil) if not an extension API.
func (s *EchoServer) checkAndProxyExtensionAPI(nc *NexusContext, crdName string, crdInfo model.NodeInfo, name string) (bool, error) {
	spec, ok := model.GetExtensionRestAPISpec(nc.NexusURI, nc.Request().Method)
	if !ok {
		return false, nil
	}

	labels := parseLabels(nc, crdInfo.ParentHierarchy)
	labels[crdName] = name
	return true, s.proxyExtensionRequest(nc, spec, labels)
}

// GetHandler is used to process GET requests.
func (s *EchoServer) GetHandler(c echo.Context) error {
	nc, ok := c.(*NexusContext)
	if !ok {
		return fmt.Errorf("context is not of type *NexusContext")
	}

	crdName, crdInfo, name, err := getCRDInfoAndName(nc)
	if err != nil {
		log.Error().Msgf("Failed to fetch CRD Info and Name failed with error: %s", err.Error())
		return nc.JSON(http.StatusBadRequest, DefaultResponse{Message: err.Error()})
	}

	// Check if this is an Extension REST API and proxy if so
	if handled, err := s.checkAndProxyExtensionAPI(nc, crdName, crdInfo, name); handled {
		return err
	}

	log.Debug().Msg("CRD Info And Name fetched successfully")
	hashedName, gvr := getHashedNameAndGVR(crdName, crdInfo, name, nc)
	obj, err := client.Client.Resource(gvr).Get(context.TODO(), hashedName, metav1.GetOptions{})
	if err != nil {
		log.Error().Msgf("Failed to create gvr resources error: %s, gvr:%s, crdname:%s, crdInfoPH:%s,hashedName:%s",
			err.Error(), gvr, crdName, crdInfo.ParentHierarchy, hashedName)
		return handleClientError(nc, err)
	}
	log.Debug().Msgf("Processing returned request.. with Obj: %#v", obj)
	return processGetRequest(nc, crdInfo, obj)
}

func getCRDInfoAndName(nc *NexusContext) (string, model.NodeInfo, string, error) {
	crdName := model.URIToCRDType[nc.NexusURI]
	crdInfo := model.CrdTypeToNodeInfo[crdName]
	name := getNameFromRequest(nc, crdInfo)
	if name == "" {
		log.Error().Msgf("could not find required param: %s", crdInfo.Name)
		return "", model.NodeInfo{}, "", fmt.Errorf("could not find required param: %s", crdInfo.Name)
	}
	return crdName, crdInfo, name, nil
}

func getNameFromRequest(nc *NexusContext, crdInfo model.NodeInfo) string {
	name := nexus.DEFAULT_KEY
	// Priority 1: URI params
	for _, param := range nc.ParamNames() {
		if param == crdInfo.Name {
			name = nc.Param(param)
			if name == "" {
				log.Debug().Msgf("Could not find required param %s for request %s", crdInfo.Name, nc.Request().RequestURI)
				return ""
			}
			return name
		}
	}

	// Priority 2: Headers (check alias first, then primary name)
	if headerVal := GetHeaderValue(nc.Request(), crdInfo.Name); headerVal != "" {
		return headerVal
	}

	// Priority 3: Query params
	if nc.QueryParams().Has(crdInfo.Name) {
		name = nc.QueryParams().Get(crdInfo.Name)
	}

	return name
}

// GetHeaderValue looks up a header value by name, checking alias first if configured.
// Priority: alias header > primary header name
func GetHeaderValue(req *http.Request, name string) string {
	if config.Cfg != nil && config.Cfg.HeaderAliases != nil {
		if alias, ok := config.Cfg.HeaderAliases[name]; ok {
			if val := req.Header.Get(alias); val != "" {
				return val
			}
		}
	}
	return req.Header.Get(name)
}

func getHashedNameAndGVR(crdName string, crdInfo model.NodeInfo,
	name string, nc *NexusContext,
) (string, schema.GroupVersionResource) {
	labels := parseLabels(nc, crdInfo.ParentHierarchy)
	hashedName := nexus.GetHashedName(crdName, crdInfo.ParentHierarchy, labels, name)

	parts := strings.Split(crdName, ".")
	gvr := schema.GroupVersionResource{
		Group:    strings.Join(parts[1:], "."),
		Version:  "v1",
		Resource: parts[0],
	}

	return hashedName, gvr
}

func processGetRequest(nc *NexusContext, crdInfo model.NodeInfo, obj *unstructured.Unstructured) error {
	status, err := extractStatus(obj)
	if err != nil {
		return nc.JSON(http.StatusInternalServerError, DefaultResponse{Message: err.Error()})
	}

	uriInfo, ok := model.GetURIInfo(nc.NexusURI)
	if ok {
		switch uriInfo.TypeOfURI {
		case model.SingleLinkURI, model.NamedLinkURI:
			return getLinkInfo(nc, uriInfo.TypeOfURI, crdInfo, obj)
		case model.StatusURI:
			return nc.JSON(http.StatusOK, status)
		default:
			log.Info().Msg("continue : to process DefaultURI ")
		}
	}

	spec, err := extractSpec(obj, crdInfo)
	if err != nil {
		return nc.JSON(http.StatusInternalServerError, DefaultResponse{Message: err.Error()})
	}

	r := map[string]interface{}{
		"spec":   spec,
		"status": status,
	}
	log.Debug().Msgf("processGetRequest with return %#v", r)

	return nc.JSON(http.StatusOK, r)
}

// getLinkInfo returns the children/links of parent node based on the requested gvk.
func getLinkInfo(nc *NexusContext, uriType model.URIType, crdInfo model.NodeInfo, obj *unstructured.Unstructured) error {
	splittedURI := strings.Split(nc.NexusURI, "/")
	if len(splittedURI) < minPathLength {
		log.Error().Msgf("Couldn't determine child object NexusURI %s", nc.NexusURI)
		return nc.JSON(http.StatusBadRequest, DefaultResponse{Message: "Couldn't determine child object"})
	}

	linkFieldName := splittedURI[len(splittedURI)-1]
	gvkField, err := determineGVKField(crdInfo, linkFieldName)
	if err != nil {
		log.Error().Msgf("Failed to determine gvkField err:%s", err.Error())
		return nc.JSON(http.StatusInternalServerError, DefaultResponse{Message: err.Error()})
	}

	spec, ok := obj.Object["spec"].(map[string]interface{})
	if !ok {
		log.Error().Msgf("Failed to parse spec of object, err: %s", err.Error())
		return nc.JSON(http.StatusInternalServerError, DefaultResponse{Message: "Failed to parse spec of object"})
	}

	log.Debug().Msgf("URI %s, splitted URI %s, childFieldName %s, gvkField %s, spec %s, spec[gvkField] %s\n", nc.NexusURI,
		splittedURI, linkFieldName, gvkField, spec, spec[gvkField])

	switch uriType {
	case model.SingleLinkURI:
		return handleSingleLinkURI(nc, gvkField, spec)
	case model.NamedLinkURI:
		return handleNamedLinkURI(nc, gvkField, spec)
	default:
		log.Warn().Msg("Something went wrong during link processing")
		return nc.JSON(http.StatusInternalServerError, DefaultResponse{Message: "Something went wrong during link processing"})
	}
}

func handleSingleLinkURI(nc *NexusContext, gvkField string, spec map[string]interface{}) error {
	l := &model.LinkGvk{}
	marshaled, err := json.Marshal(spec[gvkField])
	if err != nil {
		log.Error().Msgf("Couldn't marshal gvk of link, err: %s", err.Error())
		return nc.JSON(http.StatusInternalServerError, DefaultResponse{Message: "Couldn't marshal gvk of link"})
	}
	err = json.Unmarshal(marshaled, l)
	if err != nil {
		log.Error().Msgf("Couldn't unmarshal gvk of link: %s", err.Error())
		return nc.JSON(http.StatusInternalServerError, DefaultResponse{Message: "Couldn't unmarshal gvk of link"})
	}
	if l.Group != "" {
		resourceName := utils.GetGroupResourceName(l.Kind)
		item, err := getUnstructuredObject(l.Group, resourceName, l.Name)
		if err != nil {
			log.InfraErr(err).Msgf("Couldn't find object %q", l.Name)
			return nc.JSON(http.StatusNotFound, DefaultResponse{Message: "Couldn't find object"})
		}
		// set parent hierarchy
		crdType := utils.GetCrdType(l.Kind, l.Group)
		if crdNodeInfo, ok := model.GetCRDTypeToNodeInfo(crdType); ok {
			l.Hierarchy = utils.GetParentHierarchy(crdNodeInfo.ParentHierarchy, item.GetLabels())
		}

		// get display name of the object
		if val, ok := item.GetLabels()[utils.DisplayNameLabelConst]; ok {
			l.Name = val
		}
		l.Group += "/v1"
	}
	return nc.JSON(http.StatusOK, l)
}

func handleNamedLinkURI(nc *NexusContext, gvkField string, spec map[string]interface{}) error {
	m := make(map[string]model.LinkGvk)
	marshaled, err := json.Marshal(spec[gvkField])
	if err != nil {
		return nc.JSON(http.StatusInternalServerError, DefaultResponse{Message: "Couldn't marshal gvk of link"})
	}
	err = json.Unmarshal(marshaled, &m)
	if err != nil {
		return nc.JSON(http.StatusInternalServerError, DefaultResponse{Message: "Couldn't unmarshal gvk of link"})
	}
	list := make([]model.LinkGvk, len(m))
	i := 0
	hierarchy := []string{}
	for k, link := range m {
		// set parent hierarchy
		if i == 0 {
			resourceName := utils.GetGroupResourceName(link.Kind)
			item, err := getUnstructuredObject(link.Group, resourceName, link.Name)
			if err != nil {
				log.InfraErr(err).Msgf("Couldn't find object, skipping... %q", link.Name)
				continue
			}
			crdType := utils.GetCrdType(link.Kind, link.Group)
			if crdNodeInfo, ok := model.GetCRDTypeToNodeInfo(crdType); ok {
				hierarchy = utils.GetParentHierarchy(crdNodeInfo.ParentHierarchy, item.GetLabels())
			}
		}
		link.Hierarchy = hierarchy
		link.Name = k
		link.Group += "/v1"
		list[i] = link
		i++
	}
	return nc.JSON(http.StatusOK, list)
}

func determineGVKField(crdInfo model.NodeInfo, linkFieldName string) (string, error) {
	for _, child := range crdInfo.Children {
		if child.FieldName == linkFieldName {
			return child.FieldNameGvk, nil
		}
	}

	for _, link := range crdInfo.Links {
		if link.FieldName == linkFieldName {
			return link.FieldNameGvk, nil
		}
	}

	return "", fmt.Errorf("couldn't determine gvk of link")
}

func extractStatus(obj *unstructured.Unstructured) (map[string]interface{}, error) {
	status := make(map[string]interface{})
	if _, ok := obj.Object["status"]; ok {
		status, ok = obj.Object["status"].(map[string]interface{})
		if !ok {
			return nil, baseErrors.New("status is not of type map")
		}
	}
	delete(status, "nexus")
	return status, nil
}

func extractSpec(obj *unstructured.Unstructured, crdInfo model.NodeInfo) (map[string]interface{}, error) {
	spec := make(map[string]interface{})
	if _, ok := obj.Object["spec"]; ok {
		spec, ok = obj.Object["spec"].(map[string]interface{})
		if !ok {
			return nil, baseErrors.New("spec is not of type map")
		}
	}
	for _, v := range crdInfo.Children {
		delete(spec, v.FieldNameGvk)
	}
	for _, v := range crdInfo.Links {
		delete(spec, v.FieldNameGvk)
	}
	return spec, nil
}

// parseRBACScope parses the x-rbac-scope header value (JSON array of display names).
// Returns (allowedNames, true) when filtering should be applied, or (nil, false) when the full list should be returned.
func parseRBACScope(header string) (map[string]bool, bool) {
	if header == "" || header == "*" {
		return nil, false
	}
	var names []string
	if err := json.Unmarshal([]byte(header), &names); err != nil {
		log.Warn().Msgf("Failed to parse %s header as JSON array, skipping filtering: %s", rbacScopeHeader, err.Error())
		return nil, false
	}
	allowed := make(map[string]bool, len(names))
	for _, name := range names {
		allowed[name] = true
	}
	return allowed, true
}

// ListHandler is used to process GET list requests.
func (s *EchoServer) ListHandler(c echo.Context) error {
	nc, ok := c.(*NexusContext)
	if !ok {
		return fmt.Errorf("context is not of type *NexusContext")
	}

	crdName, crdInfo, labels := getCRDInfoAndLabels(nc)
	gvr := constructGVR(crdName)
	opts := constructListOptions(c, labels)

	var allowedNames map[string]bool
	rbacResource := GetHeaderValue(nc.Request(), rbacResourceHeader)
	rbacScope := GetHeaderValue(nc.Request(), rbacScopeHeader)

	if rbacResource != "" && rbacResource != crdInfo.Name {
		log.Warn().Msgf("x-rbac-resource mismatch: got %q, expected %q — skipping RBAC filtering", rbacResource, crdInfo.Name)
	} else {
		allowedNames, _ = parseRBACScope(rbacScope)
	}

	objs, err := client.Client.Resource(gvr).List(context.TODO(), opts)
	if err != nil {
		log.Error().Msgf("Failed to list gvr object err: %s", err.Error())
		return handleClientError(nc, err)
	}
	return nc.JSON(http.StatusOK, processListResponse(objs, crdInfo, allowedNames))
}

func getCRDInfoAndLabels(nc *NexusContext) (string, model.NodeInfo, k8sLabels.Set) {
	crdName := model.URIToCRDType[nc.NexusURI]
	crdInfo := model.CrdTypeToNodeInfo[crdName]
	labels := make(k8sLabels.Set)
	for k, v := range parseLabels(nc, crdInfo.ParentHierarchy) {
		labels[k] = v
	}
	return crdName, crdInfo, labels
}

func constructGVR(crdName string) schema.GroupVersionResource {
	parts := strings.Split(crdName, ".")
	gvr := schema.GroupVersionResource{
		Group:    strings.Join(parts[1:], "."),
		Version:  "v1",
		Resource: parts[0],
	}
	return gvr
}

func constructListOptions(c echo.Context, labels k8sLabels.Set) metav1.ListOptions {
	opts := metav1.ListOptions{
		LabelSelector: labels.AsSelector().String(),
	}

	if c.QueryParams().Has("limit") {
		i, err := strconv.ParseInt(c.QueryParams().Get("limit"), 10, 64)
		if err == nil {
			opts.Limit = i
		}
	} else {
		opts.Limit = 500
	}

	if c.QueryParams().Has("continue") {
		opts.Continue = c.QueryParams().Get("continue")
	}

	return opts
}

func processListResponse(objs *unstructured.UnstructuredList, crdInfo model.NodeInfo, allowedNames map[string]bool) []map[string]interface{} {
	resps := make([]map[string]interface{}, 0)
	for _, item := range objs.Items {
		itemName := item.GetName()
		if val, ok := item.GetLabels()[utils.DisplayNameLabelConst]; ok {
			itemName = val
		}
		if allowedNames != nil && !allowedNames[itemName] {
			continue
		}
		status, err := extractStatus(&item)
		if err != nil {
			continue
		}
		spec, err := extractSpec(&item, crdInfo)
		if err != nil {
			continue
		}

		r := map[string]interface{}{
			"name":   itemName,
			"spec":   spec,
			"status": status,
		}
		resps = append(resps, r)
	}
	return resps
}

func containsProjectUID(projects []string, status map[string]interface{}) bool {
	var projUID string
	if projStatus, ok := status["projectStatus"].(map[string]interface{}); ok {
		if uid, ok := projStatus["uID"]; ok {
			projUID = fmt.Sprintf("%v", uid)
		}
	}
	for _, proj := range projects {
		if proj == projUID {
			return true
		}
	}
	return false
}

// PutHandler is used to process PUT requests.
func (s *EchoServer) PutHandler(c echo.Context) error {
	nc, ok := c.(*NexusContext)
	if !ok {
		log.Error().Msg("context is not of type *NexusContext")
		return fmt.Errorf("context is not of type *NexusContext")
	}

	crdName, crdInfo, name, err := getCRDInfoAndName(nc)
	if err != nil {
		return nc.JSON(http.StatusBadRequest, DefaultResponse{Message: err.Error()})
	}

	// Check if this is an Extension REST API and proxy if so
	if handled, err := s.checkAndProxyExtensionAPI(nc, crdName, crdInfo, name); handled {
		return err
	}

	body, err := parseRequestBody(nc)
	if err != nil {
		return nc.JSON(http.StatusBadRequest, DefaultResponse{Message: err.Error()})
	}

	lastUpdatedBy := GetHeaderValue(nc.Request(), rbacLastUpdatedByHeader)

	hashedName, gvr := getHashedNameAndGVR(crdName, crdInfo, name, nc)
	obj, err := client.Client.Resource(gvr).Get(context.TODO(), hashedName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return handleCreateObject(nc, gvr, crdInfo, hashedName, body, name, lastUpdatedBy)
		}
		return handleClientError(nc, err)
	}

	return updateResource(nc, gvr, obj, body, crdInfo, lastUpdatedBy)
}

func parseRequestBody(nc *NexusContext) (map[string]interface{}, error) {
	body := make(map[string]interface{})
	if err := (&echo.DefaultBinder{}).BindBody(nc, &body); err != nil {
		return nil, err
	}
	return body, nil
}

func isParentFound(nc *NexusContext, crdInfo model.NodeInfo, labels map[string]string,
	name string,
) error {
	for _, parent := range crdInfo.ParentHierarchy {
		// Get crdInfo from crdName.
		crdInfo := model.CrdTypeToNodeInfo[parent]
		hashedName, gvr := getHashedNameAndGVR(parent, crdInfo, labels[parent], nc)
		_, err := client.Client.Resource(gvr).Get(context.TODO(), hashedName, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				errMsg := errors.NewNotFound(schema.GroupResource{Group: gvr.Group, Resource: gvr.Resource},
					fmt.Sprintf("parent %s/%s not found", parent, hashedName))
				log.Error().Msgf("Parent of object %s not found: %v", name, errMsg)
				return errMsg
			}
			return fmt.Errorf("failed to verify the parent %s/%s with error: %w", parent, hashedName, err)
		}
	}
	return nil
}

func handleCreateObject(nc *NexusContext, gvr schema.GroupVersionResource, crdInfo model.NodeInfo,
	hashedName string, body map[string]interface{}, name, lastUpdatedBy string,
) error {
	crdNameParts := strings.Split(crdInfo.Name, ".")
	labels := parseLabels(nc, crdInfo.ParentHierarchy)

	if err := isParentFound(nc, crdInfo, labels, name); err != nil {
		return handleClientError(nc, err)
	}

	labels["nexus/is_name_hashed"] = "true"
	labels["nexus/display_name"] = name
	labels[crdInfo.Name] = name

	var finalizers []string
	if crdInfo.DeferredDelete {
		finalizers = append(finalizers, "nexus.com/nexus-deferred-delete")
	}

	var annotations map[string]string
	if lastUpdatedBy != "" {
		annotations = map[string]string{lastUpdatedByAnnotation: lastUpdatedBy}
	}

	err := client.CreateObject(gvr, crdNameParts[1], hashedName, labels, annotations, body, finalizers)
	if err != nil {
		return handleClientError(nc, err)
	}

	return nc.JSON(http.StatusOK, DefaultResponse{Message: name})
}

// patchHandler is used to modify specific fields.
func (s *EchoServer) PatchHandler(c echo.Context) error {
	nc, ok := c.(*NexusContext)
	if !ok {
		log.Error().Msg("context is not of type *NexusContext.")
		return fmt.Errorf("context is not of type *NexusContext")
	}

	log.Debug().Msg("authenticated and authorized successfully")
	crdName, crdInfo, name, err := getCRDInfoAndName(nc)
	if err != nil {
		log.Error().Msgf("Failed to fetch getCRDInfoAndName, err: %s", err.Error())
		return nc.JSON(http.StatusBadRequest, DefaultResponse{Message: err.Error()})
	}

	// Check if this is an Extension REST API and proxy if so
	if handled, err := s.checkAndProxyExtensionAPI(nc, crdName, crdInfo, name); handled {
		return err
	}

	hashedName, gvr := getHashedNameAndGVR(crdName, crdInfo, name, nc)
	body, err := parseRequestBody(nc)
	if err != nil {
		log.Error().Msgf("Failed to parseRequestBody, err: %s", err.Error())
		return nc.JSON(http.StatusBadRequest, DefaultResponse{Message: err.Error()})
	}

	labels := parseLabels(nc, crdInfo.ParentHierarchy)
	if err := isParentFound(nc, crdInfo, labels, name); err != nil {
		return handleClientError(nc, err)
	}

	lastUpdatedBy := GetHeaderValue(nc.Request(), rbacLastUpdatedByHeader)

	uriInfo, ok := model.GetURIInfo(nc.NexusURI)
	if ok && uriInfo.TypeOfURI == model.StatusURI {
		log.Debug().Msgf("uriInfo.TypeOfURI(%v) == model.StatusURI(%v)", uriInfo.TypeOfURI, model.StatusURI)
		return handleStatusPatch(nc, gvr, hashedName, body, lastUpdatedBy)
	}

	return handleSpecPatch(nc, gvr, hashedName, body, crdInfo, lastUpdatedBy)
}

func handleStatusPatch(nc *NexusContext, gvr schema.GroupVersionResource, hashedName string, body map[string]interface{}, lastUpdatedBy string) error {
	delete(body, "nexus")

	type metadataPayload struct {
		Annotations map[string]string `json:"annotations,omitempty"`
	}
	statusPayload := struct {
		Metadata *metadataPayload       `json:"metadata,omitempty"`
		Status   map[string]interface{} `json:"status"`
	}{
		Status: body,
	}
	if lastUpdatedBy != "" {
		statusPayload.Metadata = &metadataPayload{Annotations: map[string]string{lastUpdatedByAnnotation: lastUpdatedBy}}
	}
	patchBytes, err := json.Marshal(statusPayload)
	if err != nil {
		log.Error().Msgf("error while marshaling status payload: %s", err.Error())
		return nc.JSON(http.StatusBadRequest, DefaultResponse{
			Message: fmt.Sprintf("error while marshaling status payload: %s", err.Error()),
		})
	}

	_, err = client.Client.Resource(gvr).Patch(
		context.TODO(), hashedName, types.MergePatchType, patchBytes, metav1.PatchOptions{}, "status",
	)
	if err != nil {
		log.Error().Msgf("Error in gvr Status patch, err: %s", err.Error())
		return handleClientError(nc, err)
	}
	return nc.JSON(http.StatusOK, DefaultResponse{Message: "Status patch applied successfully"})
}

func handleSpecPatch(nc *NexusContext, gvr schema.GroupVersionResource,
	hashedName string, body map[string]interface{}, crdInfo model.NodeInfo, lastUpdatedBy string,
) error {
	for _, v := range crdInfo.Children {
		delete(body, v.FieldNameGvk)
	}

	for _, v := range crdInfo.Links {
		delete(body, v.FieldNameGvk)
	}

	type metadataPayload struct {
		Annotations map[string]string `json:"annotations,omitempty"`
	}
	payload := struct {
		Metadata *metadataPayload       `json:"metadata,omitempty"`
		Spec     map[string]interface{} `json:"spec"`
	}{
		Spec: body,
	}
	if lastUpdatedBy != "" {
		payload.Metadata = &metadataPayload{Annotations: map[string]string{lastUpdatedByAnnotation: lastUpdatedBy}}
	}

	patchBytes, err := json.Marshal(payload)
	if err != nil {
		log.Error().Msgf("error while marshaling payload: %s", err.Error())
		return nc.JSON(http.StatusBadRequest, DefaultResponse{
			Message: fmt.Sprintf("error while marshaling payload: %s", err.Error()),
		})
	}

	_, err = client.Client.Resource(gvr).Patch(
		context.TODO(), hashedName, types.MergePatchType, patchBytes, metav1.PatchOptions{},
	)
	if err != nil {
		log.Error().Msgf("Error in gvr Spec patch, err: %s", err.Error())
		return handleClientError(nc, err)
	}

	return nc.JSON(http.StatusOK, DefaultResponse{Message: "Patch applied successfully"})
}

// deleteHandler is used to process DELETE requests.
func (s *EchoServer) deleteHandler(c echo.Context) error {
	nc, ok := c.(*NexusContext)
	if !ok {
		return fmt.Errorf("context is not of type *NexusContext")
	}

	crdName, crdInfo, name, err := getCRDInfoAndName(nc)
	if err != nil {
		return nc.JSON(http.StatusBadRequest, DefaultResponse{Message: err.Error()})
	}

	// Check if this is an Extension REST API and proxy if so
	if handled, err := s.checkAndProxyExtensionAPI(nc, crdName, crdInfo, name); handled {
		return err
	}

	hashedName, gvr := getHashedNameAndGVR(crdName, crdInfo, name, nc)
	err = client.DeleteObject(gvr, crdName, crdInfo, hashedName)
	if err != nil {
		log.Error().Msgf("Failed to DeleteObject, err: %s", err.Error())
		return handleClientError(nc, err)
	}
	log.Debug().Msg("authenticated and authorized successfully")
	return nc.NoContent(http.StatusOK)
}

// handleClientError is used to parse client errors and map them to the corresponding statuses from HTTPCodesResponses.
func handleClientError(c echo.Context, err error) error {
	nc, ok := c.(*NexusContext)
	if !ok {
		log.Error().Msg("context is not of type *NexusContext")
		return fmt.Errorf("context is not of type *NexusContext")
	}
	log.Warn().Msg(err.Error())

	switch {
	case errors.IsNotFound(err):
		return respondWithError(nc, http.StatusNotFound, err)
	case errors.IsAlreadyExists(err), errors.IsConflict(err):
		return respondWithError(nc, http.StatusConflict, err)
	case errors.IsInternalError(err):
		return respondWithError(nc, http.StatusInternalServerError, err)
	case errors.IsBadRequest(err):
		return respondWithError(nc, http.StatusBadRequest, err)
	case errors.IsForbidden(err):
		return respondWithError(nc, http.StatusForbidden, err)
	case errors.IsGone(err):
		return respondWithError(nc, http.StatusGone, err)
	case errors.IsInvalid(err):
		return respondWithError(nc, http.StatusUnprocessableEntity, err)
	default:
		log.Error().Msgf("Handle client error. err: %s, status: %d", err.Error(), http.StatusInternalServerError)
		return nc.JSON(http.StatusInternalServerError, DefaultResponse{Message: err.Error()})
	}
}

func respondWithError(nc *NexusContext, statusCode int, errMsg error) error {
	code := nn.ResponseCode(statusCode)
	if val, ok := nc.Codes[code]; ok {
		return nc.JSON(statusCode, DefaultResponse{Message: val.Description})
	}
	return nc.JSON(statusCode, DefaultResponse{Message: errMsg.Error()})
}

func parseLabels(c echo.Context, parents []string) map[string]string {
	nc, ok := c.(*NexusContext)
	if !ok {
		log.InfraError("context is not of type *NexusContext").Msg("")
	}
	labels := make(map[string]string)
	for _, parent := range parents {
		if c, ok := model.CrdTypeToNodeInfo[parent]; ok {
			// Priority 1: URI params
			if v := nc.Param(c.Name); v != "" {
				labels[parent] = v
				// Priority 2: Headers (check alias first, then primary name)
			} else if v := GetHeaderValue(nc.Request(), c.Name); v != "" {
				labels[parent] = v
				// Priority 3: Query params
			} else if nc.QueryParams().Has(c.Name) {
				labels[parent] = nc.QueryParams().Get(c.Name)
			} else {
				labels[parent] = nexus.DEFAULT_KEY
			}
		}
	}

	return labels
}

func getUnstructuredObject(apiGroup, resourceName, name string) (*unstructured.Unstructured, error) {
	gvr := schema.GroupVersionResource{
		Group:    apiGroup,
		Version:  "v1",
		Resource: resourceName,
	}

	item, err := client.Client.Resource(gvr).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Error().Msgf("failed to fetch unstructured obj err: %s", err.Error())
		return nil, err
	}

	return item, nil
}

type PatchOp struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func updateResource(nc *NexusContext, gvr schema.GroupVersionResource,
	obj *unstructured.Unstructured, body map[string]interface{}, crdInfo model.NodeInfo, lastUpdatedBy string,
) error {
	uriInfo, ok := model.GetURIInfo(nc.NexusURI)
	if ok && uriInfo.TypeOfURI == model.StatusURI {
		log.Debug().Msgf("update Status Resource obj: %#v, body: %#v", obj, body)
		return updateStatusResource(nc, gvr, obj, body)
	}

	if nc.QueryParam("update_if_exists") == "false" {
		log.Debug().Msg("Resource already exists")
		return nc.JSON(http.StatusForbidden, DefaultResponse{Message: "Already Exists."})
	}

	spec, ok := obj.Object["spec"].(map[string]interface{})
	if !ok {
		log.Error().Msgf("spec is not of type map, http Status: %d", http.StatusInternalServerError)
		return nc.JSON(http.StatusInternalServerError, DefaultResponse{Message: "spec is not of type map"})
	}
	for _, v := range crdInfo.Children {
		if value, ok := spec[v.FieldNameGvk]; ok {
			body[v.FieldNameGvk] = value
		}
	}
	for _, v := range crdInfo.Links {
		if value, ok := spec[v.FieldNameGvk]; ok {
			body[v.FieldNameGvk] = value
		}
	}
	obj.Object["spec"] = body

	if lastUpdatedBy != "" {
		annots := obj.GetAnnotations()
		if annots == nil {
			annots = make(map[string]string)
		}
		annots[lastUpdatedByAnnotation] = lastUpdatedBy
		obj.SetAnnotations(annots)
	}

	_, err := client.Client.Resource(gvr).Update(context.TODO(), obj, metav1.UpdateOptions{})
	if err != nil {
		log.Error().Msgf("Failed to update the object, err: %s", err.Error())
		return handleClientError(nc, err)
	}
	return nc.JSON(http.StatusOK, DefaultResponse{Message: "Updated successfully"})
}

func updateStatusResource(nc *NexusContext, gvr schema.GroupVersionResource,
	obj *unstructured.Unstructured, body map[string]interface{},
) error {
	if _, ok := body["nexus"]; ok {
		log.Error().Msgf("failed to update status resource, Status: %d, msg: %s", http.StatusBadRequest,
			"can't update nexus status subresource, only user defined status subresource update is allowed")
		return nc.JSON(http.StatusBadRequest, DefaultResponse{
			Message: "can't update nexus status subresource, only user defined status subresource update is allowed",
		})
	}

	if _, ok := obj.Object["status"]; !ok {
		m := []byte("{\"status\":{}}")
		_, err := client.Client.Resource(gvr).Patch(context.TODO(), obj.GetName(),
			types.MergePatchType, m, metav1.PatchOptions{}, "status",
		)
		if err != nil {
			log.Error().Msgf("failed to update status resource object, err: %s", err.Error())
			return handleClientError(nc, err)
		}
	}

	patch := createStatusPatch(body)
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		log.Error().Msgf(
			"Status Marshal failed, Status: %d, error while marshaling json status subresource payload: %s",
			http.StatusBadRequest, err.Error())
		return nc.JSON(http.StatusBadRequest, DefaultResponse{
			Message: fmt.Sprintf("error while marshaling json status subresource payload: %s", err.Error()),
		})
	}

	_, err = client.Client.Resource(gvr).Patch(
		context.TODO(),
		obj.GetName(),
		types.JSONPatchType,
		patchBytes, metav1.PatchOptions{}, "status",
	)
	if err != nil {
		log.Error().Msgf("failed to patch status resource object, err: %s", err.Error())
		return handleClientError(nc, err)
	}

	log.Debug().Msg("Status Updated successfully")
	return nc.JSON(http.StatusOK, DefaultResponse{Message: "Status Updated successfully"})
}

func createStatusPatch(body map[string]interface{}) []PatchOp {
	patch := []PatchOp{}
	for k, v := range body {
		p := PatchOp{
			Op:    "replace",
			Path:  "/status/" + k,
			Value: v,
		}
		patch = append(patch, p)
	}
	return patch
}

func DebugAllHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"isNexusRuntimeEnabled":                   config.Cfg.EnableNexusRuntime,
		"backendService":                          config.Cfg.BackendService,
		"crdTypeToRestUris":                       model.CrdTypeToRestUris,
		"uriToUriInfo":                            model.URIToURIInfo,
		"crdTypeToNodeInfo":                       model.CrdTypeToNodeInfo,
		"datamodelToDatamodelInfo":                model.DatamodelToDatamodelInfo,
		"declarativePaths":                        declarative.ApisList,
		"totalHttpServerRestarts":                 TotalHTTPServerRestartCounter,
		"httpServerRestartsFromOpenApiSpecUpdate": HTTPServerRestartFromOpenAPISpecUpdateCounter,
	})
}
