package grpc

import (
	"fmt"
	"net"

	"github.com/anhk/mtun/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ServerOption struct {
	Token    string // token to authenticate
	BindPort uint16
}

type GrpcServer struct {
	Token string
	proto.UnimplementedStreamServer
}

func (svc *GrpcServer) getToken(stream proto.Stream_PersistentStreamServer) *TokenAuth {
	token := &TokenAuth{}
	if meta, ok := metadata.FromIncomingContext(stream.Context()); !ok {
		return token
	} else {
		return token.ParseMap(meta)
	}
}
func (server *GrpcServer) PersistentStream(stream proto.Stream_PersistentStreamServer) error {
	token := server.getToken(stream)
	if !token.Verify(server.Token) {
		_ = stream.Send(&proto.Message{Code: proto.Type_Deny, Data: []byte("错误的认证信息")})
		return fmt.Errorf("错误的认证信息")
	}
	return nil
}

func StartGrpcServer(option *ServerOption) *GrpcServer {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", option.BindPort))
	check(err)

	server := &GrpcServer{Token: option.Token}
	grpcSvc := grpc.NewServer()
	proto.RegisterStreamServer(grpcSvc, server)
	check(grpcSvc.Serve(listen))
	return server
}
