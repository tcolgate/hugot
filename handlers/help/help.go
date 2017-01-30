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
	return "help", "provides description of handler usage"
}

func (h *Handler) Command(ctx context.Context, w hugot.ResponseWriter, m *command.Message) error {
	out := &bytes.Buffer{}
	m.Parse()
	if err := h.nh.Help(ctx, out, m); err != nil {
		w.Send(ctx, m.Reply(err.Error()))
	} else {
		w.Send(ctx, m.Reply(string(out.Bytes())))
	}
	return command.ErrSkipHears
}
