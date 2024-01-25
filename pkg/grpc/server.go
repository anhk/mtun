package grpc

import (
	"fmt"
	"net"

	"github.com/anhk/mtun/proto"
	"google.golang.org/grpc"
)

type ServerOption struct {
	Token    string // token to authenticate
	BindPort uint16
}

type GrpcServer struct {
	proto.UnimplementedStreamServer
}

func StartGrpcServer(option *ServerOption) *GrpcServer {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", option.BindPort))
	check(err)

	server := &GrpcServer{}
	grpcSvc := grpc.NewServer()
	proto.RegisterStreamServer(grpcSvc, server)
	check(grpcSvc.Serve(listen))
	return server
}
