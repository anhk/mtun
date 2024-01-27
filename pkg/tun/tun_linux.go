package tun

import (
	"os"
	"os/exec"

	"github.com/anhk/mtun/pkg/log"
	"golang.org/x/sys/unix"
)

func allocTun() *Tun {
	nfd, err := unix.Open("/dev/net/tun", unix.O_RDWR|unix.O_CLOEXEC, 0)
	check(err)

	ifr, err := unix.NewIfreq("")
	check(err)

	ifr.SetUint16(unix.IFF_TUN | unix.IFF_NO_PI)
	check(unix.IoctlIfreq(nfd, unix.TUNSETIFF, ifr))

	// check(unix.SetNonblock(nfd, true))
	return (&Tun{
		fp:   &Wrapper{os.NewFile(uintptr(nfd), "/dev/net/tun")},
		Name: ifr.Name(),
	}).up()
}

func (tun *Tun) up() *Tun {
	exec.Command("ip", "link", "set", tun.Name, "up").Run()
	return tun
}

func (tun *Tun) addRoute(cidr string) error {
	log.Debug("%v %v %v %v %v %v", "ip", "route", "add", cidr, "dev", tun.Name)
	return exec.Command("ip", "route", "add", cidr, "dev", tun.Name).Run()
}

func (tun *Tun) delRoute(cidr string) error {
	log.Debug("%v %v %v %v %v %v", "ip", "route", "del", cidr, "dev", tun.Name)
	return exec.Command("ip", "route", "del", cidr, "dev", tun.Name).Run()
}

func (tun *Tun) setAddress(addr, _ string) error {
	log.Debug("%v %v %v %v %v %v", "ip", "addr", "add", addr, "dev", tun.Name)
	return exec.Command("ip", "addr", "add", addr, "dev", tun.Name).Run()
}

func (tun *Wrapper) Read() ([]byte, error) {
	var data = make([]byte, 4096)
	datalen, err := tun.ReadWriter.Read(data)
	if err != nil {
		return nil, err
	}
	log.Debug("datalen: %d, [%x %x %x %x %x %x]", datalen, data[0], data[1], data[2], data[3], data[4], data[5])
	return data[:datalen], nil
}

func (tun *Wrapper) Write(data []byte) error {
	_, err := tun.ReadWriter.Write(data)
	return err
}
