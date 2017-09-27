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
	"github.com/tcolgate/hugot/handlers/mux"
	"github.com/tcolgate/hugot/storage"
	"github.com/tcolgate/hugot/storage/memory"
	"github.com/tcolgate/hugot/storage/prefix"

	"golang.org/x/sync/errgroup"
)

// DefaultBot is a default instance of a bot
var DefaultBot *Bot

func init() {
	DefaultBot = New()
	DefaultBot.Mux = mux.New("hugot", "")

	http.Handle("/hugot", DefaultBot.Mux)
	http.Handle("/hugot/", DefaultBot.Mux)

	DefaultBot.Commands = command.Set{}

	DefaultBot.Mux.ToBot = DefaultBot.Commands
	DefaultBot.Commands.MustAdd(DefaultBot.Mux)

	DefaultBot.Store = memory.New()
}

// Bot is the main type for implementing chat bots. They listen on one or more
// adapters and pass messages to, and from handlers.
type Bot struct {
	Store    storage.Storer
	Mux      *mux.Mux
	Commands command.Set
}

// New creates a new bot.
func New() *Bot {
	b := &Bot{}
	return b
}

// ListenAndServe runs the DefaultBot handler loop.
func ListenAndServe(ctx context.Context, h hugot.Handler, a hugot.Adapter, as ...hugot.Adapter) {
	DefaultBot.ListenAndServe(ctx, h, a, as...)
}

// ListenAndServe runs the handler h, passing all messages to/from
// the provided adapter. The context may be used to gracefully shut
// down the server.
func (b *Bot) ListenAndServe(ctx context.Context, h hugot.Handler, a hugot.Adapter, as ...hugot.Adapter) {
	ctx = hugot.NewAdapterContext(ctx, a)

	if h == nil {
		h = b.Mux
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

// Raw adds the provided handler to the Mux of the DefaultBot.
func Raw(hs ...hugot.Handler) error {
	return DefaultBot.Mux.Raw(hs...)
}

// Background adds the provided handler to the Mux of the DefaultBot.
func Background(hs ...hugot.BackgroundHandler) error {
	return DefaultBot.Mux.Background(hs...)
}

// Hears adds the provided handle to the Mux of the DefaultBotr.
func Hears(hs ...hears.Hearer) error {
	return DefaultBot.Mux.Hears(hs...)
}

// HandleHTTP adds the provided handler to the Mux of the DefaultBot.
func HandleHTTP(h hugot.WebHookHandler) {
	DefaultBot.Mux.HandleHTTP(h)
}

// URL returns the base URL for the Mux of the DefaultBot
func URL() *url.URL {
	return DefaultBot.Mux.URL()
}

// SetURL sets the base URL for web hooks on the DefaultBot.
func SetURL(b *url.URL) {
	if b.Path != "" {
		panic(errors.New("Can't set URL with path at the moment, sorry"))
	}
	DefaultBot.Mux.SetURL(b)
}

// Command adds a CLI command to the DefaultBot
func Command(c *command.Handler) {
	DefaultBot.Commands.MustAdd(c)
}

// Command adds a CLI command to the Bot
func (b *Bot) Command(c *command.Handler) {
	b.Commands.MustAdd(c)
}
