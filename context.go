package hugot

import "context"

type hugotCtxKey int

//var adapterKey = hugotCtxKey(1)
const (
	adapterKey hugotCtxKey = iota
)

// NewAdapterContext creates a context for passing an adapter. This is
// mainly used by web handlers.
func NewAdapterContext(ctx context.Context, a Adapter) context.Context {
	return context.WithValue(ctx, adapterKey, a)
}

// AdapterFromContext returns the Adapter stored in a context.
func AdapterFromContext(ctx context.Context) (Adapter, bool) {
	a, ok := ctx.Value(adapterKey).(Adapter)
	return a, ok
}

// SenderFromContext can be used to retrieve a valid sender from
// a context. This is mostly useful in WebHook handlers for sneding
// messages back to the inbound Adapter.
func SenderFromContext(ctx context.Context) (Sender, bool) {
	a, ok := ctx.Value(adapterKey).(Adapter)
	return a, ok
}
