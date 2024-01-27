package grpc

import (
	"fmt"
	"net"
	"sync"

	"github.com/anhk/mtun/pkg/ipam"
	"github.com/anhk/mtun/pkg/log"
	"github.com/anhk/mtun/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type ServerOption struct {
	Token    string // token to authenticate
	BindPort uint16
}

type Server struct {
	proto.UnimplementedStreamServer

	tun     *TunWrapper // 本地隧道
	rt      *DataStore  // 路由
	streams sync.Map    // 所有Streams集合
	ipam    *ipam.IPAM

	Token string
}

func (server *Server) getToken(stream proto.Stream_PersistentStreamServer) *TokenAuth {
	token := &TokenAuth{}
	if meta, ok := metadata.FromIncomingContext(stream.Context()); !ok {
		return token
	} else {
		return token.ParseMap(meta)
	}
}

func (server *Server) getRemoteAddr(stream proto.Stream_PersistentStreamServer) string {
	if p, ok := peer.FromContext(stream.Context()); ok {
		if tcpAddr, ok := p.Addr.(*net.TCPAddr); ok {
			return tcpAddr.IP.String()
		}
	}
	return ""
}

func (server *Server) Auth(stream proto.Stream_PersistentStreamServer) (string, bool) {
	if _, ok := stream.(*TunWrapper); ok { // TunWrapper，不用进行鉴权
		return "", true
	}
	remote := server.getRemoteAddr(stream)
	log.Info("new stream from %v", remote)
	token := server.getToken(stream)
	if !token.Verify(server.Token) {
		log.Warn("invalid authority from %v", remote)
		_ = stream.Send(&proto.Message{Code: proto.Type_Deny, Data: []byte("错误的认证信息")})
		return remote, false
	}
	return remote, true
}

// GatherRouteTo 将所有路由信息集中发送到Stream
func (server *Server) GatherRouteTo(stream proto.Stream_PersistentStreamServer) {
	server.streams.Range(func(key, value any) bool {
		cidrs, _ := value.([]string)
		for _, cidr := range cidrs {
			log.Debug("广播初始化路由：%v", cidr)
			_ = stream.Send(&proto.Message{Code: proto.Type_AddRoute, Data: []byte(cidr)})
		}
		return true
	})
}

// BroadcastRouteWithoutMe 发送广播，路由信息
func (server *Server) BroadcastRouteWithoutMe(stream proto.Stream_PersistentStreamServer, msg *proto.Message) {
	server.streams.Range(func(key, value any) bool { // TODO: 广播
		log.Debug("key: %p, stream: %p", key, stream)
		if s, ok := key.(proto.Stream_PersistentStreamServer); ok && s != stream {
			log.Debug("s: %p, stream: %p", s, stream)
			_ = s.Send(msg)
		}
		return true
	})
}

// stream退出时清理路由
func (server *Server) cleanupStream(stream proto.Stream_PersistentStreamServer) {
	log.Debug("stream 退出")
	if cidrs, ok := server.streams.LoadAndDelete(stream); ok {
		cidrs := cidrs.([]string)
		server.rt.DeleteBatch(cidrs)
		for _, cidr := range cidrs {
			log.Debug("广播删除路由：%v", cidr)
			server.BroadcastRouteWithoutMe(stream, &proto.Message{Code: proto.Type_DelRoute, Data: []byte(cidr)})
		}
	}
}

func (server *Server) PersistentStream(stream proto.Stream_PersistentStreamServer) error {
	remote, ok := server.Auth(stream)
	if !ok {
		return fmt.Errorf("错误的认证信息")
	}
	log.Debug("Stream [%v] online %p", remote, stream)

	// 发路由
	server.GatherRouteTo(stream)
	server.streams.Store(stream, []string{}) // 将本Stream加入到集合,val为路由

	log.Debug("将本Stream加入到集合 as key, stream: %p", stream)
	addr, err := server.ipam.Alloc()
	if err != nil {
		log.Error("unavailable ipam alloc address: %v", err)
		return err
	}
	server.rt.Add(fmt.Sprintf("%v/32", addr), stream) // 直连路由

	defer func() {
		server.cleanupStream(stream)
		server.ipam.Release(addr)
	}()

	log.Debug("stream: %p", stream)
	stream.Send(&proto.Message{
		Code:    proto.Type_Assign,
		Data:    []byte(fmt.Sprintf("%v/%v", addr.String(), server.ipam.Mask())),
		Gateway: server.ipam.Gateway().String(),
	})

	log.Debug("分配IP地址: %v/%v@%v", addr.String(), server.ipam.Mask(), server.ipam.Gateway().String())

	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Info("client %v exit", remote)
			return err
		}

		log.Debug("收到消息: type: %d", msg.Code)
		switch msg.Code {
		case proto.Type_AddRoute:
			log.Debug("这是添加路由的消息")
			cidr := string(msg.Data)
			cidrs, _ := server.streams.Load(stream)
			server.streams.Store(stream, append(cidrs.([]string), cidr))
			_ = server.rt.Add(cidr, stream)
			server.BroadcastRouteWithoutMe(stream, msg)
		case proto.Type_Data:
			// TODO: 解包
			log.Debug("这里数据包的Payload")
			hdr, err := ParseIpv4Hdr(msg.Data)
			if err != nil { // invalid packet
				log.Debug("解IPv4头失败: %v", err)
				continue
			}

			log.Debug("packet from %v to %v", hdr.Src.String(), hdr.Dst.String())

			// TODO: 路由
			if st, ok := server.rt.Lookup(hdr.Dst); ok {
				st.Send(msg)
			}

			//server.rt.Lookup()
			//_ = server.tun.Send(msg)
		}
	}
}

func StartGrpcServer(option *ServerOption, subnet string) *Server {
	bindAddr := fmt.Sprintf(":%d", option.BindPort)
	log.Info("bind grpc on %v", bindAddr)

	listen, err := net.Listen("tcp", bindAddr)
	check(err)

	server := &Server{
		Token: option.Token,
		tun:   NewTunWrapper(),
		rt:    NewDataStore(),
	}
	server.ipam, err = ipam.NewIPAM(subnet)
	check(err)

	grpcSvc := grpc.NewServer()
	proto.RegisterStreamServer(grpcSvc, server)
	go func() { check(grpcSvc.Serve(listen)) }()
	go func() { _ = server.PersistentStream(server.tun) }() // 将隧道封装成GRPCStream
	return server
}

// ReadMessage 从GRPCServer读取发送给Tunnel的数据
func (server *Server) ReadMessage() (*proto.Message, error) {
	return <-server.tun.tx, nil
}

// WriteMessage 本地隧道向GRPCServer写数据
func (server *Server) WriteMessage(msg *proto.Message) error {
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
		rx: make(chan *proto.Message, 2048),
		tx: make(chan *proto.Message, 2048),
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
