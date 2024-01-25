package tun

import (
	"net"
	"os"
	"runtime/debug"

	"github.com/anhk/mtun/pkg/log"
)

type TunOption struct {
	Cidr []string // cidr to claim
}

type Tun struct {
	fp   *os.File
	Name string
}

func check(e any) {
	if e != nil {
		log.Error("%s", debug.Stack())
		panic(e)
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

func (tun *Tun) Read() {
}

func (tun *Tun) Write() {
}
