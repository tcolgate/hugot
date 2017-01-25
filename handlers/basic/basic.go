package basic

import (
	"context"

	"github.com/tcolgate/hugot"
)

type basicHandler struct {
	name string
	desc string
	f    hugot.HandlerFunc
}

// New creates a new basic handler that calls the provided
// hander Func
func New(name, desc string, doFunc hugot.HandlerFunc) hugot.Handler {
	return &basicHandler{name, desc, doFunc}
}

func (bh *basicHandler) Describe() (string, string) {
	return bh.name, bh.desc
}

func (bh *basicHandler) ProcessMessage(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	return bh.f(ctx, w, m)
}
