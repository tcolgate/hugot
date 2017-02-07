package hugot

import "context"

// BackgroundHandler gets run when the bot starts listening. They are
// intended for publishing messages that are not in response to any
// specific incoming message.
type BackgroundHandler interface {
	Describer
	StartBackground(ctx context.Context, w ResponseWriter)
}

type baseBackgroundHandler struct {
	name string
	desc string
	bhf  BackgroundFunc
}

// BackgroundFunc describes the calling convention for Background handlers
type BackgroundFunc func(ctx context.Context, w ResponseWriter)

// NewBackgroundHandler wraps f up as a BackgroundHandler with the name and
// description provided.
func NewBackgroundHandler(name, desc string, f BackgroundFunc) BackgroundHandler {
	return &baseBackgroundHandler{
		name: name,
		desc: desc,
		bhf:  f,
	}
}

func (bbh *baseBackgroundHandler) Describe() (string, string) {
	return bbh.name, bbh.desc
}

func (bbh *baseBackgroundHandler) StartBackground(ctx context.Context, w ResponseWriter) {
	bbh.bhf(ctx, w)
}
