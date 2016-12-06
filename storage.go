package hugot

import (
	"strings"

	"github.com/docker/libkv/store"
)

type Storer interface {
	Get(key string) (*store.KVPair, error)
	Put(key string, value []byte, options *store.WriteOptions) error
	Delete(key string) error
}

type prefixStore struct {
	pfx  string
	base Storer
}

func (p prefixStore) Get(key string) (*store.KVPair, error) {
	res, err := p.base.Get(prepend(key, p.pfx))
	if res != nil {
		res.Key = strip(res.Key, p.pfx)
	}
	return res, err
}

func (p prefixStore) Put(key string, value []byte, options *store.WriteOptions) error {
	return p.base.Put(prepend(key, p.pfx), value, options)
}

func (p prefixStore) Delete(key string) error {
	return p.base.Delete(prepend(key, p.pfx))
}

func newPrefixedStore(s Storer, pfx string) prefixStore {
	return prefixStore{
		pfx:  pfx,
		base: s,
	}
}

func prepend(s, pfx string) string {
	ks := store.SplitKey(s)
	ks = append([]string{pfx}, ks...)
	return "/" + strings.Join(ks, "/")
}

func strip(s, pfx string) string {
	ks := store.SplitKey(s)
	if len(ks) > 1 && ks[0] == pfx {
		ks = ks[1:]
	}
	return "/" + strings.Join(ks, "/")
}
