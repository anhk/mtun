package grpc

import (
	"errors"
	"fmt"
	"github.com/anhk/mtun/pkg/log"
	"google.golang.org/grpc/peer"
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

func (server *GrpcServer) getToken(stream proto.Stream_PersistentStreamServer) *TokenAuth {
	token := &TokenAuth{}
	if meta, ok := metadata.FromIncomingContext(stream.Context()); !ok {
		return token
	} else {
		return token.ParseMap(meta)
	}
}

func (server *GrpcServer) getRemoteAddr(stream proto.Stream_PersistentStreamServer) string {
	if p, ok := peer.FromContext(stream.Context()); ok {
		if tcpAddr, ok := p.Addr.(*net.TCPAddr); ok {
			return tcpAddr.IP.String()
		}
	}
	return ""
}

func (server *GrpcServer) PersistentStream(stream proto.Stream_PersistentStreamServer) error {
	remote := server.getRemoteAddr(stream)
	log.Info("new stream from %v", remote)
	token := server.getToken(stream)
	if !token.Verify(server.Token) {
		log.Warn("invalid authority from %v", remote)
		_ = stream.Send(&proto.Message{Code: proto.Type_Deny, Data: []byte("错误的认证信息")})
		return fmt.Errorf("错误的认证信息")
	}

	select {}
	return nil
}

func StartGrpcServer(option *ServerOption) *GrpcServer {
	bindAddr := fmt.Sprintf(":%d", option.BindPort)
	log.Info("bind grpc on %v", bindAddr)

	listen, err := net.Listen("tcp", bindAddr)
	check(err)

	server := &GrpcServer{Token: option.Token}
	grpcSvc := grpc.NewServer()
	proto.RegisterStreamServer(grpcSvc, server)
	go func() { check(grpcSvc.Serve(listen)) }()
	return server
}

func (server *GrpcServer) ReadMessage() (*proto.Message, error) {
	//return nil, errors.New("not implement")
	select {}
}

func (server *GrpcServer) WriteMessage(message *proto.Message) error {
	return errors.New("not implement")
}
