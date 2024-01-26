package tun

import (
	"github.com/anhk/mtun/pkg/log"
	"io"
	"net"
	"os"
)

type Interface interface {
	Read() ([]byte, error)
	Write([]byte) error
}

// Wrapper MacOS有4字节的头需要处理一下
type Wrapper struct {
	io.ReadWriter
}

type Tun struct {
	fp   Interface
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
	log.Info("add route %v to %v", cidr, tun.Name)
	if _, _, err := net.ParseCIDR(cidr); err != nil {
		return err
	}
	return tun.addRoute(cidr)
}

func (tun *Tun) DelRoute(cidr string) error {
	log.Info("del route %v to %v", cidr, tun.Name)
	if _, _, err := net.ParseCIDR(cidr); err != nil {
		return err
	}
	return tun.delRoute(cidr)
}

func (tun *Tun) Read() ([]byte, error) {
	return tun.fp.Read()
}

func (tun *Tun) Write(p []byte) (err error) {
	return tun.fp.Write(p)
}
