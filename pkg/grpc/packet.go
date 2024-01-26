package grpc

import (
	"fmt"
	"net"
)

type IPv4Hdr struct {
	Src net.IP // source address
	Dst net.IP // destination address
}

func ParseIpv4Hdr(data []byte) (*IPv4Hdr, error) {
	var hdr IPv4Hdr
	if len(data) <= 20 {
		return nil, fmt.Errorf("data length [%d] is too short", len(data))
	}
	if len(data) < int(data[0]&0x0f)<<2 {
		return nil, fmt.Errorf("data length [%d] is too short", len(data))
	}
	hdr.Src = net.IPv4(data[12], data[13], data[14], data[15])
	hdr.Dst = net.IPv4(data[16], data[17], data[18], data[19])
	return &hdr, nil
}
