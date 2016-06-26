package hugot

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/golang/glog"
)

type muxHelp struct {
	p *Mux
}

func (mx *muxHelp) Describe() (string, string) {
	return "help", "provides help"
}

func (mx *muxHelp) Command(ctx context.Context, w ResponseWriter, m *Message) error {
	//capture the command we were called as
	initcmd := m.args[0]
	m.Parse()
	cmds := m.Args()

	// list the full help
	if len(cmds) == 0 && initcmd == "help" {
		mx.fullHelp(ctx, w, m)
	} else {
		if initcmd != "help" {
			cmds = append([]string{initcmd}, cmds...)
		}
		mx.cmdHelp(ctx, w, cmds)
	}

	return ErrSkipHears
}

func (mx *muxHelp) fullHelp(ctx context.Context, w ResponseWriter, m *Message) {
	out := &bytes.Buffer{}
	fmt.Fprintf(out, "```")
	tw := new(tabwriter.Writer)
	tw.Init(out, 0, 8, 1, ' ', 0)

	if len(mx.p.cmds.SubCommands()) > 0 {
		fmt.Fprintf(out, "Available commands are:\n")
		for _, h := range mx.p.cmds.SubCommands() {
			n, d := h.Describe()
			fmt.Fprintf(tw, "  %s\t - %s\n", n, d)
		}
		tw.Flush()
	}

	if len(mx.p.hears) > 0 {
		fmt.Fprintf(out, "Active hear handlers are patternss are:\n")
		for r, hs := range mx.p.hears {
			for _, h := range hs {
				n, d := h.Describe()
				fmt.Fprintf(tw, "  %s\t%s\t - %s\n", n, r.String(), d)
			}
		}
		tw.Flush()
	}

	if len(mx.p.bghndlrs) > 0 {
		fmt.Fprintf(out, "Active background handlers are:\n")
		for _, h := range mx.p.bghndlrs {
			n, d := h.Describe()
			fmt.Fprintf(tw, "  %s\t - %s\n", n, d)
		}
		tw.Flush()
	}

	if len(mx.p.rhndlrs) > 0 {
		fmt.Fprintf(out, "Active raw handlers are:\n")
		for _, h := range mx.p.rhndlrs {
			n, d := h.Describe()
			fmt.Fprintf(tw, "  %s\t - %s\n", n, d)
		}
		tw.Flush()
	}
	fmt.Fprint(out, " ```")

	io.Copy(w, out)
}

func (mx *muxHelp) cmdHelp(ctx context.Context, w ResponseWriter, cmds []string) {
	var ch CommandHandler

	glog.Infof("%#v\n", cmds)
	chs := mx.p.cmds.SubCommands()

	var recHelp func(map[string]CommandHandler, []string) CommandHandler
	recHelp = func(chs map[string]CommandHandler, cmds []string) CommandHandler {
		if len(cmds) == 0 {
			return ch
		}
		want := cmds[0]
		left := cmds[1:]

		if mch, ok := chs[want]; ok {
		}

		glog.Infof("got: %v, want %#v left %#v\n", chs, want, left)

		return recHelp(chs, left)
	}

	recHelp(chs, cmds)

	return
}
