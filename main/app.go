package main

import (
	"github.com/anhk/mtun/pkg/grpc"
	"github.com/anhk/mtun/pkg/log"
	"github.com/anhk/mtun/pkg/tun"
	"github.com/anhk/mtun/proto"
)

type App struct {
	tun    *tun.Tun
	sock   grpc.Socket
	Cidrs  []string
	Subnet string // 22.22.22.0/24
}

func (app *App) run() {
	stopCh := make(chan struct{})
	go app.processSocket(stopCh)
	go app.processTunnel(stopCh)
	<-stopCh
}

func (app *App) RunAsClient(clientOpt *grpc.ClientOption) {
	app.sock = grpc.StartGrpcClient(clientOpt)
	log.Info("connect ok")
	app.run()
}

func (app *App) RunAsServer(serverOpt *grpc.ServerOption) {
	app.sock = grpc.StartGrpcServer(serverOpt, app.Subnet)
	log.Info("start grpc server ok")
	app.run()
}

func (app *App) StartTunnel() *App {
	app.tun = tun.AllocTun()
	return app
}

func (app *App) processSocket(stopCh chan struct{}) {
	// Step 1: 发送路由
	for _, cdir := range app.Cidrs {
		log.Debug("发送路由%v到GRPC", cdir)
		if err := app.sock.WriteMessage(&proto.Message{Code: proto.Type_AddRoute, Data: []byte(cdir)}); err != nil {
			log.Error("write route to grpc failed: %v", err)
			stopCh <- struct{}{}
			return
		}
	}

	// Stop 2: 处理GRPC数据
loop:
	for {
		msg, err := app.sock.ReadMessage()
		if err != nil {
			log.Error("read from grpc failed: %v", err)
			break
		}

		log.Debug("receive %d", msg.Code)

		switch msg.Code {
		case proto.Type_Deny:
			log.Info("authority denied")
			break loop
		case proto.Type_AddRoute:
			app.tun.AddRoute(string(msg.Data))
		case proto.Type_DelRoute:
			app.tun.DelRoute(string(msg.Data))
		case proto.Type_Assign:
			log.Debug("收到分配地址的数据包")
			app.tun.SetAddress(string(msg.Data), msg.Gateway)
		case proto.Type_Data:
			app.tun.Write(msg.Data)
		}
	}

	log.Info("process grpc goroutine exit")
	stopCh <- struct{}{}
}

func (app *App) processTunnel(stopCh chan struct{}) {
	//var buf = make([]byte, 2048)

	for {
		if buf, err := app.tun.Read(); err != nil {
			log.Error("read from tun failed: %v", err)
			break
		} else if buf[0]&0xF0 != 0x40 { // only ipv4
			log.Debug("not ipv4")
			continue
		} else if err := app.sock.WriteMessage(&proto.Message{Code: proto.Type_Data, Data: buf}); err != nil {
			log.Error("write to grpc failed: %v", err)
			break
		}
	}
	log.Info("process tunnel goroutine exit")
	stopCh <- struct{}{}
}
