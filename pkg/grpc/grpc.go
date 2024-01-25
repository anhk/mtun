package grpc

import (
	"runtime/debug"

	"github.com/anhk/mtun/pkg/log"
)

type Socket interface {
}

func check(e any) {
	if e != nil {
		log.Error("%s", debug.Stack())
		panic(e)
	}
}
