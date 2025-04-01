/*
 * Copyright (C) 2025 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions
 * and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package tenancy_test

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	folderv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/folder.edge-orchestrator.intel.com/v1"
	orgsv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/org.edge-orchestrator.intel.com/v1"
	orgactivewatcherv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/orgactivewatcher.edge-orchestrator.intel.com/v1"
	projectv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/project.edge-orchestrator.intel.com/v1"
	projectactivewatcherv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/projectactivewatcher.edge-orchestrator.intel.com/v1"
	runtimeorgsv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/runtimeorg.edge-orchestrator.intel.com/v1"
	runtimeprojectsv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/runtimeproject.edge-orchestrator.intel.com/v1"
	nexus_client "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/nexus-client"
	"github.com/open-edge-platform/orch-utils/tenancy-manager/pkg/config"
	"github.com/open-edge-platform/orch-utils/tenancy-manager/pkg/tenancy"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = ginkgo.Describe("Org/Project Create and Delete operation", ginkgo.Ordered, func() {
	ginkgo.Context("GetExpectedWatchers", func() {
		ginkgo.When("expectedOrgWatchers is called", func() {
			ginkgo.It("should error if config object is not found", func() {
				nexusClient := nexus_client.NewFakeClient()
				tenancyReconciler := tenancy.NewReconciler(nexusClient, &config.Config{})
				_, err := tenancy.GetExpectedOrgWatchers(tenancyReconciler.Client)
				gomega.Expect(err.Error()).
					To(
						gomega.ContainSubstring(
							"failed to get config object with an error: configs.config.edge-orchestrator.intel.com",
						),
					)
			})
		})
		ginkgo.When("expectedProjectWatchers is called", func() {
			ginkgo.It("should error if config object is not found", func() {
				nexusClient := nexus_client.NewFakeClient()
				tenancyReconciler := tenancy.NewReconciler(nexusClient, &config.Config{})
				_, err := tenancy.GetExpectedProjectWatchers(tenancyReconciler.Client)
				gomega.Expect(err.Error()).
					To(
						gomega.ContainSubstring(
							"failed to get config object with an error: configs.config.edge-orchestrator.intel.com",
						),
					)
			})
		})
	})

	ginkgo.Context("Org/Project Create and Delete without watchers", func() {
		var (
			orgClient     *nexus_client.OrgOrg
			projectClient *nexus_client.ProjectProject
			err           error
		)
		ginkgo.When("config org is created", func() {
			ginkgo.It("should create the corresponding runtime org", func() {
				// Create the config org.
				orgClient, err = configClient.AddOrgs(context.Background(), constructOrgObj(org1Name))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Check if the runtime org is created successfully.
				gomega.Eventually(func() bool {
					result, err := tenancyReconciler.Client.TenancyMultiTenancy().Runtime().
						GetOrgs(context.Background(), org1Name)
					if result != nil {
						return result.Name != "" && err == nil
					}
					return false
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "runtime org to be retrieved successfully")

				// Check if the config org is set to active.
				gomega.Eventually(func() bool {
					org, err := nexusClient.DynamicClient.Resource(constructOrgGVR()).
						Get(context.Background(), org1HashedName, metav1.GetOptions{})
					if err != nil {
						return false
					}
					status, found, err := unstructured.NestedString(org.Object, "status", "orgStatus", "statusIndicator")
					if err != nil || !found {
						return false
					}
					return status == string(orgsv1.StatusIndicationIdle)
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config org to be set to active")
			})

			ginkgo.When("config project is created", func() {
				ginkgo.It("should create the corresponding runtime project", func() {
					// Create the config project.
					projectClient, err = tenancyReconciler.Client.TenancyMultiTenancy().Config().Orgs(org1Name).
						Folders(defaultName).AddProjects(context.Background(), constructProjectObj())
					gomega.Expect(err).NotTo(gomega.HaveOccurred())

					// Check if the runtime project is created successfully.
					gomega.Eventually(func() bool {
						result, err := tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs(org1Name).
							Folders(defaultName).GetProjects(context.Background(), proj1Name)
						if result != nil {
							return result.Name != "" && err == nil
						}
						return false
					}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "runtime project to be retrieved successfully")

					// Check if the config project is set to active.
					gomega.Eventually(func() bool {
						project, err := nexusClient.DynamicClient.Resource(constructProjectGVR()).
							Get(context.Background(), org1proj1ConfigHash, metav1.GetOptions{})
						if err != nil {
							return false
						}
						status, found, err := unstructured.NestedString(project.Object,
							"status", "projectStatus", "statusIndicator")
						if err != nil || !found {
							return false
						}
						return status == string(projectv1.StatusIndicationIdle)
					}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config project to be set to active")
				})
			})

			ginkgo.When("config org/project is deleted", func() {
				ginkgo.It("should delete the corresponding runtime objects", func() {
					// Trigger deletion on project object.
					err := projectClient.Update(context.Background())
					gomega.Expect(err).NotTo(gomega.HaveOccurred())

					// Check if the runtime project is deleted successfully.
					gomega.Eventually(func() bool {
						result, err := tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs(org1Name).
							Folders(defaultName).GetProjects(context.Background(), proj1Name)
						return errors.IsNotFound(err) && result == nil
					}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "runtime project to be deleted successfully")

					// Trigger deletion on org object.
					err = orgClient.Update(context.Background())
					gomega.Expect(err).NotTo(gomega.HaveOccurred())

					// Check if the runtime org is deleted successfully.
					gomega.Eventually(func() bool {
						result, err := tenancyReconciler.Client.TenancyMultiTenancy().Runtime().
							GetOrgs(context.Background(), org1Name)
						return errors.IsNotFound(err) && result == nil
					}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "runtime org to be deleted successfully")
				})
			})
		})
	})

	ginkgo.Context("Org/Project Create and Delete with watchers", func() {
		var (
			orgClient     *nexus_client.OrgOrg
			projectClient *nexus_client.ProjectProject
			err           error
		)
		ginkgo.BeforeAll(func() {
			runtime := nexusClient.TenancyMultiTenancy().Runtime()

			_, err := runtime.Orgs("*").ActiveWatchers("*").RegisterAddCallback(tenancyReconciler.ProcessOrgActiveWatcherAdd)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			_, err = runtime.Orgs("*").ActiveWatchers("*").RegisterUpdateCallback(tenancyReconciler.ProcessOrgActiveWatcherUpdate)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			_, err = runtime.Orgs("*").ActiveWatchers("*").RegisterDeleteCallback(tenancyReconciler.ProcessOrgActiveWatcherDelete)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			_, err = runtime.Orgs("*").Folders("*").Projects("*").ActiveWatchers("*").
				RegisterAddCallback(tenancyReconciler.ProcessProjectActiveWatcherAdd)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			_, err = runtime.Orgs("*").Folders("*").Projects("*").ActiveWatchers("*").
				RegisterUpdateCallback(tenancyReconciler.ProcessProjectActiveWatcherUpdate)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			_, err = runtime.Orgs("*").Folders("*").Projects("*").ActiveWatchers("*").
				RegisterDeleteCallback(tenancyReconciler.ProcessProjectActiveWatcherDelete)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.When("org is created", func() {
			ginkgo.It("should mark config org IDLE when all activewatchers have acknowledged", func() {
				// Create org watcher.
				_, err := configClient.AddOrgWatchers(context.Background(), constructOrgWatcherObj(watcher1Name))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				_, err = configClient.AddOrgWatchers(context.Background(), constructOrgWatcherObj(watcher2Name))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Create the config org.
				orgClient, err = configClient.AddOrgs(context.Background(), constructOrgObj(org3Name))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Check if the runtime org is created successfully.
				gomega.Eventually(func() bool {
					result, err := tenancyReconciler.Client.TenancyMultiTenancy().Runtime().
						GetOrgs(context.Background(), org3Name)
					if result != nil {
						return result.Name != "" && err == nil
					}
					return false
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "runtime org to be retrieved successfully")

				// Check if the config org is set to InProgress.
				// This will remain InProgress until all the active watchers have acknowledged the org.
				gomega.Eventually(func() bool {
					org, err := nexusClient.DynamicClient.Resource(constructOrgGVR()).
						Get(context.Background(), org3HashedName, metav1.GetOptions{})
					if err != nil {
						return false
					}
					status, found, err := unstructured.NestedString(org.Object, "status", "orgStatus", "statusIndicator")
					if err != nil || !found {
						return false
					}
					return status == string(orgsv1.StatusIndicationInProgress)
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config org to be set to InProgress")

				// Create the orgactivewatcher1.
				_, err = tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs(org3Name).
					AddActiveWatchers(context.Background(), constructOrgActiveWatcherObj(watcher1Name))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Create the orgactivewatcher2.
				watcherObj2 := constructOrgActiveWatcherObj(watcher2Name)
				watcherObj2.Spec.StatusIndicator = orgactivewatcherv1.StatusIndicationInProgress
				fmt.Printf("watcherObj2: %#v", watcherObj2)
				activeWatcher, err := tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs(org3Name).
					AddActiveWatchers(context.Background(), watcherObj2)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Check if the config org is set to InProgress.
				// This will remain InProgress until all the active watchers are set to IDLE.
				gomega.Eventually(func() bool {
					org, err := nexusClient.DynamicClient.Resource(constructOrgGVR()).
						Get(context.Background(), org3HashedName, metav1.GetOptions{})
					if err != nil {
						return false
					}
					message, found, err := unstructured.NestedString(org.Object, "status", "orgStatus", "message")
					if err != nil || !found {
						return false
					}
					fmt.Printf("message: %v", message)
					return message == "Waiting for watchers [cluster-orchestrator-2] to acknowledge org fanta"
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config org to be set to 'InProgress'")

				activeWatcher.Spec.StatusIndicator = orgactivewatcherv1.StatusIndicationIdle
				err = activeWatcher.Update(context.Background())
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Check if the config org is set to IDLE.
				gomega.Eventually(func() bool {
					org, err := nexusClient.DynamicClient.Resource(constructOrgGVR()).
						Get(context.Background(), org3HashedName, metav1.GetOptions{})
					if err != nil {
						return false
					}
					status, found, err := unstructured.NestedString(org.Object, "status", "orgStatus", "statusIndicator")
					if err != nil || !found {
						return false
					}
					return status == string(orgsv1.StatusIndicationIdle)
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config org to be set to 'IDLE'")
			})
		})

		ginkgo.When("project is created", func() {
			ginkgo.It("should set config project IDLE when all activewatchers have acknowledged", func() {
				// Create the project watcher.
				_, err := configClient.AddProjectWatchers(context.Background(), constructProjectWatcherObj(watcher1Name))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Create the project watcher.
				_, err = configClient.AddProjectWatchers(context.Background(), constructProjectWatcherObj(watcher2Name))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Create the config project.
				projectClient, err = tenancyReconciler.Client.TenancyMultiTenancy().Config().Orgs(org3Name).Folders(defaultName).
					AddProjects(context.Background(), constructProjectObj())
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Check if the runtime project is created successfully.
				gomega.Eventually(func() bool {
					result, err := tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs(org3Name).Folders(defaultName).
						GetProjects(context.Background(), proj1Name)
					if result != nil {
						return result.Name != "" && err == nil
					}
					return false
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "runtime project to be retrieved successfully")

				// Check if the config project is set to InProgress.
				// This will remain InProgress until all the active watchers have acknowledged the project.
				gomega.Eventually(func() bool {
					project, err := nexusClient.DynamicClient.Resource(constructProjectGVR()).
						Get(context.Background(), org3proj1ConfigHash, metav1.GetOptions{})
					if err != nil {
						return false
					}
					status, found, err := unstructured.NestedString(project.Object, "status", "projectStatus", "statusIndicator")
					if err != nil || !found {
						return false
					}
					return status == string(projectv1.StatusIndicationInProgress)
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config project to be set to InProgress")

				// Create the projectactivewatcher.
				proj1Aw1, err := tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs(org3Name).Folders(defaultName).
					Projects(proj1Name).AddActiveWatchers(context.Background(), constructProjectActiveWatcherObj(watcher1Name))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(proj1Aw1).NotTo(gomega.BeNil())

				// Create the projectactivewatcher.
				proj1Aw2 := constructProjectActiveWatcherObj(watcher2Name)
				proj1Aw2.Spec.StatusIndicator = projectactivewatcherv1.StatusIndicationInProgress
				projectActiveWatcher2, err := tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs(org3Name).
					Folders(defaultName).Projects(proj1Name).AddActiveWatchers(context.Background(), proj1Aw2)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Check if the config project is set to InProgress.
				// This will remain InProgress until all the active watchers are set to IDLE.
				gomega.Eventually(func() bool {
					project, err := nexusClient.DynamicClient.Resource(constructProjectGVR()).
						Get(context.Background(), org3proj1ConfigHash, metav1.GetOptions{})
					if err != nil {
						return false
					}
					message, found, err := unstructured.NestedString(project.Object, "status", "projectStatus", "message")
					if err != nil || !found {
						return false
					}
					return message == "Waiting for watchers [cluster-orchestrator-2] to acknowledge project foo"
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config project to be set to InProgress")

				projectActiveWatcher2.Spec.StatusIndicator = projectactivewatcherv1.StatusIndicationIdle
				err = projectActiveWatcher2.Update(context.Background())
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Check if the config project is set to active.
				gomega.Eventually(func() bool {
					project, err := nexusClient.DynamicClient.Resource(constructProjectGVR()).
						Get(context.Background(), org3proj1ConfigHash, metav1.GetOptions{})
					if err != nil {
						return false
					}
					status, found, err := unstructured.NestedString(project.Object, "status", "projectStatus", "statusIndicator")
					if err != nil || !found {
						return false
					}
					return status == string(projectv1.StatusIndicationIdle)
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config project to be set to IDLE")
			})
		})

		ginkgo.When("project is deleted", func() {
			ginkgo.It("should wait until all the activewatchers are deleted", func() {
				// Wait for create to complete.
				time.Sleep(5 * time.Second)

				// Trigger deletion on project object.
				err = projectClient.Update(context.Background())
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Check if the runtime project delete is in 'InProgress'.
				gomega.Eventually(func() bool {
					project, err := nexusClient.DynamicClient.Resource(constructProjectGVR()).
						Get(context.Background(), org3proj1ConfigHash, metav1.GetOptions{})
					if err != nil {
						return false
					}
					status, found, err := unstructured.NestedString(project.Object, "status", "projectStatus", "statusIndicator")
					if err != nil || !found {
						return false
					}
					return status == string(projectv1.StatusIndicationInProgress)
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "runtime project deletion is InProgress")

				// Delete the projectactivewatcher.
				err = tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs(org3Name).
					Folders(defaultName).Projects(proj1Name).
					DeleteActiveWatchers(context.Background(), watcher1Name)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Check if the runtime project delete is in 'InProgress'.
				gomega.Eventually(func() bool {
					project, err := nexusClient.DynamicClient.Resource(constructProjectGVR()).
						Get(context.Background(), org3proj1ConfigHash, metav1.GetOptions{})
					if err != nil {
						return false
					}
					message, found, err := unstructured.NestedString(project.Object, "status", "projectStatus", "message")
					if err != nil || !found {
						return false
					}
					return message == "Waiting for watchers [cluster-orchestrator-2] to be deleted"
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "runtime project deletion is InProgress")

				// Delete the projectactivewatcher.
				err = tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs(org3Name).Folders(defaultName).
					Projects(proj1Name).DeleteActiveWatchers(context.Background(), watcher2Name)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Check if the runtime project is deleted successfully.
				gomega.Eventually(func() bool {
					result, err := tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs(org3Name).
						Folders(defaultName).GetProjects(context.Background(), proj1Name)
					return errors.IsNotFound(err) && result == nil
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "runtime project to be deleted successfully")
			})
		})

		ginkgo.When("org is deleted", func() {
			ginkgo.It("should wait until all the activewatchers are deleted", func() {
				// Wait for create to complete.
				time.Sleep(5 * time.Second)

				// Trigger deletion on org object.
				err = orgClient.Update(context.Background())
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Check if the config org is set to 'InProgress'.
				// This will remain in 'InProgress' state until all the active watchers have acknowledged the org.
				gomega.Eventually(func() bool {
					org, err := nexusClient.DynamicClient.Resource(constructOrgGVR()).
						Get(context.Background(), org3HashedName, metav1.GetOptions{})
					if err != nil {
						return false
					}
					status, found, err := unstructured.NestedString(org.Object, "status", "orgStatus", "statusIndicator")
					if err != nil || !found {
						return false
					}
					return status == string(orgsv1.StatusIndicationInProgress)
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config org to be set to InProgress")

				// Delete the orgactivewatcher.
				err = tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs(org3Name).
					DeleteActiveWatchers(context.Background(), watcher1Name)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				gomega.Eventually(func() bool {
					org, err := nexusClient.DynamicClient.Resource(constructOrgGVR()).
						Get(context.Background(), org3HashedName, metav1.GetOptions{})
					if err != nil {
						return false
					}
					message, found, err := unstructured.NestedString(org.Object, "status", "orgStatus", "message")
					if err != nil || !found {
						return false
					}
					return message == "Waiting for watchers [cluster-orchestrator-2] to be deleted"
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config org to be set to InProgress")

				// Delete the orgactivewatcher.
				err = tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs(org3Name).
					DeleteActiveWatchers(context.Background(), watcher2Name)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Check if the runtime org is deleted successfully.
				gomega.Eventually(func() bool {
					result, err := tenancyReconciler.Client.TenancyMultiTenancy().Runtime().
						GetOrgs(context.Background(), org3Name)
					return errors.IsNotFound(err) && result == nil
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "runtime org to be deleted successfully")
			})
		})

		ginkgo.When("activewatcher spec is unchanged/non-IDLE", func() {
			ginkgo.It("should skip processing update event", func() {
				// Create the config org.
				orgClient, err := configClient.AddOrgs(context.Background(), constructOrgObj(org4Name))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(orgClient).NotTo(gomega.BeNil())

				// Create the orgactivewatcher.
				watcherObj1 := constructOrgActiveWatcherObj(watcher1Name)
				watcherObj1.Spec.StatusIndicator = orgactivewatcherv1.StatusIndicationInProgress
				activeWatcher, err := tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs(org4Name).
					AddActiveWatchers(context.Background(), watcherObj1)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				activeWatcher.Spec.StatusIndicator = orgactivewatcherv1.StatusIndicationInProgress
				err = activeWatcher.Update(context.Background())
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Verify if the update event with same spec is skipped.
				gomega.Eventually(func() bool {
					org, err := nexusClient.DynamicClient.Resource(constructOrgGVR()).
						Get(context.Background(), org4HashedName, metav1.GetOptions{})
					if err != nil {
						return false
					}
					status, found, err := unstructured.NestedString(org.Object, "status", "orgStatus", "statusIndicator")
					if err != nil || !found {
						return false
					}
					return status == string(orgsv1.StatusIndicationInProgress)
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config org to be set to 'InProgress'")

				activeWatcher.Spec.StatusIndicator = orgactivewatcherv1.StatusIndicationError
				err = activeWatcher.Update(context.Background())
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Verify if the update event with non-IDLE status is skipped.
				gomega.Eventually(func() bool {
					org, err := nexusClient.DynamicClient.Resource(constructOrgGVR()).
						Get(context.Background(), org4HashedName, metav1.GetOptions{})
					if err != nil {
						return false
					}
					status, found, err := unstructured.NestedString(org.Object, "status", "orgStatus", "statusIndicator")
					if err != nil || !found {
						return false
					}
					return status == string(orgsv1.StatusIndicationInProgress)
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config org to be set to 'InProgress'")

				// Create the config project.
				proj, err := tenancyReconciler.Client.TenancyMultiTenancy().Config().Orgs(org4Name).Folders(defaultName).
					AddProjects(context.Background(), constructProjectObj())
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(proj).NotTo(gomega.BeNil())

				// Create the projectactivewatcher.
				projWatcherObj1 := constructProjectActiveWatcherObj(watcher1Name)
				projWatcherObj1.Spec.StatusIndicator = projectactivewatcherv1.StatusIndicationInProgress
				projectActiveWatcher, err := tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs(org4Name).
					Folders(defaultName).Projects(proj1Name).AddActiveWatchers(context.Background(), projWatcherObj1)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				projectActiveWatcher.Spec.StatusIndicator = projectactivewatcherv1.StatusIndicationInProgress
				err = projectActiveWatcher.Update(context.Background())
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Verify if the update event with same spec is skipped.
				gomega.Eventually(func() bool {
					project, err := nexusClient.DynamicClient.Resource(constructProjectGVR()).
						Get(context.Background(), org4proj1ConfigHash, metav1.GetOptions{})
					if err != nil {
						return false
					}
					status, found, err := unstructured.NestedString(project.Object, "status", "projectStatus", "statusIndicator")
					if err != nil || !found {
						return false
					}
					return status == string(projectv1.StatusIndicationInProgress)
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config project to be set to InProgress")

				// Create the projectactivewatcher.
				projectActiveWatcher.Spec.StatusIndicator = projectactivewatcherv1.StatusIndicationError
				err = projectActiveWatcher.Update(context.Background())
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Verify if the update event with non-IDLE status is skipped.
				gomega.Eventually(func() bool {
					project, err := nexusClient.DynamicClient.Resource(constructProjectGVR()).
						Get(context.Background(), org4proj1ConfigHash, metav1.GetOptions{})
					if err != nil {
						return false
					}
					status, found, err := unstructured.NestedString(project.Object, "status", "projectStatus", "statusIndicator")
					if err != nil || !found {
						return false
					}
					return status == string(projectv1.StatusIndicationInProgress)
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config project to be set to InProgress")
			})
		})

		ginkgo.When("activewatchers haven't acknowledged the object within the configured timeInterval", func() {
			ginkgo.It("should set the config object status to 'Error'", func() {
				// Create the config org.
				orgClient, err := configClient.AddOrgs(context.Background(), &orgsv1.Org{
					ObjectMeta: metav1.ObjectMeta{
						Name:            org2Name,
						ResourceVersion: "1",
					},
				})
				gomega.Expect(orgClient).NotTo(gomega.BeNil())
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Check if the runtime org is created successfully.
				gomega.Eventually(func() bool {
					result, err := tenancyReconciler.Client.TenancyMultiTenancy().Runtime().
						GetOrgs(context.Background(), org2Name)
					if result != nil {
						return result.Name != "" && err == nil
					}
					return false
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "runtime org to be retrieved successfully")

				// Check if the config org is set to Error after timeout.
				gomega.Eventually(func() bool {
					org, err := nexusClient.DynamicClient.Resource(constructOrgGVR()).
						Get(context.Background(), org2HashedName, metav1.GetOptions{})
					if err != nil {
						return false
					}
					status, found, err := unstructured.NestedString(org.Object, "status", "orgStatus", "statusIndicator")
					if err != nil || !found {
						return false
					}
					return status == string(orgsv1.StatusIndicationError)
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config org to be set to 'Error'")

				// Create the config project.
				proj, err := tenancyReconciler.Client.TenancyMultiTenancy().Config().Orgs(org2Name).
					Folders(defaultName).AddProjects(context.Background(), &projectv1.Project{
					ObjectMeta: metav1.ObjectMeta{
						Name:            proj2Name,
						ResourceVersion: "1",
					},
				})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(proj).NotTo(gomega.BeNil())

				// Check if the runtime project is created successfully.
				gomega.Eventually(func() bool {
					result, err := tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs(org2Name).
						Folders(defaultName).GetProjects(context.Background(), proj2Name)
					if result != nil {
						return result.Name != "" && err == nil
					}
					return false
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "runtime project to be retrieved successfully")

				// Check if the config org is set to Error after timeout.
				gomega.Eventually(func() bool {
					project, err := nexusClient.DynamicClient.Resource(constructProjectGVR()).
						Get(context.Background(), org2proj2ConfigHash, metav1.GetOptions{})
					if err != nil {
						return false
					}
					status, found, err := unstructured.NestedString(project.Object, "status", "projectStatus", "statusIndicator")
					if err != nil || !found {
						return false
					}
					return status == string(projectv1.StatusIndicationError)
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config project to be set to 'Error'")
			})
		})

		ginkgo.When("org/project is deleted and activewatchers aren't deleted within the configured timeInterval", func() {
			ginkgo.It("should set the config object status to 'Error'", func() {
				// Create the config org.
				org, err := configClient.AddOrgs(context.Background(), constructOrgObj("balloon"))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				orgToUpdate, err := nexusClient.DynamicClient.Resource(constructOrgGVR()).
					Create(context.Background(), constructUnstructuredOrg(org.Name), metav1.CreateOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(orgToUpdate).NotTo(gomega.BeNil())

				// Create the orgactivewatcher1.
				_, err = tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs("balloon").
					AddActiveWatchers(context.Background(), constructOrgActiveWatcherObj(watcher1Name))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Create the orgactivewatcher2.
				_, err = tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs("balloon").
					AddActiveWatchers(context.Background(), constructOrgActiveWatcherObj(watcher2Name))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Check if the config org is set to IDLE.
				gomega.Eventually(func() bool {
					org, err := nexusClient.DynamicClient.Resource(constructOrgGVR()).
						Get(context.Background(), org.Name, metav1.GetOptions{})
					if err != nil {
						return false
					}
					status, found, err := unstructured.NestedString(org.Object, "status", "orgStatus", "statusIndicator")
					if err != nil || !found {
						return false
					}
					return status == string(orgsv1.StatusIndicationIdle)
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config org to be set to 'IDLE'")

				// Create the config project.
				project, err := tenancyReconciler.Client.TenancyMultiTenancy().Config().Orgs("balloon").Folders(defaultName).
					AddProjects(context.Background(), constructProjectObj())
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				project3ToUpdate, err := nexusClient.DynamicClient.Resource(constructProjectGVR()).
					Create(context.Background(), constructUnstructuredProject(project.Name), metav1.CreateOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(project3ToUpdate).NotTo(gomega.BeNil())

				// Create the projectactivewatcher.
				_, err = tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs("balloon").Folders(defaultName).
					Projects(proj1Name).AddActiveWatchers(context.Background(), constructProjectActiveWatcherObj(watcher1Name))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				_, err = tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs("balloon").Folders(defaultName).
					Projects(proj1Name).AddActiveWatchers(context.Background(), constructProjectActiveWatcherObj(watcher2Name))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Check if the config project is set to active.
				gomega.Eventually(func() bool {
					project, err := nexusClient.DynamicClient.Resource(constructProjectGVR()).
						Get(context.Background(), project.Name, metav1.GetOptions{})
					if err != nil {
						return false
					}
					status, found, err := unstructured.NestedString(project.Object, "status", "projectStatus", "statusIndicator")
					if err != nil || !found {
						return false
					}
					return status == string(projectv1.StatusIndicationIdle)
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config project to be set to IDLE")

				// Trigger deletion on project.
				err = project.Update(context.Background())
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Check if the config project is set to Error after timeout.
				gomega.Eventually(func() bool {
					project, err := nexusClient.DynamicClient.Resource(constructProjectGVR()).
						Get(context.Background(), project.Name, metav1.GetOptions{})
					if err != nil {
						return false
					}
					status, found, err := unstructured.NestedString(project.Object, "status", "projectStatus", "statusIndicator")
					if err != nil || !found {
						return false
					}
					return status == string(projectv1.StatusIndicationError)
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config project to be set to Error")

				// Trigger deletion on org.
				err = org.Update(context.Background())
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Check if the config org is set to Error after timeout.
				gomega.Eventually(func() bool {
					org, err := nexusClient.DynamicClient.Resource(constructOrgGVR()).
						Get(context.Background(), org.Name, metav1.GetOptions{})
					if err != nil {
						return false
					}
					status, found, err := unstructured.NestedString(org.Object, "status", "orgStatus", "statusIndicator")
					if err != nil || !found {
						return false
					}
					return status == string(orgsv1.StatusIndicationError)
				}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config org to be set to 'Error'")
			})
		})

		ginkgo.It("should skip processing orgactivewatcher events if config org is already 'Idle'", func() {
			// Create config object with IDLE state.
			org, err := configClient.AddOrgs(context.Background(), &orgsv1.Org{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "alpha",
					ResourceVersion: "123",
				},
				Status: orgsv1.OrgNexusStatus{
					OrgStatus: orgsv1.OrgStatus{
						StatusIndicator: orgsv1.StatusIndicationIdle,
					},
				},
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			_, err = tenancyReconciler.Client.TenancyMultiTenancy().Runtime().AddOrgs(context.Background(),
				&runtimeorgsv1.RuntimeOrg{
					ObjectMeta: metav1.ObjectMeta{
						Name: org.DisplayName(),
					},
					Spec: runtimeorgsv1.RuntimeOrgSpec{},
				})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Add an activewatcher.
			_, err = tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs("alpha").
				AddActiveWatchers(context.Background(), constructOrgActiveWatcherObj(watcher1Name))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Wait for it to skip the processing.
			time.Sleep(2 * time.Second)

			// Verify by checking if the config org is still in IDLE state.
			gomega.Eventually(func() bool {
				org, err := configClient.GetOrgs(context.Background(), "alpha")
				if err != nil {
					return false
				}
				return org.Status.OrgStatus.StatusIndicator == orgsv1.StatusIndicationIdle
			}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config org to be set to 'IDLE'")
		})

		ginkgo.It("should skip processing projectactivewatcher events if config project is already 'Idle'", func() {
			_, err = tenancyReconciler.Client.TenancyMultiTenancy().Config().Orgs("alpha").
				AddFolders(context.Background(), &folderv1.Folder{
					ObjectMeta: metav1.ObjectMeta{
						Name: "default",
					},
				})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Create the config project.
			project, err := tenancyReconciler.Client.TenancyMultiTenancy().Config().Orgs("alpha").
				Folders(defaultName).AddProjects(context.Background(), &projectv1.Project{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "alpha",
					ResourceVersion: "123",
				},
				Status: projectv1.ProjectNexusStatus{
					ProjectStatus: projectv1.ProjectStatus{
						StatusIndicator: projectv1.StatusIndicationIdle,
					},
				},
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			_, err = tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs("alpha").
				Folders("default").AddProjects(context.Background(), &runtimeprojectsv1.RuntimeProject{
				ObjectMeta: metav1.ObjectMeta{
					Name: project.DisplayName(),
				},
				Spec: runtimeprojectsv1.RuntimeProjectSpec{},
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Add an activewatcher.
			obj := constructProjectActiveWatcherObj(watcher1Name)
			_, err = tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs("alpha").Folders("default").
				Projects(project.DisplayName()).AddActiveWatchers(context.Background(), obj)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Wait for it to skip the processing.
			time.Sleep(2 * time.Second)

			// Verify by checking if the config org is still in IDLE state.
			gomega.Eventually(func() bool {
				project, err := tenancyReconciler.Client.TenancyMultiTenancy().Config().Orgs("alpha").
					Folders(defaultName).GetProjects(context.Background(), project.DisplayName())
				if err != nil {
					return false
				}
				return project.Status.ProjectStatus.StatusIndicator == projectv1.StatusIndicationIdle
			}, timeoutInterval, pollingInterval).Should(gomega.BeTrue(), "config project to be set to 'IDLE'")
		})
	})

	ginkgo.It("should skip processing org add if the status is already 'Idle'", func() {
		_, err := configClient.AddOrgs(context.Background(), &orgsv1.Org{
			ObjectMeta: metav1.ObjectMeta{
				Name: "bar",
			},
			Status: orgsv1.OrgNexusStatus{
				OrgStatus: orgsv1.OrgStatus{
					StatusIndicator: orgsv1.StatusIndicationIdle,
				},
			},
		})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// Wait for it to skip the processing.
		time.Sleep(2 * time.Second)

		// Verify by checking if the runtime object creation is skipped.
		gomega.Eventually(func() bool {
			_, err := tenancyReconciler.Client.TenancyMultiTenancy().Runtime().
				GetOrgs(context.Background(), "bar")
			return errors.IsNotFound(err)
		}, timeoutInterval, pollingInterval).Should(gomega.BeTrue())
	})

	ginkgo.It("should skip processing project add if the status is already 'Idle'", func() {
		_, err := tenancyReconciler.Client.TenancyMultiTenancy().Config().Orgs("bar").
			AddFolders(context.Background(), &folderv1.Folder{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		// Create the config project.
		_, err = tenancyReconciler.Client.TenancyMultiTenancy().Config().Orgs("bar").
			Folders(defaultName).AddProjects(context.Background(), &projectv1.Project{
			ObjectMeta: metav1.ObjectMeta{
				Name: "bar",
			},
			Status: projectv1.ProjectNexusStatus{
				ProjectStatus: projectv1.ProjectStatus{
					StatusIndicator: projectv1.StatusIndicationIdle,
				},
			},
		})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// Wait for it to skip the processing.
		time.Sleep(2 * time.Second)

		// Verify by checking if the runtime object creation is skipped.
		gomega.Eventually(func() bool {
			_, err := tenancyReconciler.Client.TenancyMultiTenancy().Runtime().Orgs("bar").
				Folders(defaultName).GetProjects(context.Background(), "bar")
			return errors.IsNotFound(err)
		}, timeoutInterval, pollingInterval).Should(gomega.BeTrue())
	})
})
