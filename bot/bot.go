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

package bot

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/golang/glog"
	"github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/handlers/command"
	"github.com/tcolgate/hugot/handlers/hears"
	"github.com/tcolgate/hugot/handlers/help"
	"github.com/tcolgate/hugot/handlers/mux"
	"github.com/tcolgate/hugot/storage"
	"github.com/tcolgate/hugot/storage/memory"
	"github.com/tcolgate/hugot/storage/prefix"

	"golang.org/x/sync/errgroup"
)

// DefaultBot is a default instance of a bot
var DefaultBot *Bot
var DefaultMux *mux.Mux
var DefaultCommands command.Set

func init() {

	http.Handle("/hugot", DefaultMux)
	http.Handle("/hugot/", DefaultMux)

	DefaultCommands = command.NewSet()

	DefaultMux = mux.New("hugot", "")
	DefaultMux.ToBot = DefaultCommands

	DefaultCommands.MustAdd(help.New(DefaultMux))

	DefaultBot = New()
	DefaultBot.Store = memory.New()

}

type Bot struct {
	Store storage.Storer
}

func New() *Bot {
	b := &Bot{}
	return b
}

// ListenAndServe
func ListenAndServe(ctx context.Context, h hugot.Handler, a hugot.Adapter, as ...hugot.Adapter) {
	DefaultBot.ListenAndServe(ctx, h, a, as...)
}

// ListenAndServe runs the handler h, passing all messages to/from
// the provided adapter. The context may be used to gracefully shut
// down the server.
func (b *Bot) ListenAndServe(ctx context.Context, h hugot.Handler, a hugot.Adapter, as ...hugot.Adapter) {
	ctx = hugot.NewAdapterContext(ctx, a)

	if h == nil {
		h = DefaultMux
	}

	an := fmt.Sprintf("%T", a)
	if bh, ok := h.(hugot.BackgroundHandler); ok {
		runBackgroundHandler(ctx, bh, hugot.NewResponseWriter(a, hugot.Message{}, an))
	}

	if wh, ok := h.(hugot.WebHookHandler); ok {
		wh.SetAdapter(a)
	}

	type smrw struct {
		w hugot.ResponseWriter
		m *hugot.Message
	}
	mrws := make(chan smrw)

	g, ctx := errgroup.WithContext(ctx)

	for _, a := range append(as, a) {
		a := a
		g.Go(func() error {
			an := fmt.Sprintf("%T", a)
			for {
				select {
				case m := <-a.Receive():
					if m == nil {
						return io.EOF
					}
					rw := hugot.NewResponseWriter(a, *m, an)
					mrws <- smrw{rw, m}
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		})
	}

	go func() {
		g.Wait()
	}()

	hn, _ := h.Describe()
	for {
		select {
		case mrw := <-mrws:
			mrw.m.Store = prefix.New(b.Store, []string{hn})
			go func(smrw) {
				if err := h.ProcessMessage(ctx, mrw.w, mrw.m); err != nil {
					mrw.w.Send(ctx, mrw.m.Replyf("%v\n", err))
				}
			}(mrw)

		case <-ctx.Done():
			return
		}
	}
}

// runBackgroundHandler starts the provided BackgroundHandler in a new
// go routine.
func runBackgroundHandler(ctx context.Context, h hugot.BackgroundHandler, w hugot.ResponseWriter) {
	glog.Infof("Starting background %v\n", h)
	go func(ctx context.Context, bh hugot.BackgroundHandler) {
		h.StartBackground(ctx, w)
	}(ctx, h)
}

// Raw adds the provided handler to the DefaultMux
func Raw(hs ...hugot.Handler) error {
	return DefaultMux.Raw(hs...)
}

// Background adds the provided handler to the DefaultMux
func Background(hs ...hugot.BackgroundHandler) error {
	return DefaultMux.Background(hs...)
}

// Hears adds the provided handler to the DefaultMux
func Hears(hs ...hears.Hearer) error {
	return DefaultMux.Hears(hs...)
}

// HandleHTTP adds the provided handler to the DefaultMux
func HandleHTTP(h hugot.WebHookHandler) {
	DefaultMux.HandleHTTP(h)
}

// URL returns the base URL for the default Mux
func URL() *url.URL {
	return DefaultMux.URL()
}

// SetURL sets the base URL for web hooks.
func SetURL(b *url.URL) {
	if b.Path != "" {
		panic(errors.New("Can't set URL with path at the moment, sorry"))
	}
	DefaultMux.SetURL(b)
}

func Command(c command.Commander) {
	DefaultCommands.MustAdd(c)
}

func Commands() command.Set {
	return DefaultCommands
}
