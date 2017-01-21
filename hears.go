package hugot

import (
	"context"
	"regexp"
)

// HeardFunc describes the calling convention for a Hears handler.
type HeardFunc func(ctx context.Context, w ResponseWriter, m *Message, submatches [][]string) // Called once a message matches, and is passed any submatches from the regexp capture groups

// HearsHandler is a handler which responds to messages matching a specific
// pattern
type HearsHandler interface {
	Handler
	Hears() *regexp.Regexp                                                          // Returns the regexp we want to hear
	Heard(ctx context.Context, w ResponseWriter, m *Message, submatches [][]string) // Called once a message matches, and is passed any submatches from the regexp capture groups
}

type baseHearsHandler struct {
	Handler
	rgxp *regexp.Regexp
	hhf  HeardFunc
}

// NewHearsHandler wraps f as a Hears handler that reponnds to the regexp provided, with the given name a description
func NewHearsHandler(name, desc string, rgxp *regexp.Regexp, f HeardFunc) HearsHandler {
	h := &baseHearsHandler{
		rgxp: rgxp,
		hhf:  f,
	}

	h.Handler = NewBasicHandler(name, desc, h.processMessage)

	return h
}

func (bhh *baseHearsHandler) Hears() *regexp.Regexp {
	return bhh.rgxp
}

func (bhh *baseHearsHandler) Heard(ctx context.Context, w ResponseWriter, m *Message, submatches [][]string) {
	bhh.hhf(ctx, w, m, submatches)
}

// runHearsHandler will match the go routine.
func (bhh *baseHearsHandler) processMessage(ctx context.Context, w ResponseWriter, m *Message) error {
	defer glogPanic()

	if mtchs := bhh.Hears().FindAllStringSubmatch(m.Text, -1); mtchs != nil {
		go bhh.Heard(ctx, w, m, mtchs)
	}
	return nil
}
