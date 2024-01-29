package tun

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

const (
	appleUTUNCtl     = "com.apple.net.utun_control"
	appleCTLIOCGINFO = (0x40000000 | 0x80000000) | ((100 & 0x1fff) << 16) | uint32(byte('N'))<<8 | 3
)

/*
 * struct sockaddr_ctl {
 *     u_char sc_len; // depends on size of bundle ID string
 *     u_char sc_family; // AF_SYSTEM
 *     u_int16_t ss_sysaddr; // AF_SYS_KERNCONTROL
 *     u_int32_t sc_id; // Controller unique identifier
 *     u_int32_t sc_unit; // Developer private unit number
 *     u_int32_t sc_reserved[5];
 * };
 */
type sockaddrCtl struct {
	scLen      uint8
	scFamily   uint8
	ssSysaddr  uint16
	scID       uint32
	scUnit     uint32
	scReserved [5]uint32
}

var sockaddrCtlSize uintptr = 32

func allocTun() *Tun {
	nfd, err := syscall.Socket(syscall.AF_SYSTEM, syscall.SOCK_DGRAM, 2)
	check(err)

	var ctlInfo = &struct {
		ctlID   uint32
		ctlName [96]byte
	}{}
	copy(ctlInfo.ctlName[:], []byte(appleUTUNCtl))

	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(nfd), uintptr(appleCTLIOCGINFO), uintptr(unsafe.Pointer(ctlInfo))); errno != 0 {
		check(fmt.Errorf("error in syscall.Syscall(syscall.SYS_IOCTL, ...): %v", errno))
	}

	addrP := unsafe.Pointer(&sockaddrCtl{
		scLen:     uint8(sockaddrCtlSize),
		scFamily:  syscall.AF_SYSTEM,
		ssSysaddr: 2, /* #define AF_SYS_CONTROL 2 */
		scID:      ctlInfo.ctlID,
		scUnit:    0,
	})

	if _, _, errno := syscall.RawSyscall(syscall.SYS_CONNECT, uintptr(nfd), uintptr(addrP), uintptr(sockaddrCtlSize)); errno != 0 {
		check(fmt.Errorf("error in syscall.RawSyscall(syscall.SYS_CONNECT, ...): %v", errno))
	}

	var ifName struct {
		name [16]byte
	}
	ifNameSize := uintptr(16)

	if _, _, errno := syscall.Syscall6(syscall.SYS_GETSOCKOPT, uintptr(nfd),
		2, /* #define SYSPROTO_CONTROL 2 */
		2, /* #define UTUN_OPT_IFNAME 2 */
		uintptr(unsafe.Pointer(&ifName)),
		uintptr(unsafe.Pointer(&ifNameSize)), 0); errno != 0 {
		check(fmt.Errorf("error in syscall.Syscall6(syscall.SYS_GETSOCKOPT, ...): %v", errno))
	}

	return (&Tun{
		fp:   &Wrapper{os.NewFile(uintptr(nfd), "/dev/net/tun")},
		Name: string(ifName.name[:ifNameSize-1]),
	}).up()
}

func (tun *Tun) up() *Tun {
	_ = exec.Command("ip", "link", "set", tun.Name, "up").Run()
	return tun
}

func (tun *Tun) addRoute(cidr string) error {
	return exec.Command("ip", "route", "add", cidr, "dev", tun.Name).Run()
}

func (tun *Tun) delRoute(cidr string) error {
	return exec.Command("ip", "route", "del", cidr, "dev", tun.Name).Run()
}

func (tun *Tun) setAddress(addr, remote string) error {

	// 设置IP地址
	if err := exec.Command("ifconfig", tun.Name, addr, remote, "up").Run(); err != nil {
		return err
	}
	// 设置直连路由
	return tun.addRoute(addr)
}

func (tun *Tun) setSNAT(_ *net.IPNet) error {
	//exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING", "-s", ipNet.String(), "-j", "MASQUERADE").Run()
	//return exec.Command("iptables", "-t", "nat", "-I", "POSTROUTING", "-s", ipNet.String(), "-j", "MASQUERADE").Run()
	return nil
}

func (tun *Wrapper) Read() ([]byte, error) {
	var data = make([]byte, 4096)
	datalen, err := tun.ReadWriter.Read(data)
	if err != nil {
		return nil, err
	}
	if datalen < 4 {
		return nil, fmt.Errorf("data is too short")
	}
	return data[4:datalen], nil
}

func (tun *Wrapper) Write(data []byte) error {
	var hdr = [4]byte{0, 0, 0, 2}
	if data[0]&0xf == 6 {
		hdr[3] = 10
	}
	_, err := tun.ReadWriter.Write(append(hdr[:], data...))
	return err
}
