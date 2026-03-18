// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package echoserver_test

import (
	"context"
	"fmt"
	"net"

	"nexus-api-gw/pkg/client"
	"nexus-api-gw/pkg/config"
	"nexus-api-gw/pkg/model"
	"nexus-api-gw/pkg/server/echoserver"

	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	k8sFake "k8s.io/client-go/kubernetes/fake"
)

func setupServer() (*echoserver.EchoServer, chan struct{}) {
	model.URIToCRDType = URIToCRDType
	model.CrdTypeToNodeInfo = CrdTypeToNodeInfo

	stopCh := make(chan struct{})

	// Setup dynamic port allocation
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	// Get the dynamically allocated port
	addr := listener.Addr()
	if addr == nil {
		panic("failed to get listener address")
	}
	tcpAddr, ok := addr.(*net.TCPAddr)
	if !ok {
		panic("failed to cast listener address to TCPAddr")
	}
	port := tcpAddr.Port

	// Set the server configuration to use the dynamic port
	config.Cfg = &config.Config{
		Server: config.ServerConfig{
			HTTPPort: fmt.Sprintf("%d", port),
		},
	}
	k8sClient := k8sFake.NewSimpleClientset()
	client.Client = fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), GRVToListKind)

	serverObj := echoserver.InitEcho(stopCh, config.Cfg, k8sClient, nil)

	return serverObj, stopCh
}

func teardownServer(serverObj *echoserver.EchoServer, stopCh chan struct{}) {
	err := serverObj.Echo.Shutdown(context.Background())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	close(stopCh)
}
