package hugot

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"context"
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
		err := mx.cmdHelp(ctx, w, cmds)
		if err != nil {
			return err
		}
	}

	return ErrSkipHears
}

func (mx *muxHelp) fullHelp(ctx context.Context, w ResponseWriter, m *Message) {
	out := &bytes.Buffer{}
	tw := new(tabwriter.Writer)
	tw.Init(out, 0, 8, 1, '\t', 0)

	if len(*mx.p.cmds) > 0 {
		fmt.Fprintf(out, "Available commands are:\n")
		_, _, hs := (*mx.p.cmds).List()
		for _, h := range hs {
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
				fmt.Fprintf(tw, "  %s\t`%s`\t - %s\n", n, r.String(), d)
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

	io.Copy(w, out)
}

func (mx *muxHelp) cmdHelp(ctx context.Context, w ResponseWriter, cmds []string) error {
	var cs *CommandSet
	var path []string

	cs = mx.p.cmds

	var cmd CommandHandler
	for {
		if len(cmds) == 0 {
			break
		}

		ok := false
		if cmd, ok = (*cs)[cmds[0]]; !ok {
			return ErrUnknownCommand
		}

		path = append(path, cmds[0])
		cmds = cmds[1:]
		if sch, ok := cmd.(CommandWithSubsHandler); ok {
			cs = sch.SubCommands()
		} else {
			break
		}
	}

	fmt.Fprint(w, cmdUsage(cmd, strings.Join(path, " "), nil))
	return nil
}

func cmdUsage(c CommandHandler, cmdStr string, err error) error {
	_, desc := c.Describe()
	m := &Message{args: []string{cmdStr, "-help"}}
	m.flagOut = &bytes.Buffer{}
	m.FlagSet = flag.NewFlagSet(cmdStr, flag.ContinueOnError)
	m.FlagSet.SetOutput(m.flagOut)

	c.Command(context.TODO(), NewNullResponseWriter(*m), m)
	if subcx, ok := c.(CommandWithSubsHandler); ok {
		subs := subcx.SubCommands()
		if subs != nil && len(*subs) > 0 {
			fmt.Fprintf(m.flagOut, "  Sub commands:\n")
			for n, s := range *subs {
				_, desc := s.Describe()
				fmt.Fprintf(m.flagOut, "    %s - %s\n", n, desc)
			}
		}
	}

	str := ""
	if err != nil {
		str = fmt.Sprintf("error, %s\n", err.Error())
	} else {
		str = fmt.Sprintf("Description: %s\n", desc)
	}
	return ErrUsage{str + m.flagOut.String()}
}
