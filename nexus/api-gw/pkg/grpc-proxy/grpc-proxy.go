package grpcproxy

import (
	"context"
	"fmt"
	"net"
	"strings"

	nexus_client "nexus/admin/api/build/nexus-client"

	"github.com/mwitkow/grpc-proxy/proxy"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type grpcProxy struct {
	nexusClient      *nexus_client.Clientset
	grpcMethodPrefix map[string]string
	server           *grpc.Server
}

func (p *grpcProxy) stop() {
	if p.server != nil {
		p.server.Stop()
		p.server = nil
	}
}

func (p *grpcProxy) wait(stopCh chan struct{}) {
	select {
	case <-stopCh:
		p.stop()
	}
}

var gProxy *grpcProxy

func GRPCProxyInit(nexusClient *nexus_client.Clientset, grpcAddr string, stopCh chan struct{}) {

	if gProxy == nil {
		gProxy = &grpcProxy{
			nexusClient:      nexusClient,
			grpcMethodPrefix: map[string]string{},
		}
	}

	director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		outCtx := metadata.NewOutgoingContext(ctx, md.Copy())

		var server string
		var found bool
		for key, value := range gProxy.grpcMethodPrefix {
			log.Debugf("director:  Contains= %v", strings.Contains(fullMethodName, key))
			if strings.Contains(fullMethodName, key) {
				server = value
				found = true
			}
		}

		if !found {
			return outCtx, nil, status.Errorf(codes.Unimplemented, "Unknown method")
		}

		conn, err := grpc.DialContext(ctx, server, grpc.WithCodec(proxy.Codec()), grpc.WithTransportCredentials(insecure.NewCredentials()))

		return outCtx, conn, err
	}

	gProxy.nexusClient.ApiNexus().Subscribe()
	gProxy.nexusClient.ApiNexus().Config().Subscribe()
	gProxy.nexusClient.ApiNexus().Config().Servers("*").Subscribe()
	gProxy.nexusClient.ApiNexus().Config().Servers("*").RegisterAddCallback(
		func(obj *nexus_client.RouteServer) {
			log.Infof("add callback called for server object: %+v\n", *obj)
			if obj.Spec.Service.Scheme == "grpc" {
				gProxy.grpcMethodPrefix[obj.Spec.Uri] = fmt.Sprintf("%s:%d", obj.Spec.Service.Name, obj.Spec.Service.Port)
			}
		})
	gProxy.nexusClient.ApiNexus().Config().Servers("*").RegisterDeleteCallback(
		func(obj *nexus_client.RouteServer) {
			log.Infof("delete callback called for server object: %+v\n", *obj)
			delete(gProxy.grpcMethodPrefix, obj.Spec.Uri)
		})

	gProxy.server = grpc.NewServer(
		grpc.UnknownServiceHandler(proxy.TransparentHandler(director)))

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Infof("grpc proxy server listening at %v", lis.Addr())
	if err := gProxy.server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
