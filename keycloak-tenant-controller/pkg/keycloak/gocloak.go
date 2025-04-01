// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package keycloak

import (
	"context"
	"fmt"

	"github.com/Clarilab/gocloaksession"
	"github.com/Nerzal/gocloak/v13"

	"github.com/open-edge-platform/orch-utils/keycloak-tenant-controller/pkg/auth/secrets"
)

type keycloakSession gocloaksession.GoCloakSession

// https://www.keycloak.org/docs-api/latest/rest-api/index.html#GroupRepresentation
type GroupRepresentation struct {
	ID   string
	Name string
}

/*
startKeycloakSession creates an returns a Keycloak session object
*/
func startKeycloakSession(keycloakRealm, keycloakurl string) (keycloakSession, error) {
	session, err := gocloaksession.NewSession(clientName, secrets.GetClientSecret(), keycloakRealm, keycloakurl)
	if err != nil {
		return nil, fmt.Errorf("failed to create Keycloak session: %v", err)
	}

	return session, nil
}

/*
ensureRealmExists returns true if the specified realm exists and is enabled in Keycloak
*/
func ensureRealmExists(session keycloakSession, ctx context.Context, realmName string) (bool, error) {
	keycloakClient := session.GetGoCloakInstance()

	jwt, err := session.GetKeycloakAuthToken()
	if err != nil {
		return false, fmt.Errorf("failed to get json web token: %v", err)
	}

	realmRepresentation, err := keycloakClient.GetRealm(ctx, jwt.AccessToken, realmName)
	if err != nil {
		return false, fmt.Errorf("failed to get realm representation for realm %s: %v", realmName, err)
	}

	if realmRepresentation == nil {
		return false, fmt.Errorf("realm representation for realm %s is nil", realmName)
	}

	if realmRepresentation.Enabled == nil || !*realmRepresentation.Enabled {
		return false, fmt.Errorf("realm representation for realm %s is nil or not enabled", realmName)
	}

	return true, nil
}

/*
createRole creates a new role with a sepcified name within a specified realm in Keycloak
*/
func createRole(session keycloakSession, ctx context.Context, realm string, role string) error {
	keycloakClient := session.GetGoCloakInstance()

	jwt, err := session.GetKeycloakAuthToken()
	if err != nil {
		return fmt.Errorf("failed to get json web token: %v", err)
	}

	newRole := gocloak.Role{
		Name: &role,
	}

	if _, err := keycloakClient.CreateRealmRole(ctx, jwt.AccessToken, realm, newRole); err != nil {
		return fmt.Errorf("failed to create role %s in realm %s: %v", role, realm, err)
	}

	return nil
}

/*
getRoles returns a list of all roles within a specified realm in Keycloak
*/
func getRoles(session gocloaksession.GoCloakSession, ctx context.Context, realm string) ([]string, error) {
	keycloakClient := session.GetGoCloakInstance()

	jwt, err := session.GetKeycloakAuthToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get json web token: %v", err)
	}

	roles, err := keycloakClient.GetRealmRoles(ctx, jwt.AccessToken, realm, gocloak.GetRoleParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to get roles in realm %s: %v", realm, err)
	}

	var rolesStr []string
	for _, role := range roles {
		rolesStr = append(rolesStr, *role.Name)
	}
	return rolesStr, nil
}

/*
deleteRole deletes a specified role within a specified realm in Keycloak
*/
func deleteRole(session keycloakSession, ctx context.Context, realm string, role string) error {
	keycloakClient := session.GetGoCloakInstance()

	jwt, err := session.GetKeycloakAuthToken()
	if err != nil {
		return fmt.Errorf("failed to get json web token: %v", err)
	}

	if err := keycloakClient.DeleteRealmRole(ctx, jwt.AccessToken, realm, role); err != nil {
		return fmt.Errorf("failed to delete role %s in realm %s: %v", role, realm, err)
	}

	return nil
}

/*
createGroup creates a new group within a specified realm in Keycloak
*/
func createGroup(session keycloakSession, ctx context.Context, realm, groupName string) error {
	keycloakClient := session.GetGoCloakInstance()

	jwt, err := session.GetKeycloakAuthToken()
	if err != nil {
		return fmt.Errorf("failed to get json web token: %v", err)
	}

	group := gocloak.Group{
		Name: &groupName,
	}

	if _, err := keycloakClient.CreateGroup(ctx, jwt.AccessToken, realm, group); err != nil {
		return fmt.Errorf("failed to create group %s in realm %s: %v", groupName, realm, err)
	}

	return nil
}

/*
getGroups returns a list of all groups within a specified realm in Keycloak
*/
func getGroups(session keycloakSession, ctx context.Context, realm string) ([]*GroupRepresentation, error) {
	keycloakClient := session.GetGoCloakInstance()

	jwt, err := session.GetKeycloakAuthToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get json web token: %v", err)
	}

	groups, err := keycloakClient.GetGroups(ctx, jwt.AccessToken, realm, gocloak.GetGroupsParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to get groups in realm %s: %v", realm, err)
	}

	var groupReps []*GroupRepresentation
	for _, group := range groups {
		g := &GroupRepresentation{
			Name: *group.Name,
			ID:   *group.ID,
		}
		groupReps = append(groupReps, g)
	}

	return groupReps, nil
}

/*
deleteGroup deletes a specified group within a specified realm in Keycloak
*/
func deleteGroup(session keycloakSession, ctx context.Context, realm, groupName string) error {
	keycloakClient := session.GetGoCloakInstance()

	jwt, err := session.GetKeycloakAuthToken()
	if err != nil {
		return fmt.Errorf("failed to get json web token: %v", err)
	}

	if err := keycloakClient.DeleteGroup(ctx, jwt.AccessToken, realm, groupName); err != nil {
		return fmt.Errorf("failed to delete group %s in realm %s: %v", groupName, realm, err)
	}

	return nil
}

/*
addRolesToGroup adds a specified list of roles to a specified group within a specified realm in Keycloak
*/
func addRolesToGroup(session keycloakSession, ctx context.Context, realm string, groupName string, roleNames []string) error {
	keycloakClient := session.GetGoCloakInstance()

	jwt, err := session.GetKeycloakAuthToken()
	if err != nil {
		return fmt.Errorf("failed to get json web token: %v", err)
	}

	group, err := getGroupByName(session, ctx, realm, groupName)
	if err != nil {
		return fmt.Errorf("failed to get group by name %s in realm %s: %v", groupName, realm, err)
	}

	roles, err := getRolesByNames(session, ctx, realm, roleNames)
	if err != nil {
		return fmt.Errorf("failed to get roles by names %v in realm %s: %v", roleNames, realm, err)
	}

	if err := keycloakClient.AddRealmRoleToGroup(ctx, jwt.AccessToken, realm, *group.ID, roles); err != nil {
		return fmt.Errorf("failed to add roles to group %s in realm %s: %v", groupName, realm, err)
	}

	return nil
}

/*
getRolesByNames returns an array of role objects that match specified role names within a specified realm in Keycloak
*/
func getRolesByNames(session keycloakSession, ctx context.Context, realm string, roleNames []string) ([]gocloak.Role, error) {
	keycloakClient := session.GetGoCloakInstance()

	jwt, err := session.GetKeycloakAuthToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get json web token: %v", err)
	}

	var rolesToAdd []gocloak.Role

	roles, err := keycloakClient.GetRealmRoles(ctx, jwt.AccessToken, realm, gocloak.GetRoleParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to get roles in realm %s: %v", realm, err)
	}

	for _, roleName := range roleNames {
		var roleExists bool
		for _, role := range roles {
			if *role.Name == roleName {
				rolesToAdd = append(rolesToAdd, *role)
				roleExists = true
				break
			}
		}
		if !roleExists {
			return nil, fmt.Errorf("role %s not found in realm %s", roleName, realm)
		}
	}

	return rolesToAdd, nil
}

/*
getGroupByName returns a role objects that matches a specified group name within a specified realm in Keycloak
*/
func getGroupByName(session keycloakSession, ctx context.Context, realm, groupName string) (*gocloak.Group, error) {
	keycloakClient := session.GetGoCloakInstance()

	jwt, err := session.GetKeycloakAuthToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get json web token: %v", err)
	}

	params := gocloak.GetGroupsParams{
		Search: &groupName,
	}

	groups, err := keycloakClient.GetGroups(ctx, jwt.AccessToken, realm, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups in realm %s: %v", realm, err)
	}

	for _, group := range groups {
		if *group.Name == groupName {
			return group, nil
		}
	}

	return nil, fmt.Errorf("group %s not found in realm %s", groupName, realm)
}
