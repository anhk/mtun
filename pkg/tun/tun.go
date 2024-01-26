package tun

import (
	"github.com/anhk/mtun/pkg/log"
	"io"
	"net"
	"os"
)

type Tun struct {
	fp   io.ReadWriter
	Name string
}

func check(e any) {
	if e != nil {
		log.Error("fatal error: %v", e)
		os.Exit(1)
	}
}

func AllocTun() *Tun {
	return allocTun()
}

func (tun *Tun) AddRoute(cidr string) error {
	if _, _, err := net.ParseCIDR(cidr); err != nil {
		return err
	}
	return tun.addRoute(cidr)
}

func (tun *Tun) DelRoute(cidr string) error {
	if _, _, err := net.ParseCIDR(cidr); err != nil {
		return err
	}
	return tun.delRoute(cidr)
}

func (tun *Tun) Read(p []byte) (n int, err error) {
	return tun.fp.Read(p)
}

func (tun *Tun) Write(p []byte) (n int, err error) {
	return tun.fp.Write(p)
}
