package hugot

import (
	"context"

	"github.com/golang/glog"
)

// BackgroundHandler gets run when the bot starts listening. They are
// intended for publishing messages that are not in response to any
// specific incoming message.
type BackgroundHandler interface {
	Handler
	StartBackground(ctx context.Context, w ResponseWriter)
}

// runBackgroundHandler starts the provided BackgroundHandler in a new
// go routine.
func runBackgroundHandler(ctx context.Context, h BackgroundHandler, w ResponseWriter) {
	glog.Infof("Starting background %v\n", h)
	go func(ctx context.Context, bh BackgroundHandler) {
		defer glogPanic()
		h.StartBackground(ctx, w)
	}(ctx, h)
}

type baseBackgroundHandler struct {
	Handler
	bhf BackgroundFunc
}

// BackgroundFunc describes the calling convention for Background handlers
type BackgroundFunc func(ctx context.Context, w ResponseWriter)

// NewBackgroundHandler wraps f up as a BackgroundHandler with the name and
// description provided.
func NewBackgroundHandler(name, desc string, f BackgroundFunc) BackgroundHandler {
	return &baseBackgroundHandler{
		Handler: NewBasicHandler(name, desc, nil),
		bhf:     f,
	}
}

func (bbh *baseBackgroundHandler) StartBackground(ctx context.Context, w ResponseWriter) {
	bbh.bhf(ctx, w)
}
