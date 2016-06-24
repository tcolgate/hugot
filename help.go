package hugot

import (
	"bytes"
	"context"
	"flag"
	"fmt"
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

func (mx *muxHelp) Command(ctx context.Context, s Sender, m *Message) error {
	log.Println(m)
	log.Println(m.FlagSet)
	m.Parse()

	if len(m.Args()) == 0 {
		out := &bytes.Buffer{}
		fmt.Fprintf(out, "```\n")
		w := new(tabwriter.Writer)
		w.Init(out, 0, 8, 1, ' ', 0)

		if len(mx.p.cmds) > 0 {
			fmt.Fprintf(w, "Available commands are:\n")
			for _, h := range mx.p.cmds {
				n, d := h.Describe()
				fmt.Fprintf(w, "  %s\t - %s\n", n, d)
			}
			w.Flush()
		}

		if len(mx.p.hears) > 0 {
			fmt.Fprintf(out, "Active hear handlers are patternss are:\n")
			for r, hs := range mx.p.hears {
				for _, h := range hs {
					n, d := h.Describe()
					fmt.Fprintf(w, "  %s\t%s\t - %s\n", n, r.String(), d)
				}
			}
			w.Flush()
		}

		if len(mx.p.bghndlrs) > 0 {
			fmt.Fprintf(out, "Active background handlers are:\n")
			for _, h := range mx.p.bghndlrs {
				n, d := h.Describe()
				fmt.Fprintf(w, "  %s\t - %s\n", n, d)
			}
			w.Flush()
		}

		if len(mx.p.rhndlrs) > 0 {
			fmt.Fprintf(out, "Active raw handlers are:\n")
			for _, h := range mx.p.rhndlrs {
				n, d := h.Describe()
				fmt.Fprintf(w, "  %s\t - %s\n", n, d)
			}
			w.Flush()
		}
		s.Send(ctx, m.Reply(out.String()+"```"))
	}

	if len(m.Args()) >= 1 {
		if c, ok := mx.p.cmds[m.Args()[0]]; ok {
			n, _ := c.Describe()
			m.Text = n + " -h"
			m.FlagSet = flag.NewFlagSet(n, flag.ContinueOnError)
			m.flagOut = &bytes.Buffer{}
			m.FlagSet.SetOutput(m.flagOut)
			err := c.Command(ctx, s, m)
			log.Println(err)
			s.Send(ctx, m.Reply("```"+m.flagOut.String()+"```"))

			return nil
		} else {
			argList := []string{}
			for n := range mx.p.cmds {
				argList = append(argList, n)
			}
			s.Send(ctx, m.Reply("Unkown command, available commands are: "+strings.Join(argList, ",")))
		}
	}

	return nil
}
