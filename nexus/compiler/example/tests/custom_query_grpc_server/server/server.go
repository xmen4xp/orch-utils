// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/generated/graphql"
	"google.golang.org/grpc"
)

type CustomQueryService struct{}

func (c *CustomQueryService) Query(_ context.Context, q *graphql.GraphQLQuery) (*graphql.GraphQLResponse, error) {
	log.Printf("Query name is %s, hierarchy is %v, user provided args are %v",
		q.Query, q.Hierarchy, q.UserProvidedArgs)

	return &graphql.GraphQLResponse{
		Code:    200,
		Message: "Hi Kacper",
		Data: map[string]string{
			"my hierarachy is": fmt.Sprintf("%v", q.Hierarchy),
			"my args are":      fmt.Sprintf("%v", q.UserProvidedArgs),
		},
	}, nil
}

func StartQueryGRPC(port int) {
	queryService := graphql.NewServerService(&CustomQueryService{})

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()

	graphql.RegisterServerService(grpcServer, queryService)

	err = grpcServer.Serve(lis)
	if err != nil {
		panic(err)
	}
}
