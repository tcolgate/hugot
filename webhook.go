package hugot

import (
	"context"
	"net/http"
	"net/url"

	"github.com/golang/glog"
)

// WebHookHandler handlers are used to expose a registered handler via a web server.
// The SetURL method is called to inform the handler what it's external URL will be.
// This will normally be done by the Mux. Other handlers can use URL to generate
// links suitable for external use.
// You can use the http.Handler Request.Context() to get a ResponseWriter to write
// into the bots adapters. You need to SetChannel the resulting ResponseWriter to
// send messages.
type WebHookHandler interface {
	Describer
	URL() *url.URL      // Is called to retrieve the location of the Handler
	SetURL(*url.URL)    // Is called after the WebHook is added, to inform it where it lives
	SetAdapter(Adapter) // Is called to set the default adapter for this handler to use
	http.Handler
}

// WebHookHandlerFunc describes the calling convention for a WebHook.
type WebHookHandlerFunc func(ctx context.Context, hw ResponseWriter, w http.ResponseWriter, r *http.Request)

type baseWebHookHandler struct {
	ctx  context.Context
	name string
	desc string
	a    Adapter
	hf   http.HandlerFunc
	url  *url.URL
}

// ServeHTTP  implement the http.Handler interface for a baseWebHandler
func (bwhh *baseWebHookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = NewAdapterContext(ctx, bwhh.a)
	r = r.WithContext(ctx)

	bwhh.hf(w, r)
}

// NewWebHookHandler creates a new WebHookHandler provided name and description.
func NewWebHookHandler(name, desc string, hf http.HandlerFunc) WebHookHandler {
	h := &baseWebHookHandler{
		name: name,
		desc: desc,
		url:  &url.URL{},
		hf:   hf,
	}
	return h
}

func (bwhh *baseWebHookHandler) Describe() (string, string) {
	return bwhh.name, bwhh.desc
}

func (bwhh *baseWebHookHandler) SetURL(u *url.URL) {
	if glog.V(2) {
		glog.Infof("SetURL to %s", *u)
	}
	bwhh.url = u
}

func (bwhh *baseWebHookHandler) URL() *url.URL {
	return bwhh.url
}

func (bwhh *baseWebHookHandler) SetAdapter(a Adapter) {
	if glog.V(3) {
		glog.Infof("WebHander adapter set to %#v", a)
	}
	bwhh.a = a
}
