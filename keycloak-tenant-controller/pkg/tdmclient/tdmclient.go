// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package tdmclient

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/open-edge-platform/orch-utils/keycloak-tenant-controller/pkg/keycloak"
	"github.com/open-edge-platform/orch-utils/keycloak-tenant-controller/pkg/log"
	orgActiveWatcherv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/orgactivewatcher.edge-orchestrator.intel.com/v1"
	orgwatcherv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/orgwatcher.edge-orchestrator.intel.com/v1"
	projectActiveWatcherv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/projectactivewatcher.edge-orchestrator.intel.com/v1"
	projectwatcherv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/projectwatcher.edge-orchestrator.intel.com/v1"
	nexus_client "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/nexus-client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	watcherTimeout = 60 * time.Second
)

type TdmClient interface {
	Init() error
	Stop()
}

type tdmclient struct {
	nexusClient *nexus_client.Clientset
	appName     string
	kcClient    keycloak.Client
}

func NewMTClient(appName string, kcClient keycloak.Client) TdmClient {
	return &tdmclient{
		appName:  appName,
		kcClient: kcClient,
	}
}

// Callback function to be invoked when Org is added.
func (tc *tdmclient) processRuntimeOrgsAdd(org *nexus_client.RuntimeorgRuntimeOrg) {
	log.Infof("Processing RuntimeOrgsAdd for: %+v\n", *org)

	// Get watcher object if it exists and set the status to IN-PROGRESS.
	err := tc.updateOrgWatcherStatus(org, orgActiveWatcherv1.StatusIndicationInProgress, "Creating")

	// If watcher does not exist, register this app as an active watcher for this org.
	if nexus_client.IsChildNotFound(err) {
		err := tc.setOrgActiveWatcher(org)
		if err != nil {
			log.Errorf("Failed to register OrgActiveWatcher: %v", err)
			return
		}
	} else if err != nil {
		log.Errorf("Error %+v while fetching watcher %s for org %s\n", err, tc.appName, org.DisplayName())
		return
	}

	err = tc.kcClient.CreateOrg(string(org.UID))
	if err != nil {
		log.Errorf("Failed to create org %s in Keycloak with an error: %v", org.DisplayName(), err)
		return
	}

	// After processing, set watcher object status to IDLE.
	setStatusErr := tc.updateOrgWatcherStatus(org, orgActiveWatcherv1.StatusIndicationIdle, "Created")
	if setStatusErr != nil {
		log.Errorf("Failed to update OrgActiveWatcher object with an error: %v", setStatusErr)
		return
	}
	log.Debugf("Active watcher updated for org: %s\n", org.DisplayName())

	log.Infof("RuntimeOrgsAdd event handled for: %+v\n", *org)
}

// Callback function to be invoked when Org is deleted.
func (tc *tdmclient) processRuntimeOrgsUpdate(_, org *nexus_client.RuntimeorgRuntimeOrg) {
	log.Infof("Processing RuntimeOrgsUpdate for: %+v\n", *org)

	if org.Spec.Deleted {
		log.Debugf("Orgs: %+v marked for deletion\n", org.DisplayName())

		tc.kcClient.DeleteOrg(string(org.UID))

		// Stop watching the org as it is marked for deletion.
		err := org.DeleteActiveWatchers(context.Background(), tc.appName)
		if nexus_client.IsChildNotFound(err) {
			// This app has already stopped watching the org.
			log.Debugf("App %s DOES NOT watch org %s\n", tc.appName, org.DisplayName())
			return
		} else if err != nil {
			log.Debugf("Error %+v while deleting watch %s for org %s\n", err, tc.appName, org.DisplayName())
			return
		}
		log.Debugf("Active watcher %s deleted for Org %s\n", tc.appName, org.DisplayName())
	}

	log.Infof("RuntimeOrgsUpdate event handled for: %+v\n", *org)
}

// Callback function to be invoked when Project is added.
func (tc *tdmclient) processRuntimeProjectsAdd(proj *nexus_client.RuntimeprojectRuntimeProject) {
	log.Debugf("Processing RuntimeProjectsAdd for: %+v\n", *proj)

	// Get watcher object if it exists and set the status to IN-PROGRESS.
	err := tc.updateProjWatcherStatus(proj, projectActiveWatcherv1.StatusIndicationInProgress, "Creating")

	// If watcher does not exist, register this app as an active watcher for this proj.
	if nexus_client.IsChildNotFound(err) {
		log.Infof("Creating new ProjectActiveWatcher for proj: %s", proj.DisplayName())
		err := tc.setProjectActiveWatcher(proj)
		if err != nil {
			log.Errorf("Failed to register ProjectActiveWatcher: %v", err)
			return
		}
	} else if err != nil {
		log.Errorf("Error %+v while fetching watcher %s for project %s\n", err, tc.appName, proj.DisplayName())
		return
	}

	folderOrgs, err := proj.GetParent(context.Background())
	if err != nil {
		fmt.Printf("Error while creating looking up Iam runtime object: %v\n", err)
		return
	}

	org, err := folderOrgs.GetParent(context.Background())
	if err != nil {
		fmt.Printf("Error while looking up Iam runtime org object: %v\n", err)
		return
	}

	err = tc.kcClient.CreateProject(string(org.UID), string(proj.UID))
	if err != nil {
		log.Errorf("Failed to create project %s in Keycloak with an error: %v", proj.DisplayName(), err)
		return
	}

	// After processing, set the status to IDLE.
	setStatusErr := tc.updateProjWatcherStatus(proj, projectActiveWatcherv1.StatusIndicationIdle, "Created")
	if setStatusErr != nil {
		log.Errorf("Failed to update ProjectActiveWatcher object with an error: %v", setStatusErr)
		return
	}
	log.Debugf("Active watcher updated for Project %s\n", proj.DisplayName())

	log.Infof("RuntimeProjectsAdd event handled for: %+v\n", *proj)
}

// Callback function to be invoked when Project is deleted.
func (tc *tdmclient) processRuntimeProjectsUpdate(_, proj *nexus_client.RuntimeprojectRuntimeProject) {
	log.Infof("Processing RuntimeProjectsUpdate for: %+v\n", *proj)

	if proj.Spec.Deleted {
		log.Debugf("Project: %+v marked for deletion\n", proj.DisplayName())

		err := tc.kcClient.DeleteProject(string(proj.UID))
		if err != nil {
			log.Errorf("Failed to delete project %s in Keycloak with an error: %v", proj.DisplayName(), err)
			return
		}

		// Stop watching the project as it is marked for deletion.
		err = proj.DeleteActiveWatchers(context.Background(), tc.appName)
		if nexus_client.IsChildNotFound(err) {
			// This app has already stopped watching the project.
			log.Debugf("App %s DOES NOT watch project %s\n", tc.appName, proj.DisplayName())
			return
		} else if err != nil {
			log.Debugf("Error %+v while deleting watch %s for project %s\n", err, tc.appName, proj.DisplayName())
			return
		}
		log.Debugf("Active watcher %s deleted for project %s\n", tc.appName, proj.DisplayName())
	}

	log.Infof("RuntimeProjectsUpdate event handled for: %+v\n", *proj)
}

func (tc *tdmclient) Init() error {

	// Initialize Nexus SDK, by pointing it to the K8s API endpoint where CRD's are to be stored.
	config, err := getK8sConfig()
	if err != nil {
		log.Errorf("error getting kubeconfig %s", err.Error())
		return fmt.Errorf("error getting kubeconfig %s", err.Error())
	}

	tc.nexusClient, _ = nexus_client.NewForConfig(config)

	// Subscribe to Multi-Tenancy graph.
	// Subscribe() api empowers subscription to objects from datamodel.
	// What subscription does is to keep the local cache in sync with datamodel changes.
	// This sync is done in the background.
	tc.nexusClient.SubscribeAll()

	// Create a watcher for Project
	if err := tc.addProjectWatcher(); err != nil {
		return fmt.Errorf("failed to create project watcher: %w", err)
	}

	// Create a watcher for Org
	if err := tc.addOrgWatcher(); err != nil {
		return fmt.Errorf("failed to create org watcher: %w", err)
	}

	// API to subscribe and register a callback function that is invoked when a Org is added in the datamodel.
	// Register*Callback() has the effect of subscription and also invoking a callback to the application code
	// when there are datamodel changes to the objects of interest.
	tc.nexusClient.TenancyMultiTenancy().Runtime().Orgs("*").RegisterAddCallback(tc.processRuntimeOrgsAdd)
	tc.nexusClient.TenancyMultiTenancy().Runtime().Orgs("*").RegisterUpdateCallback(tc.processRuntimeOrgsUpdate)

	// API to subscribe and register a callback function that is invoked when a Project is added in the datamodel.
	// Register*Callback() has the effect of subscription and also invoking a callback to the application code
	// when there are datamodel changes to the objects of interest.
	tc.nexusClient.TenancyMultiTenancy().Runtime().Orgs("*").Folders("*").Projects("*").RegisterAddCallback(tc.processRuntimeProjectsAdd)
	tc.nexusClient.TenancyMultiTenancy().Runtime().Orgs("*").Folders("*").Projects("*").RegisterUpdateCallback(tc.processRuntimeProjectsUpdate)

	return nil
}

func (tc *tdmclient) setOrgActiveWatcher(org *nexus_client.RuntimeorgRuntimeOrg) error {
	_, err := org.AddActiveWatchers(context.Background(), &orgActiveWatcherv1.OrgActiveWatcher{
		ObjectMeta: metav1.ObjectMeta{
			Name: tc.appName,
		},
		Spec: orgActiveWatcherv1.OrgActiveWatcherSpec{
			StatusIndicator: orgActiveWatcherv1.StatusIndicationInProgress,
			Message:         "Creating",
			TimeStamp:       safeUnixTime(),
		},
	})
	if err != nil {
		log.Errorf("Error while creating watcher %s for org %s\n", tc.appName, org.DisplayName())
		return err
	}

	return nil
}

func (tc *tdmclient) setProjectActiveWatcher(proj *nexus_client.RuntimeprojectRuntimeProject) error {
	_, err := proj.AddActiveWatchers(context.Background(), &projectActiveWatcherv1.ProjectActiveWatcher{
		ObjectMeta: metav1.ObjectMeta{
			Name: tc.appName,
		},
		Spec: projectActiveWatcherv1.ProjectActiveWatcherSpec{
			StatusIndicator: projectActiveWatcherv1.StatusIndicationInProgress,
			Message:         "Creating",
			TimeStamp:       safeUnixTime(),
		},
	})
	if err != nil {
		log.Errorf("Error while creating watcher %s for project %s\n", tc.appName, proj.DisplayName())
		return err
	}

	return nil
}

func (tc *tdmclient) updateOrgWatcherStatus(org *nexus_client.RuntimeorgRuntimeOrg, statusInd orgActiveWatcherv1.ActiveWatcherStatus, status string) error {
	// Get the watcher object.
	watcherObj, err := org.GetActiveWatchers(context.Background(), tc.appName)
	if err == nil && watcherObj != nil {
		// If watcher exists and has same StatusIndicator, simply return.
		if watcherObj.Spec.StatusIndicator == statusInd {
			log.Infof("Skipping processing of orgactivewatcher %v as it is already created and set to %s", tc.appName, statusInd)
			return nil
		}

		setStatusErr := tc.setOrgWatcherStatus(watcherObj, statusInd, status)
		if setStatusErr != nil {
			log.Errorf("Failed to update OrgActiveWatcher object with an error: %v", setStatusErr)
			return setStatusErr
		}
		log.Infof("OrgActiveWatcher %v is set to %s", tc.appName, statusInd)
		return nil
	} else {
		return err
	}
}

func (tc *tdmclient) setOrgWatcherStatus(watcherObj *nexus_client.OrgactivewatcherOrgActiveWatcher, statusInd orgActiveWatcherv1.ActiveWatcherStatus, status string) error {

	watcherObj.Spec.StatusIndicator = statusInd
	watcherObj.Spec.Message = status
	watcherObj.Spec.TimeStamp = safeUnixTime()
	log.Debugf("OrgWatcher object to update: %+v", watcherObj)

	err := watcherObj.Update(context.Background())
	if err != nil {
		log.Errorf("Failed to update OrgActiveWatcher object with an error: %v", err)
		return err
	}
	return nil
}

func (tc *tdmclient) updateProjWatcherStatus(proj *nexus_client.RuntimeprojectRuntimeProject, statusInd projectActiveWatcherv1.ActiveWatcherStatus, status string) error {

	watcherObj, err := proj.GetActiveWatchers(context.Background(), tc.appName)
	if err == nil && watcherObj != nil {
		// If watcher exists and has same StatusIndicator, simply return.
		if watcherObj.Spec.StatusIndicator == statusInd {
			log.Infof("Skipping processing of projectactivewatcher %v as it is already created and set to %s", tc.appName, statusInd)
			return nil
		}

		setStatusErr := tc.setProjWatcherStatus(watcherObj, statusInd, status)
		if setStatusErr != nil {
			log.Errorf("Failed to update ProjectActiveWatcher object with an error: %v", setStatusErr)
			return setStatusErr
		}
		log.Infof("ProjectActiveWatcher %v is set to %s", tc.appName, statusInd)
		return nil
	} else {
		return err
	}
}

func (tc *tdmclient) setProjWatcherStatus(watcherObj *nexus_client.ProjectactivewatcherProjectActiveWatcher, statusInd projectActiveWatcherv1.ActiveWatcherStatus, status string) error {

	watcherObj.Spec.StatusIndicator = statusInd
	watcherObj.Spec.Message = status
	watcherObj.Spec.TimeStamp = safeUnixTime()
	log.Debugf("ProjWatcher object to update: %+v", watcherObj)

	err := watcherObj.Update(context.Background())
	if err != nil {
		log.Errorf("Failed to update ProjectActiveWatcher object with an error: %v", err)
		return err
	}
	return nil
}

func (tc *tdmclient) Stop() {
	tc.nexusClient.UnsubscribeAll()
	if err := tc.deleteProjectWatcher(); err != nil {
		log.Infof("Failed to delete Project watcher: %v", err)
	}

	if err := tc.deleteOrgWatcher(); err != nil {
		log.Infof("Failed to delete Org watcher: %v", err)
	}
}

func (tc *tdmclient) addOrgWatcher() error {
	ctx, cancel := context.WithTimeout(context.Background(), watcherTimeout)
	defer cancel()

	_, err := tc.nexusClient.TenancyMultiTenancy().Config().AddOrgWatchers(ctx, &orgwatcherv1.OrgWatcher{ObjectMeta: metav1.ObjectMeta{
		Name: tc.appName,
	}})

	if nexus_client.IsAlreadyExists(err) {
		log.Infof("Org watcher already exists")
	} else if err != nil {
		return err
	}
	log.Infof("Org watcher is created")

	return nil
}

func (tc *tdmclient) deleteOrgWatcher() error {
	ctx, cancel := context.WithTimeout(context.Background(), watcherTimeout)
	defer cancel()

	err := tc.nexusClient.TenancyMultiTenancy().Config().DeleteOrgWatchers(ctx, tc.appName)

	if nexus_client.IsChildNotFound(err) {
		log.Infof("Org watcher already deleted")
	} else if err != nil {
		return err
	}
	return nil
}

func (tc *tdmclient) addProjectWatcher() error {
	ctx, cancel := context.WithTimeout(context.Background(), watcherTimeout)
	defer cancel()

	_, err := tc.nexusClient.TenancyMultiTenancy().Config().AddProjectWatchers(ctx, &projectwatcherv1.ProjectWatcher{ObjectMeta: metav1.ObjectMeta{
		Name: tc.appName,
	}})

	if nexus_client.IsAlreadyExists(err) {
		log.Infof("Project watcher already exists")
	} else if err != nil {
		return err
	}
	log.Infof("Project watcher is created")

	return nil
}

func (tc *tdmclient) deleteProjectWatcher() error {
	ctx, cancel := context.WithTimeout(context.Background(), watcherTimeout)
	defer cancel()

	err := tc.nexusClient.TenancyMultiTenancy().Config().DeleteProjectWatchers(ctx, tc.appName)

	if nexus_client.IsChildNotFound(err) {
		log.Infof("Project watcher already deleted")
	} else if err != nil {
		return err
	}
	return nil
}

func getK8sConfig() (*rest.Config, error) {
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfigPath := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(kubeconfigPath); err == nil {
			kubeconfig = kubeconfigPath
			log.Infof("k8s rest client using kubeconfig file:s %s", kubeconfigPath)
		}
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	log.Infof("k8s rest client using service account")
	return config, nil
}

func safeUnixTime() uint64 {
	t := time.Now().Unix()
	if t < 0 {
		return 0
	}
	return uint64(t)
}
