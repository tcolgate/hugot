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
	"bytes"
	"context"
	"errors"
	"flag"
	"io"
	"regexp"
	"runtime/debug"

	"github.com/golang/glog"
	"github.com/mattn/go-shellwords"
)

var (
	ErrAskNicely    = errors.New("potentially dangerous, ask nicely")
	ErrUnAuthorized = errors.New("you are not authorized to perform this action")
	ErrNeedsPrivacy = errors.New("potentially dangerous, ask me in private")
)

// Describer can return a name and description.
type Describer interface {
	Describe() (string, string)
}

// Handler is a handler with no actual functionality
type Handler interface {
	Describer
}

type ResponseWriter interface {
	Sender
	io.Writer

	// Force this message to a certain channel
	SetChannel(s string)

	// Force this message to a certain user
	SetTo(to string)
}

type responseWriter struct {
	snd Sender
	msg Message
}

func newResponseWriter(s Sender, m Message) ResponseWriter {
	return &responseWriter{s, m}
}

func (w *responseWriter) Write(bs []byte) (int, error) {
	w.msg.Text = string(bs)
	w.snd.Send(context.TODO(), &w.msg)
	return len(bs), nil
}

func (w *responseWriter) SetChannel(s string) {
	w.msg.To = s
}

func (w *responseWriter) SetTo(s string) {
	w.msg.Channel = s
}

func (w *responseWriter) Send(ctx context.Context, m *Message) {
	w.snd.Send(ctx, m)
}

// RawHandler will recieve every message  sent to the handler, without
// any filtering.
type RawHandler interface {
	Handler
	Handle(ctx context.Context, w ResponseWriter, m *Message) error
}

// BackgroundHandler gets run when the bot starts listening. They are
// intended for publishing messages that are not in response to any
// specific incoming message.
type BackgroundHandler interface {
	Handler
	BackgroundHandle(ctx context.Context, w ResponseWriter)
}

// HearsHandler is a handler which responds to messages matching a specific
// pattern.
type HearsHandler interface {
	Handler
	Hears() *regexp.Regexp                                                          // Returns the regexp we want to hear
	Heard(ctx context.Context, w ResponseWriter, m *Message, submatches [][]string) // Called once a message matches, and is passed any submatches from the regexp capture groups
}

// HearsHandler handlers are used to implement CLI style commands
type CommandHandler interface {
	Handler
	Command(ctx context.Context, w ResponseWriter, m *Message) error
}

type SubCommandHandler interface {
	CommandHandler
	SubCommands() map[string]CommandHandler
}

func glogPanic() {
	err := recover()
	if err != nil {
		glog.Error(err)
		glog.Error(string(debug.Stack()))
	}
}

func runHandlers(ctx context.Context, a Adapter, h Handler) {
	if bh, ok := h.(BackgroundHandler); ok {
		runBGHandler(ctx, newResponseWriter(a, Message{}), bh)
	}

	for {
		select {
		case m := <-a.Receive():
			if rh, ok := h.(RawHandler); ok {
				go runRawHandler(ctx, newResponseWriter(a, *m), rh, m)
			}

			if hh, ok := h.(HearsHandler); ok {
				go runHearsHandler(ctx, newResponseWriter(a, *m), hh, m)
			}
		case <-ctx.Done():
			return
		}
	}
}

func runOneHandler(ctx context.Context, w ResponseWriter, h Handler, m *Message) {
	if rh, ok := h.(RawHandler); ok {
		go runRawHandler(ctx, w, rh, m)
	}

	if hh, ok := h.(HearsHandler); ok {
		go runHearsHandler(ctx, w, hh, m)
	}

	if hh, ok := h.(CommandHandler); ok {
		go runCommandHandler(ctx, w, hh, m)
	}
}

func runBGHandler(ctx context.Context, w ResponseWriter, h BackgroundHandler) {
	glog.Infof("Starting background %v\n", h)
	go func(ctx context.Context, bh BackgroundHandler) {
		defer glogPanic()
		h.BackgroundHandle(ctx, w)
	}(ctx, h)
}

func runRawHandler(ctx context.Context, w ResponseWriter, h RawHandler, m *Message) {
	defer glogPanic()

	h.Handle(ctx, w, m)
}

func runHearsHandler(ctx context.Context, w ResponseWriter, h HearsHandler, m *Message) bool {
	defer glogPanic()

	if mtchs := h.Hears().FindAllStringSubmatch(m.Text, -1); mtchs != nil {
		h.Heard(ctx, w, m, mtchs)
		return true
	}
	return false
}

func runCommandHandler(ctx context.Context, w ResponseWriter, h CommandHandler, m *Message) {
	defer glogPanic()
	var err error

	if m.args == nil {
		m.args, err = shellwords.Parse(m.Text)
		if err != nil {
			w.Send(ctx, m.Reply("Could not parse as command line, "+err.Error()))
		}
	}

	m.flagOut = &bytes.Buffer{}
	m.FlagSet = flag.NewFlagSet(m.args[0], flag.ContinueOnError)
	m.FlagSet.SetOutput(m.flagOut)

	err = h.Command(ctx, w, m)

	switch err {
	case nil:
	case ErrNextCommand:
	case ErrAskNicely:
		w.Send(ctx, m.Reply("You should ask Nicely"))
	case ErrUnAuthorized:
		w.Send(ctx, m.Reply("You are not authorized to do that"))
	case ErrNeedsPrivacy:
		w.Send(ctx, m.Reply("You should ask that in private"))
	default:
		w.Send(ctx, m.Replyf("error, %v", err.Error()))
	}
}
