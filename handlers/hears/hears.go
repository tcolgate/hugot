package hears

import (
	"context"
	"regexp"

	"github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/handlers/basic"
)

// Hearer descibres a handler that can be used to match regular expressions.
type Hearer interface {
	hugot.Describer
	Heard(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message, submatches [][]string) error // Called once a message matches, and is passed any submatches from the regexp capture groups
	Hears() *regexp.Regexp                                                                            // Returns the regexp we want to hear
}

// HeardFunc describes a function that can be used to hear regex matches.
type HeardFunc func(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message, submatches [][]string) error

// Handler iimplements both the Hearer, and hugot.Handler interfaces
type Handler struct {
	hugot.Handler
	rgxp *regexp.Regexp
	f    HeardFunc
}

// New wraps f as a Hears handler that reponnds to the regexp provided, with the given name a description
func New(name, desc string, rgxp *regexp.Regexp, f HeardFunc) Hearer {
	h := &Handler{
		rgxp: rgxp,
		f:    f,
	}
	h.Handler = basic.New(name, desc, h.processMessage)

	return h
}

// Describe returns the description of this handler
func (h *Handler) Describe() (string, string) {
	return h.Handler.Describe()
}

// Hears returns the regylar expressions this handler wants to match.
func (h *Handler) Hears() *regexp.Regexp {
	return h.rgxp
}

// Heard is called when the handler hears a message matching the regex
func (h *Handler) Heard(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message, matches [][]string) error {
	return h.f(ctx, w, m, matches)
}

// processMessage will pass the given message to Heard if it matches the provided regex.
// It can be used to attach a hears handler to as a regular handler
func (h *Handler) processMessage(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	if matches := h.rgxp.FindAllStringSubmatch(m.Text, -1); matches != nil {
		return h.Heard(ctx, w, m, matches)
	}
	return nil
}
