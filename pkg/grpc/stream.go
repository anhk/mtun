package grpc

import "sync"

type StreamStore struct {
	sync.Map
}
