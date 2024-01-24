package grpc

import (
	"github.com/anhk/mtun/proto"
)

type ServerOption struct {
	Token    string // token to authenticate
	BindAddr string
	BindPort uint16
}

type GrpcServer struct {
	proto.UnimplementedStreamServer
}

func StartGrpcServer(option *ServerOption) {

}
