package hugot

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"strings"
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
	var cx *CommandMux
	var path []string

	cx = mx.p.cmds

	for {
		glog.Infof("%#v\n", cx)
		if len(cmds) == 0 {
			break
		}
		subs := cx.SubCommands()
		if len(subs) == 0 {
			glog.Info("ran out of commands")
			return
		}

		if cmd, ok := subs[cmds[0]]; ok {
			path = append(path, cmds[0])
			cmds = cmds[1:]
			cx = cmd
		} else {
			fmt.Fprintf(w, "unknown command %s", cmds[0])
			return
		}
		glog.Infof("%v %v %v\n", path, cmds, cx)
	}

	cmdStr := strings.Join(path, " ")
	m := &Message{args: []string{cmdStr, "-help"}}
	m.flagOut = &bytes.Buffer{}
	m.FlagSet = flag.NewFlagSet(cmdStr, flag.ContinueOnError)
	m.FlagSet.SetOutput(m.flagOut)

	cx.Command(ctx, w, m)
	subs := cx.SubCommands()
	if len(subs) > 0 {
		fmt.Fprintf(m.flagOut, "  Sub commands:\n")
		for n, s := range subs {
			_, desc := s.Describe()
			fmt.Fprintf(m.flagOut, "    %s - %s\n", n, desc)
		}
	}

	_, desc := cx.Describe()
	fmt.Fprintf(w, "```Description: %s\n%s ```", desc, m.flagOut.String())

	return
}

func formatHelp() string {
}
