package tun

import (
	"fmt"
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
		fp:   os.NewFile(uintptr(nfd), "/dev/net/tun"),
		Name: string(ifName.name[:ifNameSize-1]),
	}).setAddress("22.22.22.252/31", "22.22.22.253")
}

func (tun *Tun) addRoute(cidr string) error {
	return exec.Command("ip", "route", "add", cidr, "dev", tun.Name).Run()
}

func (tun *Tun) delRoute(cidr string) error {
	return exec.Command("ip", "route", "del", cidr, "dev", tun.Name).Run()
}

func (tun *Tun) setAddress(addr, remote string) *Tun {
	check(exec.Command("ifconfig", tun.Name, addr, remote, "up").Run())
	return tun
}