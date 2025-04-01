// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package tenancy_test

import (
	"context"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	configv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/config.edge-orchestrator.intel.com/v1"
	orgv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/org.edge-orchestrator.intel.com/v1"
	orgactivewatcherv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/orgactivewatcher.edge-orchestrator.intel.com/v1"
	orgwatcherv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/orgwatcher.edge-orchestrator.intel.com/v1"
	projectv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/project.edge-orchestrator.intel.com/v1"
	projectactivewatcherv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/projectactivewatcher.edge-orchestrator.intel.com/v1"
	projectwatcherv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/projectwatcher.edge-orchestrator.intel.com/v1"
	tenancyv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/tenancy.edge-orchestrator.intel.com/v1"
	nexus_client "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/nexus-client"
	"github.com/open-edge-platform/orch-utils/tenancy-manager/pkg/config"
	"github.com/open-edge-platform/orch-utils/tenancy-manager/pkg/tenancy"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

const (
	defaultName = "default"
	finalizer   = "nexus.com/nexus-deferred-delete"

	org1Name = "coke"
	org2Name = "zoo"
	org3Name = "fanta"
	org4Name = "pepsi"

	proj1Name    = "foo"
	proj2Name    = "bar"
	watcher1Name = "cluster-orchestrator"
	watcher2Name = "cluster-orchestrator-2"

	org1HashedName = "0140f5a72a168ee97905bed39b52170aaa9ef554"
	org2HashedName = "c51fa12e956c2b6cbf89f2ee7d42634afb6e34c9"
	org3HashedName = "9811914870bcfacb7fd4c6fb7a60b2f2ae51f7b0"
	org4HashedName = "cb5907e3721297a5044be192c6122ddb0b7745dd"

	org1proj1ConfigHash = "63cb6e0a01fb2c807f174eaf2a5926b74192eed2"
	org3proj1ConfigHash = "974532094b875891b37e5638b2fcd407679ed4e0"
	org2proj2ConfigHash = "40f4a77fa21008d90079767d51b3ff6c63cc7260"
	org4proj1ConfigHash = "e3d672b0bf484fba7f3899eddb7fdc2e8385bfb3"

	pollingInterval = 500 * time.Millisecond
	timeoutInterval = 30 * time.Second
)

var (
	nexusClient       *nexus_client.Clientset
	configClient      *nexus_client.ConfigConfig
	tenancyReconciler *tenancy.Reconciler
)

var _ = ginkgo.BeforeSuite(func() {
	tenancy.Testing = true
	nexusClient = nexus_client.NewFakeClient()
	nexusClient.DynamicClient = fake.NewSimpleDynamicClient(runtime.NewScheme())

	mockObjects()
	createParentNodes()

	conf := &config.Config{
		OrgCreateTimeoutInSecs:     8,
		OrgDeleteTimeoutInSecs:     8,
		ProjectCreateTimeoutInSecs: 8,
		ProjectDeleteTimeoutInSecs: 8,
	}

	tenancyReconciler = tenancy.NewReconciler(nexusClient, conf)
	config := nexusClient.TenancyMultiTenancy().Config()

	// Register all the Org and Project handlers to process the events.
	_, err := config.Orgs("*").RegisterAddCallback(tenancyReconciler.ProcessOrgsAdd)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	_, err = config.Orgs("*").RegisterUpdateCallback(tenancyReconciler.ProcessOrgsUpdate)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	_, err = config.Orgs("*").Folders("*").Projects("*").RegisterAddCallback(tenancyReconciler.ProcessProjectsAdd)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	_, err = config.Orgs("*").Folders("*").Projects("*").RegisterUpdateCallback(tenancyReconciler.ProcessProjectsUpdate)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
})

func TestTenancyAPIs(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Tenancy Suite")
}

// mockObjects mocks org and project objects using dynamic client for status updates.
func mockObjects() {
	orgToUpdate, err := nexusClient.DynamicClient.Resource(constructOrgGVR()).
		Create(context.Background(), constructUnstructuredOrg(org1HashedName), metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(orgToUpdate).NotTo(gomega.BeNil())

	org2ToUpdate, err := nexusClient.DynamicClient.Resource(constructOrgGVR()).
		Create(context.Background(), constructUnstructuredOrg(org2HashedName), metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(org2ToUpdate).NotTo(gomega.BeNil())

	org3ToUpdate, err := nexusClient.DynamicClient.Resource(constructOrgGVR()).
		Create(context.Background(), constructUnstructuredOrg(org3HashedName), metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(org3ToUpdate).NotTo(gomega.BeNil())

	org4ToUpdate, err := nexusClient.DynamicClient.Resource(constructOrgGVR()).
		Create(context.Background(), constructUnstructuredOrg(org4HashedName), metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(org4ToUpdate).NotTo(gomega.BeNil())

	projectToUpdate, err := nexusClient.DynamicClient.Resource(constructProjectGVR()).
		Create(context.Background(), constructUnstructuredProject(org1proj1ConfigHash), metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(projectToUpdate).NotTo(gomega.BeNil())

	project2ToUpdate, err := nexusClient.DynamicClient.Resource(constructProjectGVR()).
		Create(context.Background(), constructUnstructuredProject(org2proj2ConfigHash), metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(project2ToUpdate).NotTo(gomega.BeNil())

	project3ToUpdate, err := nexusClient.DynamicClient.Resource(constructProjectGVR()).
		Create(context.Background(), constructUnstructuredProject(org3proj1ConfigHash), metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(project3ToUpdate).NotTo(gomega.BeNil())

	project4ToUpdate, err := nexusClient.DynamicClient.Resource(constructProjectGVR()).
		Create(context.Background(), constructUnstructuredProject(org4proj1ConfigHash), metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(project4ToUpdate).NotTo(gomega.BeNil())
}

// createParentNodes creates parent nodes.
func createParentNodes() {
	tenancyClient, err := nexusClient.AddTenancyMultiTenancy(context.Background(), &tenancyv1.MultiTenancy{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultName,
		},
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	configClient, err = tenancyClient.AddConfig(context.Background(), &configv1.Config{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultName,
		},
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

func constructOrgObj(name string) *orgv1.Org {
	return &orgv1.Org{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			ResourceVersion:   "1",
			Finalizers:        []string{finalizer},
			DeletionTimestamp: &metav1.Time{Time: time.Now().UTC()},
		},
	}
}

func constructOrgWatcherObj(name string) *orgwatcherv1.OrgWatcher {
	return &orgwatcherv1.OrgWatcher{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func constructOrgActiveWatcherObj(name string) *orgactivewatcherv1.OrgActiveWatcher {
	return &orgactivewatcherv1.OrgActiveWatcher{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			ResourceVersion: "1",
		},
		Spec: orgactivewatcherv1.OrgActiveWatcherSpec{
			StatusIndicator: orgactivewatcherv1.StatusIndicationIdle,
			Message:         "Active watcher creation is initiated",
			TimeStamp:       12345,
		},
	}
}

func constructProjectObj() *projectv1.Project {
	return &projectv1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name:              proj1Name,
			ResourceVersion:   "1",
			Finalizers:        []string{finalizer},
			DeletionTimestamp: &metav1.Time{Time: time.Now().UTC()},
		},
	}
}

func constructProjectWatcherObj(name string) *projectwatcherv1.ProjectWatcher {
	return &projectwatcherv1.ProjectWatcher{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func constructProjectActiveWatcherObj(name string) *projectactivewatcherv1.ProjectActiveWatcher {
	return &projectactivewatcherv1.ProjectActiveWatcher{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			ResourceVersion: "1",
		},
		Spec: projectactivewatcherv1.ProjectActiveWatcherSpec{
			StatusIndicator: projectactivewatcherv1.StatusIndicationIdle,
			Message:         "Active watcher creation is initiated",
			TimeStamp:       12345,
		},
	}
}

func constructOrgGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "org.edge-orchestrator.intel.com",
		Version:  "v1",
		Resource: "orgs",
	}
}

func constructProjectGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "project.edge-orchestrator.intel.com",
		Version:  "v1",
		Resource: "projects",
	}
}

func constructUnstructuredOrg(hashedName string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "org.edge-orchestrator.intel.com/v1",
			"kind":       "orgs",
			"metadata": map[string]interface{}{
				"name":            hashedName,
				"resourceVersion": "1",
			},
		},
	}
}

func constructUnstructuredProject(hashedName string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "project.edge-orchestrator.intel.com/v1",
			"kind":       "projects",
			"metadata": map[string]interface{}{
				"name":            hashedName,
				"resourceVersion": "1",
			},
		},
	}
}
