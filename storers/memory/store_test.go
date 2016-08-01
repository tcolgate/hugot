package memory

import (
	"testing"

	"github.com/tcolgate/hugot"
)

func TestStore(t *testing.T) {
	var i interface{}
	s := New()
	i = s
	_, ok := i.(hugot.Storer)

	if !ok {
		t.Fatalf("%T does not support hugot.Storer", s)
	}
}

func TestMemStore_Get(t *testing.T) {
	s := New()
	s.Set([]byte("test"), []byte("testval"))

	v, ok, err := s.Get([]byte("test"))
	if string(v) != "testval" || !ok || err != nil {
		t.Fatalf("Get failed, ", v, ok, err)
	}
}

func TestMemStore_Set(t *testing.T) {
	s := New()
	s.Set([]byte("test"), []byte("testval"))

	v, ok, err := s.Get([]byte("test"))
	if string(v) != "testval" || !ok || err != nil {
		t.Fatalf("Set failed, %v, %v , %v", v, ok, err)
	}

	s.Set([]byte("test"), []byte("anothertest"))

	v, ok, err = s.Get([]byte("test"))
	if string(v) != "anothertest" || !ok || err != nil {
		t.Fatalf("Set failed, %v, %v , %v", string(v), ok, err)
	}
}

func TestMemStore_Unet(t *testing.T) {
	s := New()
	s.Set([]byte("test"), []byte("testval"))

	v, ok, err := s.Get([]byte("test"))
	if string(v) != "testval" || !ok || err != nil {
		t.Fatalf("Set failed, %v, %v , %v", v, ok, err)
	}

	s.Unset([]byte("test"))

	v, ok, err = s.Get([]byte("test"))
	if ok || err != nil {
		t.Fatalf("Unset failed, %v, %v , %v", string(v), ok, err)
	}
}
