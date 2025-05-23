// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0
// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package graphql

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ServerClient is the client API for Server service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ServerClient interface {
	Query(ctx context.Context, in *GraphQLQuery, opts ...grpc.CallOption) (*GraphQLResponse, error)
}

type serverClient struct {
	cc grpc.ClientConnInterface
}

func NewServerClient(cc grpc.ClientConnInterface) ServerClient {
	return &serverClient{cc}
}

func (c *serverClient) Query(ctx context.Context, in *GraphQLQuery, opts ...grpc.CallOption) (*GraphQLResponse, error) {
	out := new(GraphQLResponse)
	err := c.cc.Invoke(ctx, "/graphql.Server/Query", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ServerServer is the server API for Server service.
// All implementations must embed UnimplementedServerServer
// for forward compatibility
type ServerServer interface {
	Query(context.Context, *GraphQLQuery) (*GraphQLResponse, error)
	mustEmbedUnimplementedServerServer()
}

// UnimplementedServerServer must be embedded to have forward compatible implementations.
type UnimplementedServerServer struct {
}

func (UnimplementedServerServer) Query(context.Context, *GraphQLQuery) (*GraphQLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Query not implemented")
}
func (UnimplementedServerServer) mustEmbedUnimplementedServerServer() {}

// UnsafeServerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ServerServer will
// result in compilation errors.
type UnsafeServerServer interface {
	mustEmbedUnimplementedServerServer()
}

func RegisterServerServer(s grpc.ServiceRegistrar, srv ServerServer) {
	s.RegisterService(&Server_ServiceDesc, srv)
}

func _Server_Query_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GraphQLQuery)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServerServer).Query(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/graphql.Server/Query",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServerServer).Query(ctx, req.(*GraphQLQuery))
	}
	return interceptor(ctx, in, info, handler)
}

// Server_ServiceDesc is the grpc.ServiceDesc for Server service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Server_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "graphql.Server",
	HandlerType: (*ServerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Query",
			Handler:    _Server_Query_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/graphql/query.proto",
}
