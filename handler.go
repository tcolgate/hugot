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
	"context"
	"errors"
	"regexp"
	"runtime/debug"

	"github.com/golang/glog"
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

// RawHandler will recieve every message  sent to the handler, without
// any filtering.
type RawHandler interface {
	Handler
	Handle(ctx context.Context, s Sender, m *Message) error
}

// BackgroundHandler gets run when the bot starts listening. They are
// intended for publishing messages that are not in response to any
// specific incoming message.
type BackgroundHandler interface {
	Handler
	BackgroundHandle(ctx context.Context, s Sender)
}

// HearsHandler is a handler which responds to messages matching a specific
// pattern.
type HearsHandler interface {
	Handler
	Hears() *regexp.Regexp                                                  // Returns the regexp we want to hear
	Heard(ctx context.Context, s Sender, m *Message, submatches [][]string) // Called once a message matches, and is passed any submatches from the regexp capture groups
}

// HearsHandler handlers are used to implement CLI style commands
type CommandHandler interface {
	Handler
	Command(ctx context.Context, s Sender, m *Message) error
}

func glogPanic() {
	err := recover()
	if err != nil {
		glog.Error(err)
		glog.Error(string(debug.Stack()))
	}
}

func runHandlers(ctx context.Context, a SenderReceiver, h Handler) {
	if bh, ok := h.(BackgroundHandler); ok {
		runBGHandler(ctx, a, bh)
	}

	for {
		select {
		case m := <-a.Receive():
			glog.Infoln(m)
			m.SenderReceiver = a

			if rh, ok := h.(RawHandler); ok {
				go runRawHandler(ctx, rh, m)
			}

			if hh, ok := h.(HearsHandler); ok {
				go runHearsHandler(ctx, hh, m)
			}
		case <-ctx.Done():
			return
		}
	}
}

func runOneHandler(ctx context.Context, h Handler, m *Message) {
	if rh, ok := h.(RawHandler); ok {
		go runRawHandler(ctx, rh, m)
	}

	if hh, ok := h.(HearsHandler); ok {
		go runHearsHandler(ctx, hh, m)
	}

	if hh, ok := h.(CommandHandler); ok {
		go runCommandHandler(ctx, hh, m)
	}
}

func runBGHandler(ctx context.Context, s Sender, h BackgroundHandler) {
	glog.Infof("Starting background %v\n", h)
	go func(ctx context.Context, bh BackgroundHandler) {
		defer glogPanic()
		h.BackgroundHandle(ctx, s)
	}(ctx, h)
}

func runRawHandler(ctx context.Context, h RawHandler, m *Message) {
	defer glogPanic()

	h.Handle(ctx, m.SenderReceiver, m)
}

func runHearsHandler(ctx context.Context, h HearsHandler, m *Message) bool {
	defer glogPanic()

	if mtchs := h.Hears().FindAllStringSubmatch(m.Text, -1); mtchs != nil {
		h.Heard(ctx, m.SenderReceiver, m, mtchs)
		return true
	}
	return false
}

func runCommandHandler(ctx context.Context, h CommandHandler, m *Message) {
	defer glogPanic()

	h.Command(ctx, m.SenderReceiver, m)
}
