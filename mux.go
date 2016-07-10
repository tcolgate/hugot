// Copyright (c) 2016 Tristan Colgate-McFarlane
//
// This file is part of hugot.
//
// hugot is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// hugot is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with hugot.  If not, see <http://www.gnu.org/licenses/>.

package hugot

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sync"

	"github.com/golang/glog"

	"context"
)

func init() {
	DefaultMux = NewMux("hugot", "")
	http.Handle("/hugot", DefaultMux)
	http.Handle("/hugot/", DefaultMux)
}

// Mux is a Handler that multiplexes messages to a set of Command, Hears, and
// Raw handlers.
type Mux struct {
	name string
	desc string

	burl *url.URL

	*sync.RWMutex
	hndlrs   []Handler                         // All the handlers
	rhndlrs  []RawHandler                      // Raw handlers
	bghndlrs []BackgroundHandler               // Long running background handlers
	whhndlrs map[string]WebHookHandler         // WebHooks
	hears    map[*regexp.Regexp][]HearsHandler // Hearing handlers
	cmds     *CommandSet                       // Command handlers
	httpm    *http.ServeMux                    // http Mux
	whsndr   Sender                            // Sender to be used by webhooks
}

// DefaultMux is a default Mux instance, http Handlers will be added to
// http.DefaultServeMux
var DefaultMux *Mux

// NewMux creates a new Mux.
func NewMux(name, desc string) *Mux {
	mx := &Mux{
		name:     name,
		desc:     desc,
		RWMutex:  &sync.RWMutex{},
		rhndlrs:  []RawHandler{},
		bghndlrs: []BackgroundHandler{},
		whhndlrs: map[string]WebHookHandler{},
		hears:    map[*regexp.Regexp][]HearsHandler{},
		cmds:     NewCommandSet(),
		httpm:    http.NewServeMux(),
		whsndr:   nil,
		burl:     &url.URL{Path: "/" + name},
	}
	mx.HandleCommand(&muxHelp{mx})
	return mx
}

// Describe implements the Describe method of Handler for
// the Mux
func (mx *Mux) Describe() (string, string) {
	return mx.name, mx.desc
}

// URL returns the base URL for the default Mux
func URL() *url.URL {
	return DefaultMux.URL()
}

// URL returns the base URL for this Mux
func (mx *Mux) URL() *url.URL {
	mx.RLock()
	defer mx.RUnlock()
	return mx.url()
}

func (mx *Mux) url() *url.URL {
	return mx.burl
}

// SetURL sets the base URL for web hooks.
func SetURL(b *url.URL) {
	DefaultMux.SetURL(b)
}

// SetURL sets the base URL for this mux's web hooks.
func (mx *Mux) SetURL(b *url.URL) {
	mx.Lock()
	defer mx.Unlock()

	mx.burl = b
	for _, h := range mx.whhndlrs {
		n, _ := h.Describe()
		p := fmt.Sprintf("%s/%s/%s/", b.Path, mx.name, n)
		nu := *b
		nu.Path = p
		h.SetURL(&nu)
	}
}

// StartBackground starts any registered background handlers.
func (mx *Mux) StartBackground(ctx context.Context, w ResponseWriter) {
	mx.Lock()
	defer mx.Unlock()
	mx.whsndr = w

	for _, h := range mx.bghndlrs {
		go RunBackgroundHandler(ctx, h, w)
	}
}

// ProcessMessage implements the Handler interface. Message will first be passed to
// any registered RawHandlers. If the message has been deemed, by the Adapter
// to have been sent directly to the bot, any comand handlers will be processed.
// Then, if appropriate, the message will be matched against any Hears patterns
// and all matching Heard functions will then be called.
// Any unrecognized errors from the Command handlers will be passed back to the
// user that sent us the message.
func (mx *Mux) ProcessMessage(ctx context.Context, w ResponseWriter, m *Message) error {
	mx.RLock()
	defer mx.RUnlock()
	var err error

	// We run all raw message handlers
	for _, rh := range mx.rhndlrs {
		mc := *m
		go rh.ProcessMessage(ctx, w, &mc)
	}

	if m.ToBot {
		err = mx.cmds.NextCommand(ctx, w, m)
	}

	if err == ErrSkipHears {
		return nil
	}

	for _, hhs := range mx.hears {
		for _, hh := range hhs {
			mc := *m
			if RunHearsHandler(ctx, hh, w, &mc) {
				err = nil
			}
		}
	}

	if err != nil {
		fmt.Fprintf(w, "error, %s", err.Error())
	}

	return nil
}

// Handle adds the provided handler to the DefaultMux
func Handle(h Handler) error {
	return DefaultMux.Handle(h)
}

// Handle adds a generic handler that supports one or more of the handler
// types. WARNING: This may be removed in the future. Prefer to
// the specific Add*Handler methods.
func (mx *Mux) Handle(h Handler) error {
	var used bool
	if h, ok := h.(RawHandler); ok {
		mx.HandleRaw(h)
		used = true
	}

	if h, ok := h.(BackgroundHandler); ok {
		mx.HandleBackground(h)
		used = true
	}

	if h, ok := h.(CommandHandler); ok {
		mx.HandleCommand(h)
		used = true
	}

	if h, ok := h.(HearsHandler); ok {
		mx.HandleHears(h)
		used = true
	}

	if h, ok := h.(WebHookHandler); ok {
		mx.HandleHTTP(h)
		used = true
	}

	mx.Lock()
	defer mx.Unlock()

	if !used {
		return fmt.Errorf("Don't know how to use %T as a handler", h)
	}

	mx.hndlrs = append(mx.hndlrs, h)

	return nil
}

// HandleRaw adds the provided handler to the DefaultMux
func HandleRaw(h RawHandler) error {
	return DefaultMux.HandleRaw(h)
}

// HandleRaw adds the provided handler to the Mux. All
// messages sent to the mux will be forwarded to this handler.
func (mx *Mux) HandleRaw(h RawHandler) error {
	mx.Lock()
	defer mx.Unlock()

	if h, ok := h.(RawHandler); ok {
		mx.rhndlrs = append(mx.rhndlrs, h)
	}

	return nil
}

// HandleBackground adds the provided handler to the DefaultMux
func HandleBackground(h BackgroundHandler) error {
	return DefaultMux.HandleBackground(h)
}

// HandleBackground adds the provided handler to the Mux. It
// will be started with the Mux is started.
func (mx *Mux) HandleBackground(h BackgroundHandler) error {
	mx.Lock()
	defer mx.Unlock()
	//name, _ := h.Describe()

	mx.bghndlrs = append(mx.bghndlrs, h)

	return nil
}

// HandleHears adds the provided handler to the DefaultMux
func HandleHears(h HearsHandler) error {
	return DefaultMux.HandleHears(h)
}

// HandleHears adds the provided handler to the mux. All
// messages matching the Hears patterns will be forwarded to
// the handler.
func (mx *Mux) HandleHears(h HearsHandler) error {
	mx.Lock()
	defer mx.Unlock()

	r := h.Hears()
	mx.hears[r] = append(mx.hears[r], h)

	return nil
}

// HandleCommand adds the provided handler to the DefaultMux
func HandleCommand(h CommandHandler) {
	DefaultMux.HandleCommand(h)
}

// HandleCommand adds the provided handler to the mux.
func (mx *Mux) HandleCommand(h CommandHandler) {
	mx.Lock()
	defer mx.Unlock()

	mx.cmds.AddCommandHandler(h)
}

type webHookBridge struct {
	nh http.Handler
}

func (whb *webHookBridge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if glog.V(2) {
		glog.Infof("webHookBridge ServeHTTP %v %s\n", *whb, *r)
	}
	whb.nh.ServeHTTP(w, r)
}

// HandleHTTP adds the provided handler to the DefaultMux
func HandleHTTP(h WebHookHandler) {
	DefaultMux.HandleHTTP(h)
}

// HandleHTTP registers h as a WebHook handler. The name
// of the Mux, and the name of the handler are used to
// construct a unique URL that can be used to send web
// requests to this handler
func (mx *Mux) HandleHTTP(h WebHookHandler) {
	mx.Lock()
	defer mx.Unlock()

	n, _ := h.Describe()
	p := fmt.Sprintf("/%s/%s", mx.name, n)
	mx.httpm.Handle(p, h)
	mx.httpm.Handle(p+"/", h)
	if glog.V(2) {
		glog.Infof("registering %v at %s, on %v\n", h, p, mx.httpm)
	}
	mx.whhndlrs[n] = h

	mu := mx.url()
	nu := *mu
	nu.Path = fmt.Sprintf("/%s/%s/", mx.name, n)
	h.SetURL(&nu)
}

// ServeHTTP iplements http.ServeHTTP for a Mux to allow it to
// act as a web server.
func (mx *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "in here at least")
	if glog.V(2) {
		glog.Infof("Mux ServeHTTP %s\n", *r)
	}
	mx.httpm.ServeHTTP(w, r)
}
