package hugot

import (
	"testing"

	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/mock"
)

func newMock() (store.Store, error) {
	return mock.New([]string{}, nil)
}

func TestPrefixStore_Put(t *testing.T) {
	s, _ := newMock()

	s.(*mock.Mock).On("Put", "/myhandler/a key", []byte("value"), (*store.WriteOptions)(nil)).Return(nil)
	ps := newPrefixedStore(s, "myhandler")

	ps.Put("a key", []byte("value"), nil)

	s.(*mock.Mock).AssertExpectations(t)
}

func TestPrefixStore_Get(t *testing.T) {
	s, _ := newMock()

	s.(*mock.Mock).On("Get", "/myhandler/a key").Return(&store.KVPair{Key: "/myhandler/a key", Value: []byte("value")}, nil)

	ps := newPrefixedStore(s, "myhandler")

	v, _ := ps.Get("a key")

	s.(*mock.Mock).AssertExpectations(t)
	if v.Key != "a key" {
		t.Fatal("wrong key value")
	}
}

func TestPrefixedStore_Get(t *testing.T) {
}
