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

package mux

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sync"

	"github.com/golang/glog"
	"github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/handlers/hears"
	"github.com/tcolgate/hugot/storers/memory"

	"context"
)

func init() {
	m := New("hugot", "")
	hugot.DefaultHandler = m
	http.Handle("/hugot", m)
	http.Handle("/hugot/", m)

	//m.HandleCommand(NewMuxHelpHamdler(DefaultMux))
	//m.ah = NewAliasHandler(NewPrefixedStore(DefaultMux.store, "aliases"))
	//m.HandleCommand(DefaultMux.ah)

	hugot.DefaultHandler = m
}

// Mux is a Handler that multiplexes messages to a set of Command, Hears, and
// Raw handlers.
type Mux struct {
	name string
	desc string

	burl  *url.URL
	httpm *http.ServeMux // http Mux

	*sync.RWMutex
	hndlrs   []hugot.Handler                  // All the handlers
	rhndlrs  []hugot.RawHandler               // Raw handlers
	bghndlrs []hugot.BackgroundHandler        // Long running background handlers
	whhndlrs map[string]hugot.WebHookHandler  // WebHooks
	hears    map[*regexp.Regexp][]hears.Hears // Hearing handlers
	//cmds     *CommandSet                       // Command handlers

	store hugot.Storer
	//ah    *AliasHandler
}

// Opt functions are used to set options on the Mux
type Opt func(*Mux)

// DefaultMux is a default Mux instance, http Handlers will be added to
// http.DefaultServeMux
var DefaultMux *Mux

// New creates a new Mux.
func New(name, desc string, opts ...Opt) *Mux {
	mx := &Mux{
		name:     name,
		desc:     desc,
		RWMutex:  &sync.RWMutex{},
		rhndlrs:  []hugot.RawHandler{},
		bghndlrs: []hugot.BackgroundHandler{},
		whhndlrs: map[string]hugot.WebHookHandler{},
		hears:    map[*regexp.Regexp][]hears.Hears{},
		//		cmds:     NewCommandSet(),
		httpm: http.NewServeMux(),
		burl:  &url.URL{Path: "/" + name},
		store: memory.New(),
	}

	for _, opt := range opts {
		opt(mx)
	}

	return mx
}

// WithStore is a Mux option to set the store to be used
// for managing aliases, and property lookup.
func WithStore(s hugot.Storer) Opt {
	return func(m *Mux) {
		m.store = s
	}
}

// Describe implements the Describe method of Handler for
// the Mux
func (mx *Mux) Describe() (string, string) {
	return mx.name, mx.desc
}

// StartBackground starts any registered background handlers.
func (mx *Mux) StartBackground(ctx context.Context, w hugot.ResponseWriter) {
	mx.Lock()
	defer mx.Unlock()

	for _, h := range mx.bghndlrs {
		go h.StartBackground(ctx, w.Copy())
	}
}

// SetAdapter sets the adapter on all the webhook of this mux.
func (mx *Mux) SetAdapter(a hugot.Adapter) {
	mx.Lock()
	defer mx.Unlock()

	for _, wh := range mx.whhndlrs {
		wh.SetAdapter(a)
	}
}

// ProcessMessage implements the Handler interface. Message will first be passed to
// any registered RawHandlers. If the message has been deemed, by the Adapter
// to have been sent directly to the bot, any comand handlers will be processed.
// Then, if appropriate, the message will be matched against any Hears patterns
// and all matching Heard functions will then be called.
// Any unrecognized errors from the Command handlers will be passed back to the
// user that sent us the message.
func (mx *Mux) ProcessMessage(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	mx.RLock()
	defer mx.RUnlock()
	var err error

	// We run all raw message handlers
	for _, rh := range mx.rhndlrs {
		nm := m.Copy()
		go rh.ProcessMessage(ctx, w, nm)
	}

	/*
		if m.ToBot && m.Text != "" {
			nm := m.Copy()
			err = mx.cmds.ProcessMessage(ctx, w, nm)
		}
	*/

	/*
		if err == ErrSkipHears {
			return nil
		}
	*/

	for _, hhs := range mx.hears {
		for _, hh := range hhs {
			if hh.Hears().MatchString(m.Text) {
				nm := m.Copy()
				hn, _ := hh.Describe()
				nm.Store = hugot.NewPrefixedStore(hugot.DefaultStore, hn)
				err = hh.ProcessMessage(ctx, w, nm)
			}
		}
	}

	if err != nil {
		fmt.Fprintf(w, "error, %s", err.Error())
	}

	return nil
}

// HandleRaw adds the provided handler to the DefaultMux
func HandleRaw(h hugot.RawHandler) error {
	return DefaultMux.HandleRaw(h)
}

// HandleRaw adds the provided handler to the Mux. All
// messages sent to the mux will be forwarded to this handler.
func (mx *Mux) HandleRaw(h hugot.RawHandler) error {
	mx.Lock()
	defer mx.Unlock()

	if h, ok := h.(hugot.RawHandler); ok {
		mx.rhndlrs = append(mx.rhndlrs, h)
	}

	return nil
}

// HandleBackground adds the provided handler to the DefaultMux
func HandleBackground(h hugot.BackgroundHandler) error {
	return DefaultMux.HandleBackground(h)
}

// HandleBackground adds the provided handler to the Mux. It
// will be started with the Mux is started.
func (mx *Mux) HandleBackground(h hugot.BackgroundHandler) error {
	mx.Lock()
	defer mx.Unlock()
	//name, _ := h.Describe()

	mx.bghndlrs = append(mx.bghndlrs, h)

	return nil
}

/*
// HandleHears adds the provided handler to the DefaultMux
func HandleHears(h HearsHandler) error {
	return DefaultMux.HandleHears(h)
}

// HandleHears adds the provided handler to the mux. All
// messages matching the Hears patterns will be forwarded to
// the handler.
func (mx *Mux) HandleHears(h hugot.HearsHandler) error {
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
*/

// HandleHTTP adds the provided handler to the DefaultMux
func HandleHTTP(h hugot.WebHookHandler) {
	DefaultMux.HandleHTTP(h)
}

// HandleHTTP registers h as a WebHook handler. The name
// of the Mux, and the name of the handler are used to
// construct a unique URL that can be used to send web
// requests to this handler
func (mx *Mux) HandleHTTP(h hugot.WebHookHandler) {
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
	if glog.V(2) {
		glog.Infof("Mux ServeHTTP %s\n", *r)
	}
	mx.httpm.ServeHTTP(w, r)
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
	if b.Path != "" {
		panic(errors.New("Can't set URL with path at the moment, sorry"))
	}
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
