// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package internal_test

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	proto "github.com/open-edge-platform/o11y-tenant-controller/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/open-edge-platform/orch-utils/auth-service/internal"
)

var (
	lis        *bufconn.Listener
	conn       *grpc.ClientConn
	server     *grpc.Server
	mockServer *mockStreamingServer
	ctx        context.Context
	cancel     context.CancelFunc
)

type mockStreamingServer struct {
	proto.UnimplementedProjectServiceServer

	mutex    *sync.RWMutex
	projects []*proto.ProjectEntry
}

func newMockStreamingServer() *mockStreamingServer {
	return &mockStreamingServer{
		mutex:    &sync.RWMutex{},
		projects: make([]*proto.ProjectEntry, 0),
	}
}

func (m *mockStreamingServer) addProjects(projects ...*proto.ProjectEntry) {
	mockServer.mutex.Lock()
	defer mockServer.mutex.Unlock()
	log.Printf("Adding projects: %v", projects)
	m.projects = append(m.projects, projects...)
}

func (m *mockStreamingServer) StreamProjectUpdates(_ *proto.EmptyRequest, stream proto.ProjectService_StreamProjectUpdatesServer) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return stream.Send(&proto.ProjectUpdate{
		Projects: m.projects,
	})
}

var _ = Describe("RoleStore", Ordered, func() {
	BeforeAll(func() {
		lis = bufconn.Listen(1024 * 1024)

		// Create and register the mock server
		server = grpc.NewServer()
		mockServer = newMockStreamingServer()
		proto.RegisterProjectServiceServer(server, mockServer)

		go func() {
			defer GinkgoRecover()
			if err := server.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
				log.Printf("Error serving server: %v", err)
			}
		}()

		var err error
		conn, err = grpc.NewClient(
			"passthrough://bufnet",
			grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
				return lis.Dial()
			}),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterAll(func() {
		server.Stop()
		Expect(lis.Close()).To(Succeed())
	})

	BeforeEach(func() {
		// Reset projects for each test
		mockServer.mutex.Lock()
		mockServer.projects = make([]*proto.ProjectEntry, 0)
		mockServer.mutex.Unlock()

		ctx, cancel = context.WithCancel(context.Background())
	})

	AfterEach(func() {
		// Stop the goroutine that fetches project updates
		cancel()
	})

	Context("RoleStore project updates", func() {
		It("should update dynamic roles when project update is received", func() {
			mockServer.addProjects(&proto.ProjectEntry{
				Key: "project1",
				Data: &proto.ProjectData{
					Status:      "Created",
					ProjectName: "project1",
				},
			})

			rs := internal.NewRoleStore(expectedDynamicClaimRole)
			go rs.FetchProjectUpdates(ctx, conn)
			Eventually(func() []string {
				return rs.GetRoles()
			}, "2s", "100ms").Should(
				ContainElement(ContainSubstring("project1")))
		})

		It("should handle multiple project updates", func() {
			mockServer.addProjects(
				&proto.ProjectEntry{
					Key: "project1",
					Data: &proto.ProjectData{
						Status:      "Created",
						ProjectName: "project1",
					},
				},
				&proto.ProjectEntry{
					Key: "project2",
					Data: &proto.ProjectData{
						Status:      "Created",
						ProjectName: "project2",
					},
				},
			)

			rs := internal.NewRoleStore(expectedDynamicClaimRole)
			go rs.FetchProjectUpdates(ctx, conn)

			Eventually(func() []string {
				return rs.GetRoles()
			}, "2s", "100ms").Should(And(
				ContainElement(ContainSubstring("project1")),
				ContainElement(ContainSubstring("project2")),
			))
		})

		It("should handle project deletion", func() {
			mockServer.addProjects(
				&proto.ProjectEntry{
					Key: "project1",
					Data: &proto.ProjectData{
						Status:      "Created",
						ProjectName: "project1",
					},
				},
				&proto.ProjectEntry{
					Key: "project2",
					Data: &proto.ProjectData{
						Status:      "Created",
						ProjectName: "project2",
					},
				},
			)

			rs := internal.NewRoleStore(expectedDynamicClaimRole)
			go rs.FetchProjectUpdates(ctx, conn)

			Eventually(func() []string {
				return rs.GetRoles()
			}, "2s", "100ms").Should(And(
				ContainElement(ContainSubstring("project1")),
				ContainElement(ContainSubstring("project2")),
			))

			mockServer.mutex.Lock()
			originalProject := mockServer.projects[1]
			updatedProject := &proto.ProjectEntry{
				Key: originalProject.Key,
				Data: &proto.ProjectData{
					Status:      "Deleted",
					ProjectName: originalProject.Data.ProjectName,
				},
			}
			mockServer.projects[1] = updatedProject
			mockServer.mutex.Unlock()

			Eventually(func() []string {
				return rs.GetRoles()
			}, "2s", "100ms").Should(And(
				ContainElement(ContainSubstring("project1")),
				Not(ContainElement(ContainSubstring("project2"))),
			))
		})

		It("should exit gracefully when context canceled", func() {
			rs := internal.NewRoleStore(expectedDynamicClaimRole)
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				rs.FetchProjectUpdates(ctx, conn)
			}()

			cancel()
			Eventually(func() bool {
				wg.Wait()
				return true
			}, "2s", "100ms").Should(BeTrue())
		})
	})
})
