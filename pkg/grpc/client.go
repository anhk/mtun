package grpc

type ClientOption struct {
	Token      string // token to authenticate
	ServerAddr string // the address of server
	ServerPort uint16 // the port of server
}

type GrpcClient struct {
}

func StartGrpcClient(option *ClientOption) {

}
