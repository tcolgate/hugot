package hugot_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/hugottest"
)

type testHears struct {
	re *regexp.Regexp
}

func (th *testHears) Describe() (string, string) {
	return "th", "th desc"
}

func (th *testHears) Hears() *regexp.Regexp {
	return th.re
}

func (th *testHears) ProcessMessage(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	fmt.Fprintf(w, "heard")
	return nil
}

func TestHears_Match(t *testing.T) {
	th := &testHears{regexp.MustCompile("testing")}
	mx := hugot.NewMux("test", "test mux")
	mx.HandleHears(th)
	in := make(chan *hugot.Message)
	out := make(chan hugot.Message)
	ta := hugottest.NewAdapter(in, out)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hugot.ListenAndServe(ctx, mx, ta)

	ta.MessagesIn <- &hugot.Message{Text: "testing testing 123"}
	m := <-out
	expected := "heard"
	if m.Text != expected {
		t.Fatalf("message did not contant expected test, wanted = %q got = %q", expected, m.Text)
	}
}

func TestHears_NoMatch(t *testing.T) {
	th := &testHears{regexp.MustCompile("testing")}
	mx := hugot.NewMux("test", "test mux")
	mx.HandleHears(th)
	in := make(chan *hugot.Message)
	out := make(chan hugot.Message)
	ta := hugottest.NewAdapter(in, out)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hugot.ListenAndServe(ctx, mx, ta)

	ta.MessagesIn <- &hugot.Message{Text: "123"}
	select {
	case m := <-out:
		t.Fatalf("Should not have got message, but got m = %#v", m)
	case <-time.After(5 * time.Millisecond):
	}
}
