package help

import (
	"bytes"
	"io"

	"github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/handlers/command"

	"context"
)

type Helper interface {
	Help(ctx context.Context, w io.Writer, m *command.Message) error
}

type Handler struct {
	nh Helper
}

func New(nh Helper) *Handler {
	h := &Handler{nh}
	return h
}

func (h *Handler) Describe() (string, string) {
	return "help", "provdes description of handler usage"
}

func (h *Handler) Command(ctx context.Context, w hugot.ResponseWriter, m *command.Message) error {
	out := &bytes.Buffer{}
	m.Parse()
	if err := h.nh.Help(ctx, out, m); err != nil {
		return err
	}
	w.Send(ctx, m.Reply(string(out.Bytes())))
	return command.ErrSkipHears
}

/*
func (h *MuxHelp) cmdHelp(ctx context.Context, w hugot.ResponseWriter, cmds []string) error {
	var cs command.Set
	var path []string

	cs = h.mx.Commands

	var cmd command.Commander
	for {
		if len(cmds) == 0 {
			break
		}

		ok := false
		if cmd, ok = cs[cmds[0]]; !ok {
			return command.ErrUnknownCommand
		}

		path = append(path, cmds[0])
		cmds = cmds[1:]
		if sch, ok := cmd.(command.CommanderWithSubs); ok {
			cs = sch.SubCommands()
		} else {
			break
		}
	}

	fmt.Fprint(w, cmdUsage(cmd, strings.Join(path, " "), nil))
	return nil
}

func cmdUsage(c command.Commander, cmdStr string, err error) error {
	_, desc := c.Describe()
	m := &command.Message{}}
	m.FlagOut = &bytes.Buffer{}
	m.FlagSet = flag.NewFlagSet(cmdStr, flag.ContinueOnError)
	m.FlagSet.SetOutput(m.FlagOut)

	c.Command(context.TODO(), hugot.NewNullResponseWriter(*m.Message), m)
	if subcx, ok := c.(command.CommanderWithSubs); ok {
		subs := subcx.SubCommands()
		if subs != nil && len(subs) > 0 {
			fmt.Fprintf(m.FlagOut, "  Sub commands:\n")
			for n, s := range subs {
				_, desc := s.Describe()
				fmt.Fprintf(m.FlagOut, "    %s - %s\n", n, desc)
			}
		}
	}

	str := ""
	if err != nil {
		str = fmt.Sprintf("error, %s\n", err.Error())
	} else {
		str = fmt.Sprintf("Description: %s\n", desc)
	}
	return command.ErrUsage(str + m.FlagOut.String())
}
*/
