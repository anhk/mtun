package grpc

import (
	"github.com/anhk/mtun/proto"
	"github.com/gaissmai/cidrtree"
	"net/netip"
)

type DataStore struct {
	rt *cidrtree.Table[proto.Stream_PersistentStreamServer]
}

func NewDataStore() *DataStore {
	return &DataStore{rt: &cidrtree.Table[proto.Stream_PersistentStreamServer]{}}
}

func (ds *DataStore) Add(cidr string, val proto.Stream_PersistentStreamServer) error {
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return err
	}
	ds.rt.Insert(prefix, val)
	return nil
}

func (ds *DataStore) Delete(cidr string) error {
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return err
	}
	ds.rt.Delete(prefix)
	return nil
}

func (ds *DataStore) DeleteBatch(cidrs []string) {
	for _, cidr := range cidrs {
		if prefix, err := netip.ParsePrefix(cidr); err == nil {
			ds.rt.Delete(prefix)
		}
	}
}

func (ds *DataStore) Lookup(addr string) (t proto.Stream_PersistentStreamServer, ok bool) {
	ipaddr, err := netip.ParseAddr(addr)
	if err != nil {
		return t, false
	}
	_, t, ok = ds.rt.Lookup(ipaddr)
	return t, ok
}
