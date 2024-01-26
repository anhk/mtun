package grpc

import (
	"github.com/anhk/mtun/pkg/log"
	"github.com/anhk/mtun/proto"
	"os"
)

type Socket interface {
	ReadMessage() (*proto.Message, error)
	WriteMessage(message *proto.Message) error
}

func check(e any) {
	if e != nil {
		log.Error("fatal error: %v", e)
		os.Exit(1)
	}
}
