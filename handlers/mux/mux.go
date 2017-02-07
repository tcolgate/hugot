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
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"text/tabwriter"

	"github.com/golang/glog"
	"github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/handlers/command"
	"github.com/tcolgate/hugot/handlers/hears"
	"github.com/tcolgate/hugot/handlers/help"
	"github.com/tcolgate/hugot/storage"
	"github.com/tcolgate/hugot/storage/memory"
	"github.com/tcolgate/hugot/storage/prefix"

	"context"
)

// Mux is a Handler that multiplexes messages to a set of Command, Hears, and
// Raw handlers.
type Mux struct {
	name  string
	desc  string
	store storage.Storer

	burl  *url.URL
	httpm *http.ServeMux // http Mux

	sync.RWMutex
	ToBot         hugot.Handler                     // Handles message aimed directly at the bot
	RawHandlers   []hugot.Handler                   // Raw handlers
	BGHandlers    []hugot.BackgroundHandler         // Long running background handlers
	Webhooks      map[string]hugot.WebHookHandler   // WebHooks
	HearsHandlers map[*regexp.Regexp][]hears.Hearer // Hearing handlers
}

// Opt functions are used to set options on the Mux
type Opt func(*Mux)

// New creates a new Mux.
func New(name, desc string, opts ...Opt) *Mux {
	mx := &Mux{
		name: name,
		desc: desc,

		RWMutex:       sync.RWMutex{},
		RawHandlers:   []hugot.Handler{},
		BGHandlers:    []hugot.BackgroundHandler{},
		Webhooks:      map[string]hugot.WebHookHandler{},
		HearsHandlers: map[*regexp.Regexp][]hears.Hearer{},
		httpm:         http.NewServeMux(),
		burl:          &url.URL{Path: "/" + name},
		store:         memory.New(),
	}

	for _, opt := range opts {
		opt(mx)
	}

	return mx
}

// WithStore is a Mux option to set the store to be used
// for managing aliases, and property lookup.
func WithStore(s storage.Storer) Opt {
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

	for _, h := range mx.BGHandlers {
		go h.StartBackground(ctx, w.Copy())
	}
}

// SetAdapter sets the adapter on all the webhook of this mux.
func (mx *Mux) SetAdapter(a hugot.Adapter) {
	mx.Lock()
	defer mx.Unlock()

	for _, wh := range mx.Webhooks {
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
	for _, rh := range mx.RawHandlers {
		nm := m.Copy()
		go rh.ProcessMessage(ctx, w, nm)
	}

	if m.ToBot && m.Text != "" {
		nm := m.Copy()
		err = mx.ToBot.ProcessMessage(ctx, w, nm)
	}

	if err == command.ErrSkipHears {
		return nil
	}

	for _, hhs := range mx.HearsHandlers {
		for _, hh := range hhs {
			if ms := hh.Hears().FindAllStringSubmatch(m.Text, -1); ms != nil {
				nm := m.Copy()
				hn, _ := hh.Describe()
				nm.Store = prefix.New(mx.store, []string{hn})
				err = hh.Heard(ctx, w, nm, ms)
			}
		}
	}

	return err
}

// Raw adds the provided handlers to the Mux. All
// messages sent to the mux will be forwarded to this handler.
func (mx *Mux) Raw(hs ...hugot.Handler) error {
	mx.Lock()
	defer mx.Unlock()

	mx.RawHandlers = append(mx.RawHandlers, hs...)

	return nil
}

// Background adds the provided handler to the Mux. It
// will be started with the Mux is started.
func (mx *Mux) Background(hs ...hugot.BackgroundHandler) error {
	mx.Lock()
	defer mx.Unlock()

	mx.BGHandlers = append(mx.BGHandlers, hs...)

	return nil
}

// Hears adds the provided handler to the mux. all
// messages matching the hears patterns will be forwarded to
// the handler.
func (mx *Mux) Hears(hs ...hears.Hearer) error {
	mx.Lock()
	defer mx.Unlock()

	for _, h := range hs {
		r := h.Hears()
		mx.HearsHandlers[r] = append(mx.HearsHandlers[r], h)
	}
	return nil
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
	mx.Webhooks[n] = h

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

// URL returns the base URL for this Mux
func (mx *Mux) URL() *url.URL {
	mx.RLock()
	defer mx.RUnlock()
	return mx.url()
}

func (mx *Mux) url() *url.URL {
	return mx.burl
}

// SetURL sets the base URL for this mux's web hooks.
func (mx *Mux) SetURL(b *url.URL) {
	mx.Lock()
	defer mx.Unlock()

	mx.burl = b
	for _, h := range mx.Webhooks {
		n, _ := h.Describe()
		p := fmt.Sprintf("%s/%s/%s/", b.Path, mx.name, n)
		nu := *b
		nu.Path = p
		h.SetURL(&nu)
	}
}

// Help will send help about all the handlers in the mux to the user
func (mx *Mux) Help(ctx context.Context, w io.Writer, m *command.Message) error {
	if len(m.Args()) > 0 {
		return mx.cmdHelp(ctx, w, m)
	}

	out := &bytes.Buffer{}
	tw := new(tabwriter.Writer)
	tw.Init(out, 0, 8, 1, '\t', 0)

	if hh, ok := mx.ToBot.(help.Helper); ok {
		hh.Help(ctx, w, m)
	}

	if len(mx.HearsHandlers) > 0 {
		fmt.Fprintf(out, "Active hear handlers are patternss are:\n")
		for r, hs := range mx.HearsHandlers {
			for _, h := range hs {
				n, d := h.Describe()
				fmt.Fprintf(tw, "  %s\t`%s`\t - %s\n", n, r.String(), d)
			}
		}
		tw.Flush()
	}

	if len(mx.BGHandlers) > 0 {
		fmt.Fprintf(out, "Active background handlers are:\n")
		for _, h := range mx.BGHandlers {
			n, d := h.Describe()
			fmt.Fprintf(tw, "  %s\t - %s\n", n, d)
		}
		tw.Flush()
	}

	if len(mx.RawHandlers) > 0 {
		fmt.Fprintf(out, "Active raw handlers are:\n")
		for _, h := range mx.RawHandlers {
			n, d := h.Describe()
			fmt.Fprintf(tw, "  %s\t - %s\n", n, d)
		}
		tw.Flush()
	}

	io.Copy(w, out)
	return nil
}

func (mx *Mux) cmdHelp(ctx context.Context, w io.Writer, m *command.Message) error {
	cmd := m.Arg(1)

	allhs := []hugot.Describer{}
	for _, h := range mx.RawHandlers {
		allhs = append(allhs, h)
	}
	for _, h := range mx.BGHandlers {
		allhs = append(allhs, h)
	}

	for _, h := range allhs {
		hn, desc := h.Describe()
		if hn == cmd {
			if hh, ok := h.(help.Helper); ok {
				m.SetArgs(m.Args())
				return hh.Help(ctx, w, m)
			}
			fmt.Fprintln(w, desc)
		}
	}

	if hh, ok := mx.ToBot.(help.Helper); ok {
		m.SetArgs(m.Args())
		return hh.Help(ctx, w, m)
	}

	return nil
}
