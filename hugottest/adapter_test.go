package hugottest

import (
	"testing"

	"github.com/tcolgate/hugot"
)

func TestAdapter_NewAdapter(t *testing.T) {
	var a hugot.Adapter
	expect := "test message"
	a = NewAdapter([]*hugot.Message{{Text: expect}}...)

	m := <-a.Receive()
	if m.Text != expect {
		t.Fatalf("expect Text %#v, got %#v", expect, m.Text)
		return
	}

	m = <-a.Receive()
	if m != nil {
		t.Fatalf("expect nil, got %#v", m)
	}
}
