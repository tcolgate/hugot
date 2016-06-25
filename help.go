package hugot

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
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
		if initcmd == "help" {
			cmds = cmds[1:]
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
	var ok bool

	chs := mx.p.cmds.SubCommands()

	//locate the right handler
	for {
		if ch, ok = chs[cmds[0]]; !ok {
			// no such command, return an error
			fmt.Fprintf(w, "NO SUCH COMMAND IN HERE")
			return
		}
		if chsubs, ok := ch.(CommandWithSubsHandler); ok {
			chs = chsubs.SubCommands()
		}
		cmds := cmds[1:]
		if len(cmds) == 0 {
			break
		}
	}

	// we found a matching handler, ask it for help
	if c, ok := mx.p.cmds.SubCommands()[cmd]; ok {
		for _, n := range m.Args() {
		}
		m.FlagSet = flag.NewFlagSet(n, flag.ContinueOnError)
		m.flagOut = &bytes.Buffer{}
		fmt.Fprintf(m.flagOut, "```Description: %s\n", desc)
		m.FlagSet.SetOutput(m.flagOut)
		c.Command(ctx, w, m)
		fmt.Fprint(m.flagOut, " ```")
		io.Copy(w, m.flagOut)

		return ErrSkipHears
	}

	cmdList := []string{}
	for n := range mx.p.cmds.SubCommands() {
		cmdList = append(cmdList, n)
	}

	cmdStr := strings.Join(cmdList, ",")

	fmt.Fprintf(w, "Unknown command, available commands are: %s", cmdStr)

	return ErrSkipHears
}
