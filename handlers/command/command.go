package command

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io"

	shellwords "github.com/mattn/go-shellwords"
	"github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/handlers/basic"
)

type ctxKey int

const (
	ctxPathKey ctxKey = iota
)

// Message is a message struct adapted for handling command style
// operations
type Message struct {
	*hugot.Message
	*flag.FlagSet
	// FlatOut is a bytes.Buffer containing the output of any
	// actions on the FlagSet
	FlagOut *bytes.Buffer

	args []string
}

// Func describes the calling convention for CommandHandler
type Func func(ctx context.Context, w hugot.ResponseWriter, m *Message) error

// Commander handlers are used to implement CLI style commands. Before the
// Command method is called, the in the incoming message m will have the Text
// of the message parsed into invidual strings, accouting for quoting.
// m.Args[0] will be the name of the command as the handler was called, as per
// os.Args(). Command should add any requires flags to m and then call m.Parse()
type Commander interface {
	hugot.Describer
	Command(ctx context.Context, w hugot.ResponseWriter, m *Message) error
}

// Handler implements a command handler that acts like a command
// line tool.
type Handler struct {
	hugot.Handler
	bcf  Func
	subs Set
}

// New wraps the given function f as a CommandHandler with the
// provided name and description.
func New(name, desc string, f Func) *Handler {
	h := &Handler{
		bcf: f,
	}
	h.Handler = basic.New(name, desc, h.ProcessMessage)
	return h
}

//Command implements the Commander interface.
func (bch *Handler) Command(ctx context.Context, w hugot.ResponseWriter, m *Message) error {
	return bch.bcf(ctx, w, m)
}

// PathFromContext returns the path used to get to
// this command handler
func PathFromContext(ctx context.Context) []string {
	iv := ctx.Value(ctxPathKey)

	if iv == nil {
		return []string{}
	}

	v := iv.([]string)
	return v
}

// ProcessMessage allows you to use a command handler as a raw message
func (bch *Handler) ProcessMessage(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	var err error

	cm := &Message{}

	cm.args, err = shellwords.Parse(m.Text)
	if err != nil {
		return ErrBadCLI
	}

	if len(cm.args) == 0 {
		return errors.New("command handler called with no possible arguments")
	}

	name := cm.args[0]
	cm.FlagOut = &bytes.Buffer{}
	cm.FlagSet = flag.NewFlagSet(name, flag.ContinueOnError)
	cm.FlagSet.SetOutput(cm.FlagOut)

	err = bch.Command(ctx, w, cm)
	return err
}

// Help sends help about this command to the user.
func (bch *Handler) Help(ctx context.Context, w io.Writer, m *Message) error {
	/*
		//capture the command we were called as
		initcmd := m.Args()[0]
		m.FlagSet.Parse()
		cmds := m.Args()

		// list the full help
		if len(cmds) == 0 && initcmd == "help" {
			h.fullhelp(ctx, w, m)
		} else {
			if initcmd != "help" {
				cmds = append([]string{initcmd}, cmds...)
			}
			err := h.cmdhelp(ctx, w, cmds)
			if err != nil {
				return err
			}
		}

		return command.errskiphears
	*/
	return nil
}
