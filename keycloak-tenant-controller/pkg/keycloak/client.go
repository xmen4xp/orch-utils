// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/open-edge-platform/orch-utils/keycloak-tenant-controller/pkg/log"
)

const (
	clientName = "ktc-m2m-client"

	defaultKeycloakUrl   = "http://platform-keycloak.orch-platform:8080"
	defaultKeycloakRealm = "master"

	envKeycloakUrl   = "KEYCLOAK_URL"
	envKeycloakRealm = "KEYCLOAK_REALM"
	envOrgGroups     = "KEYCLOAK_ORG_GROUPS"
	envProjGroups    = "KEYCLOAK_PROJ_GROUPS"

	orgPrefix  = "<org-id>"
	projPrefix = "<project-id>"

	retryAttempts  = 10
	retrySleep     = 3 * time.Second
	defaultTimeout = 60 * time.Second // should be > retryAttempts * retrySleep
)

type Client interface {
	Init() error
	CreateOrg(orgId string) error
	DeleteOrg(orgId string) error
	CreateProject(orgId string, projId string) error
	DeleteProject(projId string) error
}

type client struct {
	mu            sync.Mutex
	session       keycloakSession
	orgGroups     map[string][]string
	projGroups    map[string][]string
	siGroups      map[string][]string
	keycloakurl   string
	keycloakRealm string
}

func NewClient() Client {
	return &client{
		keycloakurl:   defaultKeycloakUrl,
		keycloakRealm: defaultKeycloakRealm,
	}
}

/*
Init initializes the Keycloak client
*/
func (c *client) Init() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.init()
}

/*
CreateOrg will create the appropriate roles and groups for a new org.
The roles and groups are read from a JSON environment variable, defined above.
Roles and groups containing an org id prefix will have that prefix replaced with the orgID passed into this function.
*/
func (c *client) CreateOrg(orgID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.createRolesAndGroups(orgID, "", c.orgGroups)
}

/*
DeleteOrg takes an orgID and will delete all groups and roles prepended with this ID.
*/
func (c *client) DeleteOrg(orgID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.deleteRolesAndGroups(orgID)
}

/*
CreateProject will create the appropriate roles and groups for a new project.
The roles and groups are read from a JSON environment variable, defined above.
"KEYCLOAK_PROJ_GROUPS" at time of writing
Roles and groups containing a project id prefix will have that prefix replaced with the projectID passed into this function.
*/
func (c *client) CreateProject(orgID string, projID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.createRolesAndGroups(orgID, projID, c.projGroups)
}

/*
DeleteProject takes a projID and will delete all groups and roles prepended with this ID.
*/
func (c *client) DeleteProject(projID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.deleteRolesAndGroups(projID)
}

/*
init does the work for Init()
*/
func (c *client) init() error {
	if envVar := os.Getenv(envKeycloakUrl); envVar != "" {
		c.keycloakurl = envVar
	}
	log.Infof("Keycloak URL: %s", c.keycloakurl)

	if envVar := os.Getenv(envKeycloakRealm); envVar != "" {
		c.keycloakRealm = envVar
	}
	log.Infof("Keycloak realm: %s", c.keycloakRealm)

	if err := json.Unmarshal([]byte(os.Getenv(envOrgGroups)), &c.orgGroups); err != nil {
		log.Errorf("Error unmarshalling org groups: %v", err)
	}
	log.Infof("Per org groups:")
	for groupName, roleNames := range c.orgGroups {
		log.Infof("   %s", groupName)
		for _, roleName := range roleNames {
			log.Infof("      %s", roleName)
		}
	}

	if err := json.Unmarshal([]byte(os.Getenv(envProjGroups)), &c.projGroups); err != nil {
		log.Errorf("Error unmarshalling proj groups: %v", err)
	}
	log.Infof("Per proj groups:")
	for groupName, roleNames := range c.projGroups {
		log.Infof("   %s", groupName)
		for _, roleName := range roleNames {
			log.Infof("      %s", roleName)
		}
	}

	var err error
	c.session, err = startKeycloakSession(c.keycloakRealm, c.keycloakurl)
	if err != nil {
		log.Errorf("Error starting keycloak session: %v", err)
		return err
	}

	if err := c.createRolesAndGroups("", "", c.siGroups); err != nil {
		log.Errorf("Error creating cross SI groups: %v", err)
	}

	return nil
}

/*
createRolesAndGroups does the work for CreateOrg() and CreateProject()
*/
func (c *client) createRolesAndGroups(orgID string, projID string, groups map[string][]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	_, err := ensureRealmExists(c.session, ctx, c.keycloakRealm)
	if err != nil {
		log.Errorf("Error checking if realm %s exists: %v", c.keycloakRealm, err)
		return err
	}

	var count int
	for {
		for groupName, roleNames := range groups {
			groupName = strings.ReplaceAll(groupName, orgPrefix, orgID)
			groupName = strings.ReplaceAll(groupName, projPrefix, projID)

			if err = createGroup(c.session, ctx, c.keycloakRealm, groupName); err != nil {
				if isErrorForAlreadyExists(err) {
					log.Infof("Group %s already exists", groupName)
				} else {
					log.Errorf("Error creating group %s: %v", groupName, err)
					return err
				}
			} else {
				log.Infof("Group %s created", groupName)
			}

			var updatedRoleNames []string
			for _, roleName := range roleNames {
				roleName = strings.ReplaceAll(roleName, orgPrefix, orgID)
				roleName = strings.ReplaceAll(roleName, projPrefix, projID)

				if err = createRole(c.session, ctx, c.keycloakRealm, roleName); err != nil {
					if isErrorForAlreadyExists(err) {
						log.Infof("Role %s already exists", roleName)
					} else {
						log.Errorf("Error creating role %s: %v", roleName, err)
						return err
					}
				} else {
					log.Infof("Role %s created", roleName)
				}

				updatedRoleNames = append(updatedRoleNames, roleName)
			}

			if err := addRolesToGroup(c.session, ctx, c.keycloakRealm, groupName, updatedRoleNames); err != nil {
				log.Errorf("Error adding roles to group %s : %v", groupName, err)
				return err
			}
		}

		if err := c.checkRolesAndGroupsCreated(orgID, projID, groups); err == nil {
			log.Infof("Roles and groups created successfully - Org: %s Proj: %s", orgID, projID)
			break
		}

		log.Errorf("%v", err)

		count++
		if count > retryAttempts {
			log.Errorf("Exceed retry attempts")
			return fmt.Errorf("role or group does not exist in Keycloak after %d attempts to create", retryAttempts+1)
		}

		log.Infof("Attempting retry %d of %d", count, retryAttempts)
		time.Sleep(retrySleep) // short sleep to allow Keycloak to recover
	}

	return nil
}

/*
deleteRolesAndGroups does the work for DeleteOrg() and DeleteProject()
*/
func (c *client) deleteRolesAndGroups(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	_, err := ensureRealmExists(c.session, ctx, c.keycloakRealm)
	if err != nil {
		log.Errorf("Error checking if realm %s exists: %v", c.keycloakRealm, err)
		return err
	}

	roles, err := getRoles(c.session, ctx, c.keycloakRealm)
	if err != nil {
		log.Errorf("Error getting Keycloak roles: %v", err)
		return err
	}

	for _, role := range roles {
		if strings.Contains(role, id) {
			if err = deleteRole(c.session, ctx, c.keycloakRealm, role); err != nil {
				log.Errorf("Error deleting role %s: %v", role, err)
				return err
			} else {
				log.Infof("Role %s deleted", role)
			}
		}
	}

	groups, err := getGroups(c.session, ctx, c.keycloakRealm)
	if err != nil {
		log.Errorf("Error getting Keycloak groups: %v", err)
		return err
	}

	for _, group := range groups {
		if strings.Contains(group.Name, id) {
			if err = deleteGroup(c.session, ctx, c.keycloakRealm, group.ID); err != nil {
				log.Errorf("Error deleting group %s: %v", group, err)
				return err
			} else {
				log.Infof("Group %s deleted", group.Name)
			}
		}
	}

	return nil
}

/*
checkRolesAndGroupsCreated was introduced after discovering a bug where Keycloak returns 200, but roles and groups were not created
*/
func (c *client) checkRolesAndGroupsCreated(orgID string, projID string, groups map[string][]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	contains := func(existingGroups []*GroupRepresentation, groupName string) bool {
		for _, existingGroup := range existingGroups {
			if existingGroup.Name == groupName {
				return true
			}
		}
		return false
	}

	existingGroups, err := getGroups(c.session, ctx, c.keycloakRealm)
	if err != nil {
		return err
	}

	existingRoles, err := getRoles(c.session, ctx, c.keycloakRealm)
	if err != nil {
		return err
	}

	for groupName, roleNames := range groups {
		groupName = strings.ReplaceAll(groupName, orgPrefix, orgID)
		groupName = strings.ReplaceAll(groupName, projPrefix, projID)

		if !contains(existingGroups, groupName) {
			return fmt.Errorf("keycloak realm %s does not contain group %s", c.keycloakRealm, groupName)
		}

		for _, roleName := range roleNames {
			roleName = strings.ReplaceAll(roleName, orgPrefix, orgID)
			roleName = strings.ReplaceAll(roleName, projPrefix, projID)
			if !slices.Contains(existingRoles, roleName) {
				return fmt.Errorf("keycloak realm %s does not contain role %s", c.keycloakRealm, roleName)
			}
		}
	}

	return nil
}

/*
isErrorForAlreadyExists returns true if the error provided is a 409 "already exists"
*/
func isErrorForAlreadyExists(err error) bool {
	return strings.Contains(err.Error(), "409")
}
