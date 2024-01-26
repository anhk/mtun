package tun

import (
	"os"
	"os/exec"

	"golang.org/x/sys/unix"
)

func allocTun() *Tun {
	nfd, err := unix.Open("/dev/net/tun", unix.O_RDWR|unix.O_CLOEXEC, 0)
	check(err)

	ifr, err := unix.NewIfreq("")
	check(err)

	ifr.SetUint16(unix.IFF_TUN | unix.IFF_NO_PI | unix.IFF_VNET_HDR)
	check(unix.IoctlIfreq(nfd, unix.TUNSETIFF, ifr))

	// check(unix.SetNonblock(nfd, true))
	return (&Tun{
		fp:   &Wrapper{os.NewFile(uintptr(nfd), "/dev/net/tun")},
		Name: ifr.Name(),
	}).setAddress("22.22.22.252/31")
}

func (tun *Tun) addRoute(cidr string) error {
	return exec.Command("ip", "route", "add", cidr, "dev", tun.Name).Run()
}

func (tun *Tun) delRoute(cidr string) error {
	return exec.Command("ip", "route", "del", cidr, "dev", tun.Name).Run()
}

func (tun *Tun) setAddress(addr string) *Tun {
	check(exec.Command("ip", "addr", "add", addr, "dev", tun.Name).Run())
	return tun
}

func (tun *Wrapper) Read() ([]byte, error) {
	var data = make([]byte, 4096)
	datalen, err := tun.ReadWriter.Read(data)
	if err != nil {
		return nil, err
	}
	return data[:datalen], nil
}

func (tun *Wrapper) Write(data []byte) error {
	_, err := tun.ReadWriter.Write(data)
	return err
}
