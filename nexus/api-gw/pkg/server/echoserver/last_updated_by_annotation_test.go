// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package echoserver_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"nexus-api-gw/pkg/client"
	"nexus-api-gw/pkg/model"
	"nexus-api-gw/pkg/server/echoserver"

	"github.com/labstack/echo/v4"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	testLastUpdatedByUser = "test-user@cisco.com"
	hdaiLastUpdatedByKey  = "hdai/last-updated-by"
	rootCRDType           = "roots.root.vmware.org"
	rootNexusURI          = "/v1/roots/{Root.root}"
)

var rootGVR = schema.GroupVersionResource{
	Group:    "root.vmware.org",
	Version:  "v1",
	Resource: "roots",
}

var rootNodeInfo = model.NodeInfo{
	Name:            "Root.root",
	ParentHierarchy: []string{},
	Children:        map[string]model.NodeHelperChild{},
	Links:           map[string]model.NodeHelperChild{},
	IsSingleton:     false,
	DeferredDelete:  false,
}

func setupRootCRD() {
	model.URIToCRDType[rootNexusURI] = rootCRDType
	model.CrdTypeToNodeInfo[rootCRDType] = rootNodeInfo
}

func makePutContext(e *echo.Echo, name, body string, headers map[string]string) (*echoserver.NexusContext, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPut, "/v1/roots/"+name, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("Root.root")
	c.SetParamValues(name)
	nc := &echoserver.NexusContext{
		Context:  c,
		NexusURI: rootNexusURI,
	}
	return nc, rec
}

func makePatchContext(e *echo.Echo, name, body string, headers map[string]string) (*echoserver.NexusContext, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPatch, "/v1/roots/"+name, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("Root.root")
	c.SetParamValues(name)
	nc := &echoserver.NexusContext{
		Context:  c,
		NexusURI: rootNexusURI,
	}
	return nc, rec
}

var _ = ginkgo.Describe("last-updated-by annotation", ginkgo.Ordered, func() {
	var serverObj *echoserver.EchoServer
	var stopCh chan struct{}
	var e *echo.Echo

	ginkgo.BeforeAll(func() {
		serverObj, stopCh = setupServer()
		e = echo.New()
		setupRootCRD()
	})

	ginkgo.AfterAll(func() {
		teardownServer(serverObj, stopCh)
	})

	ginkgo.It("PUT (create) with x-rbac-last-updated-by sets hdai/last-updated-by annotation", func() {
		nc, rec := makePutContext(e, "myroot", `{"foo":"bar"}`, map[string]string{
			"x-rbac-last-updated-by": testLastUpdatedByUser,
		})

		err := serverObj.PutHandler(nc)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))

		list, err := client.Client.Resource(rootGVR).List(context.TODO(), metav1.ListOptions{})
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(list.Items).NotTo(gomega.BeEmpty())

		var created *map[string]interface{}
		for i := range list.Items {
			labels := list.Items[i].GetLabels()
			if labels["nexus/display_name"] == "myroot" {
				obj := list.Items[i].Object
				created = &obj
				break
			}
		}
		gomega.Expect(created).NotTo(gomega.BeNil(), "created object not found in fake client")

		metadata := (*created)["metadata"].(map[string]interface{})
		annotations, ok := metadata["annotations"].(map[string]interface{})
		gomega.Expect(ok).To(gomega.BeTrue(), "annotations not set on created object")
		gomega.Expect(annotations[hdaiLastUpdatedByKey]).To(gomega.Equal(testLastUpdatedByUser))
	})

	ginkgo.It("PUT (create) without x-rbac-last-updated-by does not set annotation", func() {
		nc, rec := makePutContext(e, "myroot2", `{"foo":"baz"}`, nil)

		err := serverObj.PutHandler(nc)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))

		list, err := client.Client.Resource(rootGVR).List(context.TODO(), metav1.ListOptions{})
		gomega.Expect(err).To(gomega.BeNil())

		var created *map[string]interface{}
		for i := range list.Items {
			labels := list.Items[i].GetLabels()
			if labels["nexus/display_name"] == "myroot2" {
				obj := list.Items[i].Object
				created = &obj
				break
			}
		}
		gomega.Expect(created).NotTo(gomega.BeNil(), "created object not found")

		metadata := (*created)["metadata"].(map[string]interface{})
		annotations, _ := metadata["annotations"].(map[string]interface{})
		gomega.Expect(annotations[hdaiLastUpdatedByKey]).To(gomega.BeEmpty())
	})

	ginkgo.It("PUT (update) with x-rbac-last-updated-by sets hdai/last-updated-by annotation on existing object", func() {
		nc, rec := makePutContext(e, "myroot", `{"foo":"updated"}`, map[string]string{
			"x-rbac-last-updated-by": "updater@cisco.com",
		})

		err := serverObj.PutHandler(nc)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))

		list, err := client.Client.Resource(rootGVR).List(context.TODO(), metav1.ListOptions{})
		gomega.Expect(err).To(gomega.BeNil())

		var updated *map[string]interface{}
		for i := range list.Items {
			if list.Items[i].GetLabels()["nexus/display_name"] == "myroot" {
				obj := list.Items[i].Object
				updated = &obj
				break
			}
		}
		gomega.Expect(updated).NotTo(gomega.BeNil())

		metadata := (*updated)["metadata"].(map[string]interface{})
		annotations := metadata["annotations"].(map[string]interface{})
		gomega.Expect(annotations[hdaiLastUpdatedByKey]).To(gomega.Equal("updater@cisco.com"))
	})

	ginkgo.It("PATCH spec with x-rbac-last-updated-by returns 200 with annotation in patch", func() {
		patchBody := `{"foo":"patched"}`
		nc, rec := makePatchContext(e, "myroot", patchBody, map[string]string{
			"x-rbac-last-updated-by": "patcher@cisco.com",
		})

		err := serverObj.PatchHandler(nc)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))

		var resp map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(resp["message"]).To(gomega.Equal("Patch applied successfully"))
	})

	ginkgo.It("PATCH spec without x-rbac-last-updated-by returns 200", func() {
		patchBody := `{"foo":"patched-no-user"}`
		nc, rec := makePatchContext(e, "myroot", patchBody, nil)

		err := serverObj.PatchHandler(nc)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
	})
})
