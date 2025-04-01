// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"context"
	"errors"
	"io"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	proto "github.com/open-edge-platform/o11y-tenant-controller/api"
	"google.golang.org/grpc"
)

const (
	reconnectDelay   = 100 * time.Millisecond
	streamErrorDelay = 1 * time.Second
	streamFailDelay  = 5 * time.Second
)

type RoleStore struct {
	staticRoles        []string
	templates          []string
	templatesAvailable bool // templatesAvailable flag is used to determine if there are any templates and if we need to fetch project updates

	mutex      sync.RWMutex
	projectIDs []string

	dynamicRoles atomic.Pointer[[]string]
}

func NewRoleStore(roles []string) *RoleStore {
	var staticRoles []string
	var templates []string

	for _, role := range roles {
		if strings.Contains(role, "{projectId}") {
			templates = append(templates, role)
		} else {
			staticRoles = append(staticRoles, role)
		}
	}
	var templatesAvailable bool
	if len(templates) > 0 {
		templatesAvailable = true
	}

	rs := &RoleStore{
		staticRoles:        staticRoles,
		templates:          templates,
		templatesAvailable: templatesAvailable,
	}
	rs.UpdateDynamicRoles()

	return rs
}

func (rs *RoleStore) SetProjectIDs(ids []string) {
	rs.mutex.Lock()
	rs.projectIDs = ids
	rs.mutex.Unlock()
}

func (rs *RoleStore) UpdateDynamicRoles() {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	expandedRoles := make([]string, 0, len(rs.staticRoles)+len(rs.templates)*len(rs.projectIDs))
	expandedRoles = append(expandedRoles, rs.staticRoles...)
	for _, template := range rs.templates {
		for _, projectID := range rs.projectIDs {
			expanded := strings.ReplaceAll(template, "{projectId}", projectID)
			expandedRoles = append(expandedRoles, expanded)
		}
	}

	rs.dynamicRoles.Store(&expandedRoles)
}

func (rs *RoleStore) GetRoles() []string {
	return *rs.dynamicRoles.Load()
}

func (rs *RoleStore) FetchProjectUpdates(ctx context.Context, tcConn *grpc.ClientConn) {
	tc := proto.NewProjectServiceClient(tcConn)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Context done, exiting project updates fetcher")
			return
		default:
			stream, err := tc.StreamProjectUpdates(ctx, &proto.EmptyRequest{})
			if err != nil {
				log.Printf("Failed to stream project updates: %v", err)
				time.Sleep(streamFailDelay)
				continue
			}

			for {
				update, err := stream.Recv()
				if errors.Is(err, io.EOF) {
					log.Printf("Stream terminated gracefully, reconnecting")
					time.Sleep(reconnectDelay)
					break
				} else if err != nil {
					log.Printf("Failed to receive update: %v", err)
					time.Sleep(streamErrorDelay)
					break
				}
				projects := make([]string, 0, len(update.GetProjects()))
				for _, project := range update.GetProjects() {
					if project.Data.Status == "Created" {
						projects = append(projects, project.GetKey())
					}
				}

				rs.SetProjectIDs(projects)
				rs.UpdateDynamicRoles()
			}
		}
	}
}

func (rs *RoleStore) HasTemplatesAvailable() bool {
	return rs.templatesAvailable
}
