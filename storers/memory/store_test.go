package memory_test

import (
	"testing"

	"github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/storers/memory"
)

func TestStore(t *testing.T) {
	var i interface{}
	s := memory.New()
	i = s
	_, ok := i.(hugot.Storer)

	if !ok {
		t.Fatalf("%T does not support hugot.Storer", s)
	}
}

func TestMemStore_Get(t *testing.T) {
	s := memory.New()
	s.Set("test", "testval")

	v, ok, err := s.Get("test")
	if string(v) != "testval" || !ok || err != nil {
		t.Fatalf("Get failed, v = %v, ok = %v, err = %v ", v, ok, err)
	}
}

func TestMemStore_Set(t *testing.T) {
	s := memory.New()
	s.Set("test", "testval")

	v, ok, err := s.Get("test")
	if string(v) != "testval" || !ok || err != nil {
		t.Fatalf("Set failed, %v, %v , %v", v, ok, err)
	}

	s.Set("test", "anothertest")

	v, ok, err = s.Get("test")
	if string(v) != "anothertest" || !ok || err != nil {
		t.Fatalf("Set failed, %v, %v , %v", string(v), ok, err)
	}
}

func TestMemStore_Unet(t *testing.T) {
	s := memory.New()
	s.Set("test", "testval")

	v, ok, err := s.Get("test")
	if string(v) != "testval" || !ok || err != nil {
		t.Fatalf("Set failed, %v, %v , %v", v, ok, err)
	}

	s.Unset("test")

	v, ok, err = s.Get("test")
	if ok || err != nil {
		t.Fatalf("Unset failed, %v, %v , %v", string(v), ok, err)
	}
}
