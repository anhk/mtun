package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/anhk/mtun/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientOption struct {
	Token      string // token to authenticate
	ServerAddr string // the address of server
	ServerPort uint16 // the port of server
}

type Client struct {
	c      grpc.ClientConnInterface
	client proto.StreamClient
	stream proto.Stream_PersistentStreamClient
}

func StartGrpcClient(option *ClientOption) *Client {
	token := &TokenAuth{}
	token.T = time.Now().Unix()
	token.S = token.Sign(option.Token)

	c, err := grpc.Dial(fmt.Sprintf("%s:%d", option.ServerAddr, option.ServerPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(token))
	check(err)

	client := proto.NewStreamClient(c)
	stream, err := client.PersistentStream(context.Background())
	check(err)

	return &Client{c: c, client: client, stream: stream}
}

func (client *Client) ReadMessage() (*proto.Message, error) {
	return client.stream.Recv()
}

func (client *Client) WriteMessage(message *proto.Message) error {
	return client.stream.Send(message)
}
