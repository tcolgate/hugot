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
	"fmt"
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

	// ErrIgnore
	ErrIgnored = errors.New("the handler ignored the message")

	// ErrSkipHears is returned if the command wishes any
	// following hear handlers to be skipped, (e.g used for
	// help messages.
	ErrSkipHears = errors.New("skip hear messages")

	// ErrNextCommand is returned if the command wishes the message
	// to be passed to one of the SubCommands.
	ErrNextCommand = errors.New("pass this to the next command")
)

type ErrUnknownCommand struct {
	Available []string
	Wanted    string
}

func (err ErrUnknownCommand) Error() string {
	return fmt.Sprintf("Unknown Command %s, available commands are %s",
		err.Wanted,
		err.Available)
}

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

func glogPanic() {
	err := recover()
	if err != nil {
		glog.Error(err)
		glog.Error(string(debug.Stack()))
	}
}

func RunHandlers(ctx context.Context, h Handler, a Adapter) {
	if bh, ok := h.(BackgroundHandler); ok {
		RunBackgroundHandler(ctx, bh, newResponseWriter(a, Message{}))
	}

	for {
		select {
		case m := <-a.Receive():
			if rh, ok := h.(RawHandler); ok {
				go RunRawHandler(ctx, rh, newResponseWriter(a, *m), m)
			}

			if hh, ok := h.(HearsHandler); ok {
				go RunHearsHandler(ctx, hh, newResponseWriter(a, *m), m)
			}

			if ch, ok := h.(CommandHandler); ok {
				go RunCommandHandler(ctx, ch, newResponseWriter(a, *m), m)
			}
		case <-ctx.Done():
			return
		}
	}
}

func RunBackgroundHandler(ctx context.Context, h BackgroundHandler, w ResponseWriter) {
	glog.Infof("Starting background %v\n", h)
	go func(ctx context.Context, bh BackgroundHandler) {
		defer glogPanic()
		h.BackgroundHandle(ctx, w)
	}(ctx, h)
}

func RunRawHandler(ctx context.Context, h RawHandler, w ResponseWriter, m *Message) bool {
	defer glogPanic()
	h.Handle(ctx, w, m)

	return false
}

func RunHearsHandler(ctx context.Context, h HearsHandler, w ResponseWriter, m *Message) bool {
	defer glogPanic()

	if mtchs := h.Hears().FindAllStringSubmatch(m.Text, -1); mtchs != nil {
		go h.Heard(ctx, w, m, mtchs)
		return true
	}
	return false
}

func RunCommandHandler(ctx context.Context, h CommandHandler, w ResponseWriter, m *Message) error {
	defer glogPanic()
	var err error

	if m.args == nil {
		m.args, err = shellwords.Parse(m.Text)
		if err != nil {
			w.Send(ctx, m.Reply("Could not parse as command line, "+err.Error()))
		}
	}

	if len(m.args) == 0 {
		//nothing to do.
		return nil
	}

	m.flagOut = &bytes.Buffer{}
	m.FlagSet = flag.NewFlagSet(m.args[0], flag.ContinueOnError)
	m.FlagSet.SetOutput(m.flagOut)

	return h.Command(ctx, w, m)
}
