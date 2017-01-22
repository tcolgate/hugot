package hugot

import "regexp"

// HearsHandler is a handler which responds to messages matching a specific
// pattern
type HearsHandler interface {
	Handler
	Hears() *regexp.Regexp // Returns the regexp we want to hear
}

type baseHearsHandler struct {
	Handler
	rgxp *regexp.Regexp
}

// NewHearsHandler wraps f as a Hears handler that reponnds to the regexp provided, with the given name a description
func NewHearsHandler(name, desc string, rgxp *regexp.Regexp, f HandlerFunc) HearsHandler {
	h := &baseHearsHandler{
		rgxp: rgxp,
	}

	h.Handler = NewBasicHandler(name, desc, f)

	return h
}

func (bhh *baseHearsHandler) Hears() *regexp.Regexp {
	return bhh.rgxp
}
