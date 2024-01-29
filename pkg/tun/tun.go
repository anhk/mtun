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

func (tun *Tun) SetAddress(addr, gateway string) error {
	if _, ipNet, err := net.ParseCIDR(addr); err != nil {
		log.Error("parse addr failed: %v", addr)
		return err
	} else if err := tun.setAddress(addr, gateway); err != nil {
		return err
	} else if err := tun.setSNAT(ipNet); err != nil {
		return err
	}
	return nil
}

func (tun *Tun) AddRoute(cidr string) error {
	log.Info("add route %v to %v", cidr, tun.Name)
	if _, _, err := net.ParseCIDR(cidr); err != nil {
		log.Error("invalid route: %v", cidr)
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

func (tun *Tun) setSNAT(ipNet *net.IPNet) error {
	//exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING", "-s", ipNet.String(), "-j", "MASQUERADE").Run()
	//return exec.Command("iptables", "-t", "nat", "-I", "POSTROUTING", "-s", ipNet.String(), "-j", "MASQUERADE").Run()
	return nil
}

func (tun *Tun) Read() ([]byte, error) {
	return tun.fp.Read()
}

func (tun *Tun) Write(p []byte) (err error) {
	return tun.fp.Write(p)
}
