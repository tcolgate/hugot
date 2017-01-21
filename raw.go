package hugot

// RawHandler will recieve every message sent to the handler, without
// any filtering.
type RawHandler interface {
	Handler
	ReceieveAll()
}

type baseRawHandler struct {
	Handler
}

// NewRawHandler will wrap the function f as a RawHandler with the name
// and description provided
func NewRawHandler(name, desc string, f HandlerFunc) *baseRawHandler {
	return &baseRawHandler{
		Handler: NewBasicHandler(name, desc, f),
	}
}

func (brh *baseRawHandler) ReceieveAll() {
}
