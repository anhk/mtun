package grpc

import (
	"net"
	"net/netip"

	"github.com/anhk/mtun/proto"
	"github.com/gaissmai/cidrtree"
)

type RouteTable struct {
	rt *cidrtree.Table[proto.Stream_PersistentStreamServer]
}

func NewRouteTable() *RouteTable {
	return &RouteTable{rt: &cidrtree.Table[proto.Stream_PersistentStreamServer]{}}
}

func (ds *RouteTable) Add(cidr string, val proto.Stream_PersistentStreamServer) error {
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return err
	}
	ds.rt.Insert(prefix, val)
	return nil
}

func (ds *RouteTable) Delete(cidr string) error {
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return err
	}
	ds.rt.Delete(prefix)
	return nil
}

func (ds *RouteTable) DeleteBatch(cidrs []string) {
	for _, cidr := range cidrs {
		if prefix, err := netip.ParsePrefix(cidr); err == nil {
			ds.rt.Delete(prefix)
		}
	}
}

//func (ds *RouteTable) Lookup(addr string) (t proto.Stream_PersistentStreamServer, ok bool) {
//	ipaddr, err := netip.ParseAddr(addr)
//	if err != nil {
//		return t, false
//	}
//	_, t, ok = ds.rt.Lookup(ipaddr)
//	return t, ok
//}

func (ds *RouteTable) Lookup(addr net.IP) (t proto.Stream_PersistentStreamServer, ok bool) {
	ipaddr, ok := netip.AddrFromSlice(addr.To4())
	if !ok {
		return t, false
	}
	_, t, ok = ds.rt.Lookup(ipaddr)
	return t, ok
}
