package hugot

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
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
	log.Println(m)
	m.Parse()
	log.Println(m)

	if len(m.Args()) == 0 {
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

		return nil
	}

	if c, ok := mx.p.cmds.SubCommands()[m.Args()[0]]; ok {
		n, desc := c.Describe()
		m.Text = n + " -h"
		if len(m.Args()) > 1 {
			m.Text = "help " + n
		}
		m.FlagSet = flag.NewFlagSet(n, flag.ContinueOnError)
		m.flagOut = &bytes.Buffer{}
		fmt.Fprintf(m.flagOut, "```Description: %s\n", desc)
		m.FlagSet.SetOutput(m.flagOut)
		c.Command(ctx, w, m)
		fmt.Fprint(m.flagOut, " ```")
		io.Copy(w, m.flagOut)

		return nil
	}

	cmdList := []string{}
	for n := range mx.p.cmds.SubCommands() {
		cmdList = append(cmdList, n)
	}

	cmdStr := strings.Join(cmdList, ",")

	fmt.Fprintf(w, "Unkown command, available commands are: %s", cmdStr)

	return nil
}
