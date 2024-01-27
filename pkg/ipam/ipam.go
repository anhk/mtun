package ipam

import (
	"errors"
	"math/big"
	"net"

	"github.com/RoaringBitmap/roaring"
	"github.com/anhk/mtun/pkg/log"
)

type IPAM struct {
	m *roaring.Bitmap

	base *big.Int
	max  uint32
	mask int
}

func NewIPAM(cidr string) (*IPAM, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	ones, bits := ipNet.Mask.Size()
	ipam := &IPAM{m: roaring.NewBitmap()}
	ipam.base = big.NewInt(0).SetBytes(ipNet.IP)
	ipam.max = 1 << (bits - ones)
	ipam.mask = ones
	return ipam, nil
}

func (ipam *IPAM) Alloc() (net.IP, error) {
	for i := uint32(2); i < ipam.max; i++ {
		if ok := ipam.m.CheckedAdd(i); ok {
			addr := net.IP(big.NewInt(0).Add(ipam.base, big.NewInt(int64(i))).Bytes())
			log.Info("分配IP地址: %v", addr.String())
			return addr, nil
		}
	}

	return nil, errors.New("unavailable")
}

func (ipam *IPAM) Release(addr net.IP) {
	log.Info("释放IP地址: %v", addr.String())
	offset := big.NewInt(0).Sub(big.NewInt(0).SetBytes(addr), ipam.base).Int64()
	ipam.m.CheckedRemove(uint32(offset))
}

func (ipam *IPAM) Mask() int {
	return ipam.mask
}

func (ipam *IPAM) Gateway() net.IP {
	return net.IP(big.NewInt(0).Add(ipam.base, big.NewInt(int64(1))).Bytes())
}
