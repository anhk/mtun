// 注册中心
package datastore

import (
	"net/netip"

	"github.com/gaissmai/cidrtree"
)

type DataStore[T any] struct {
	rt *cidrtree.Table[T]
}

func NewDataStore[T any]() *DataStore[T] {
	return &DataStore[T]{
		rt: &cidrtree.Table[T]{},
	}
}

func (ds *DataStore[T]) AddRoute(cidr string, val T) error {
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return err
	}
	ds.rt.Insert(prefix, val)
	return nil
}

func (ds *DataStore[T]) DelRoute(cidr string) error {
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return err
	}
	ds.rt.Delete(prefix)
	return nil
}

func (ds *DataStore[T]) Lookup(addr string) (t T, ok bool) {
	ipaddr, err := netip.ParseAddr(addr)
	if err != nil {
		return t, false
	}
	_, t, ok = ds.rt.Lookup(ipaddr)
	return t, ok
}
