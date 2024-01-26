package grpc

import (
	"fmt"
	"github.com/anhk/mtun/pkg/log"
	"google.golang.org/grpc/peer"
	"net"
	"sync"

	"github.com/anhk/mtun/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ServerOption struct {
	Token    string // token to authenticate
	BindPort uint16
}

type GrpcServer struct {
	proto.UnimplementedStreamServer

	tun     *TunWrapper // 本地隧道
	rt      *DataStore  // 路由
	streams sync.Map    // 所有Streams集合

	Token string
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
	var remote string
	if _, ok := stream.(*TunWrapper); !ok { // 不是TunWrapper，进行鉴权
		remote = server.getRemoteAddr(stream)
		log.Info("new stream from %v", remote)
		token := server.getToken(stream)
		if !token.Verify(server.Token) {
			log.Warn("invalid authority from %v", remote)
			_ = stream.Send(&proto.Message{Code: proto.Type_Deny, Data: []byte("错误的认证信息")})
			return fmt.Errorf("错误的认证信息")
		}
	}

	log.Debug("Stream online %p", stream)

	// 发路由
	server.streams.Range(func(key, value any) bool {
		cidrs, _ := value.([]string)
		for _, cidr := range cidrs {
			log.Debug("广播初始化路由：%v", cidr)
			stream.Send(&proto.Message{Code: proto.Type_AddRoute, Data: []byte(cidr)})
		}
		return true
	})

	log.Debug("将本Stream加入到集合 as key, stream: %p", stream)
	server.streams.Store(stream, []string{}) // 将本Stream加入到集合,val为路由
	defer func() {
		log.Debug("stream 退出")
		if cidrs, ok := server.streams.LoadAndDelete(stream); ok {
			server.rt.DeleteBatch(cidrs.([]string))
		}
	}()

	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Info("client %v exit", remote)
			return err
		}

		log.Debug("收到消息：type: %d", msg.Code)
		switch msg.Code {
		case proto.Type_AddRoute:
			log.Debug("这是添加路由的消息")
			cidr := string(msg.Data)
			cidrs, _ := server.streams.Load(stream)
			server.streams.Store(stream, append(cidrs.([]string), cidr))
			server.rt.Add(cidr, stream)
			server.streams.Range(func(key, value any) bool { // TODO: 广播
				log.Debug("key: %p, stream: %p", key, stream)
				if s, ok := key.(proto.Stream_PersistentStreamServer); ok && s != stream {
					log.Debug("s: %p, stream: %p", s, stream)
					s.Send(msg)
				}
				return true
			})
		case proto.Type_Data:
			// TODO: 路由
			// TODO: 解包
			//server.rt.Lookup()
			server.tun.Send(msg)
		}
	}
}

func StartGrpcServer(option *ServerOption) *GrpcServer {
	bindAddr := fmt.Sprintf(":%d", option.BindPort)
	log.Info("bind grpc on %v", bindAddr)

	listen, err := net.Listen("tcp", bindAddr)
	check(err)

	server := &GrpcServer{
		Token: option.Token,
		tun:   NewTunWrapper(),
		rt:    NewDataStore(),
	}
	grpcSvc := grpc.NewServer()
	proto.RegisterStreamServer(grpcSvc, server)
	go func() { check(grpcSvc.Serve(listen)) }()
	go func() { server.PersistentStream(server.tun) }() // 将隧道封装成GRPCStream
	return server
}

// ReadMessage 从GRPCServer读取发送给Tunnel的数据
func (server *GrpcServer) ReadMessage() (*proto.Message, error) {
	return <-server.tun.tx, nil
}

// WriteMessage 本地隧道向GRPCServer写数据
func (server *GrpcServer) WriteMessage(msg *proto.Message) error {
	server.tun.rx <- msg
	return nil
}

// TunWrapper 将Tunnel模拟GRPC Stream
type TunWrapper struct {
	tx, rx chan *proto.Message
	grpc.ServerStream
}

func NewTunWrapper() *TunWrapper {
	return &TunWrapper{
		rx: make(chan *proto.Message),
		tx: make(chan *proto.Message),
	}
}

// Send 向隧道发送数据
func (tun *TunWrapper) Send(msg *proto.Message) error {
	tun.tx <- msg
	return nil
}

// Recv 从隧道读数据，由外部APP调用
func (tun *TunWrapper) Recv() (*proto.Message, error) {
	return <-tun.rx, nil
}
