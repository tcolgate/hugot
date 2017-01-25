package help_test

import (
	"context"
	"testing"

	"github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/handlers/command"
	"github.com/tcolgate/hugot/handlers/command/testcli"
	"github.com/tcolgate/hugot/handlers/mux"
	"github.com/tcolgate/hugot/hugottest"
)

func TestHelp_Match(t *testing.T) {
	mx := mux.New("test", "test mux")
	cs := command.NewSet()
	cs.Add(testcli.New())
	mx.ToBot = cs

	in := make(chan *hugot.Message)
	out := make(chan hugot.Message)
	ta := hugottest.NewAdapter(in, out)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hugot.ListenAndServe(ctx, mx, ta)

	ta.MessagesIn <- &hugot.Message{Text: "help"}
	m := <-out
	expected := "heard"
	if m.Text != expected {
		t.Fatalf("message did not contant expected test, wanted = %q got = %q", expected, m.Text)
	}
}
