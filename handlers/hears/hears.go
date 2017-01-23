package hears

import (
	"regexp"

	"github.com/tcolgate/hugot"
)

// Hears is a handler which responds to messages matching a specific
// pattern
type Hears interface {
	hugot.Handler
	Hears() *regexp.Regexp // Returns the regexp we want to hear
}

type baseHearsHandler struct {
	hugot.Handler
	rgxp *regexp.Regexp
}

// New wraps f as a Hears handler that reponnds to the regexp provided, with the given name a description
func New(name, desc string, rgxp *regexp.Regexp, f hugot.HandlerFunc) Hears {
	h := &baseHearsHandler{
		rgxp: rgxp,
	}

	h.Handler = hugot.NewBasicHandler(name, desc, f)

	return h
}

func (bhh *baseHearsHandler) Hears() *regexp.Regexp {
	return bhh.rgxp
}
