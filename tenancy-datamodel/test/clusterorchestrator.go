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

package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	orgActiveWatcherv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/orgactivewatcher.edge-orchestrator.intel.com/v1"
	projectActiveWatcherv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/projectactivewatcher.edge-orchestrator.intel.com/v1"
	nexus_client "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/nexus-client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Handle to initialized nexus sdk client.
var (
	nexusClient *nexus_client.Clientset
	appName     string = "cluster-orchestrator"
)

// Callback function to be invoked when Org is added.
func processRuntimeOrgsAdd(org *nexus_client.RuntimeorgRuntimeOrg) {
	fmt.Printf("Runtime Orgs: %+v created\n", *org)

	//  **********************************************************************
	//   BUSINESS LOGIC: Implement ORG Creation Handling
	//  **********************************************************************

	// Register this app as an active watcher for this org.
	watcherObj, err := org.AddActiveWatchers(context.Background(), &orgActiveWatcherv1.OrgActiveWatcher{
		ObjectMeta: metav1.ObjectMeta{
			Name: appName,
		},
	})

	if nexus_client.IsAlreadyExists(err) {
		fmt.Printf("watch %s already exists for org %s\n", watcherObj.DisplayName(), org.DisplayName())
	} else if err != nil {
		fmt.Printf("Error %+v while creating watch %s for org %s\n", err, appName, org.DisplayName())
	}

	fmt.Printf("Active watcher %s created for Org %s\n", watcherObj.DisplayName(), org.DisplayName())
}

// Callback function to be invoked when Org is deleted.
func processRuntimeOrgsUpdate(_, newObj *nexus_client.RuntimeorgRuntimeOrg) {
	if newObj.Spec.Deleted == true {
		fmt.Printf("Orgs: %+v marked for deletion\n", newObj.DisplayName())

		//  **********************************************************************
		//   BUSINESS LOGIC: Implement ORG Deletion Handling
		//  **********************************************************************

		// Stop watching the org as it is marked for deletion.
		err := newObj.DeleteActiveWatchers(context.Background(), appName)
		if nexus_client.IsChildNotFound(err) {
			// This app has already stopped watching the org.
			fmt.Printf("App %s DOES NOT watch org %s\n", appName, newObj.DisplayName())
			return
		} else if err != nil {
			fmt.Printf("Error %+v while deleting watch %s for org %s\n", err, appName, newObj.DisplayName())
			return
		}
		fmt.Printf("Active watcher %s deleted for Org %s\n", appName, newObj.DisplayName())
	}
}

// Callback function to be invoked when Project is added.
func processRuntimeProjectsAdd(proj *nexus_client.RuntimeprojectRuntimeProject) {
	fmt.Printf("Runtime Project: %+v created\n", *proj)

	//  **********************************************************************
	//   BUSINESS LOGIC: Implement Project Creation Handling
	//  **********************************************************************

	// Register this app as an active watcher for this project.
	watcherObj, err := proj.AddActiveWatchers(context.Background(), &projectActiveWatcherv1.ProjectActiveWatcher{
		ObjectMeta: metav1.ObjectMeta{
			Name: appName,
		},
	})

	if nexus_client.IsAlreadyExists(err) {
		fmt.Printf("watch %s already exists for project %s\n", watcherObj.DisplayName(), proj.DisplayName())
	} else if err != nil {
		fmt.Printf("Error %+v while creating watch %s for project %s\n", err, appName, proj.DisplayName())
	}

	fmt.Printf("Active watcher %s created for Project %s\n", watcherObj.DisplayName(), proj.DisplayName())
}

// Callback function to be invoked when Project is deleted.
func processRuntimeProjectsUpdate(_, newObj *nexus_client.RuntimeprojectRuntimeProject) {
	if newObj.Spec.Deleted == true {
		fmt.Printf("Project: %+v marked for deletion\n", newObj.DisplayName())

		//  **********************************************************************
		//   BUSINESS LOGIC: Implement PROJECT Deletion Handling
		//  **********************************************************************

		// Stop watching the project as it is marked for deletion.
		err := newObj.DeleteActiveWatchers(context.Background(), appName)
		if nexus_client.IsChildNotFound(err) {
			// This app has already stopped watching the project.
			fmt.Printf("App %s DOES NOT watch project %s\n", appName, newObj.DisplayName())
			return
		} else if err != nil {
			fmt.Printf("Error %+v while deleting watch %s for project %s\n", err, appName, newObj.DisplayName())
			return
		}
		fmt.Printf("Active watcher %s deleted for project %s\n", appName, newObj.DisplayName())
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	var kubeconfig string
	flag.StringVar(&kubeconfig, "k", "", "Absolute path to the kubeconfig file. Defaults to ~/.kube/config.")
	flag.Parse()

	// Initialize Nexus SDK, by pointing it to the K8s API endpoint where CRD's are to be stored.
	var config *rest.Config
	if len(kubeconfig) != 0 {
		var err error
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err)
		}
	} else {
		config = &rest.Config{Host: "localhost:9000"}
	}
	nexusClient, _ = nexus_client.NewForConfig(config)

	// Subscribe to Multi-Tenancy graph.
	// Subscribe() api empowers subscription to objects from datamodel.
	// What subscription does is to keep the local cache in sync with datamodel changes.
	// This sync is done in the background.
	nexusClient.SubscribeAll()

	// API to subscribe and register a callback function that is invoked when a Org is added in the datamodel.
	// Register*Callback() has the effect of subscription and also invoking a callback to the application code
	// when there are datamodel changes to the objects of interest.
	nexusClient.TenancyMultiTenancy().Runtime().Orgs("*").RegisterAddCallback(processRuntimeOrgsAdd)
	nexusClient.TenancyMultiTenancy().Runtime().Orgs("*").RegisterUpdateCallback(processRuntimeOrgsUpdate)

	// API to subscribe and register a callback function that is invoked when a Project is added in the datamodel.
	// Register*Callback() has the effect of subscription and also invoking a callback to the application code
	// when there are datamodel changes to the objects of interest.
	nexusClient.TenancyMultiTenancy().Runtime().Orgs("*").Folders("*").Projects("*").RegisterAddCallback(processRuntimeProjectsAdd)
	nexusClient.TenancyMultiTenancy().Runtime().Orgs("*").Folders("*").Projects("*").RegisterUpdateCallback(processRuntimeProjectsUpdate)

	// Dummy code to make the program block, while it waits for events.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)

	go func() {
		<-sigs
		done <- true
	}()
	<-done
	fmt.Println("exiting")
}
