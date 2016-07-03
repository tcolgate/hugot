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
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"runtime/debug"

	"golang.org/x/net/context"

	"github.com/golang/glog"
	"github.com/mattn/go-shellwords"
)

var (
	// ErrSkipHears is returned if the command wishes any
	// following hear handlers to be skipped, (e.g used for
	// help messages.
	ErrSkipHears = errors.New("skip hear messages")

	// ErrNextCommand is returned if the command wishes the message
	// to be passed to one of the SubCommands.
	ErrNextCommand = errors.New("pass this to the next command")

	ErrUnknownCommand = errors.New("unknown command")
)

type ErrUsage struct {
	string
}

func (e ErrUsage) Error() string {
	return e.string
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
	nmsg := w.msg
	nmsg.Text = string(bs)
	w.snd.Send(context.TODO(), &nmsg)
	return len(bs), nil
}

func (w *responseWriter) SetChannel(s string) {
	w.msg.Channel = s
}

func (w *responseWriter) SetTo(s string) {
	w.msg.To = s
}

func (w *responseWriter) Send(ctx context.Context, m *Message) {
	w.snd.Send(ctx, m)
}

type nullSender struct {
}

func (nullSender) Send(ctx context.Context, m *Message) {
}

func NewNullResponseWriter(m Message) ResponseWriter {
	return newResponseWriter(nullSender{}, m)
}

type baseHandler struct {
	name string
	desc string
}

func (bh *baseHandler) Describe() (string, string) {
	return bh.name, bh.desc
}

func newBaseHandler(name, desc string) Handler {
	return &baseHandler{name, desc}
}

type RawFunc func(ctx context.Context, w ResponseWriter, m *Message) error

// RawHandler will recieve every message  sent to the handler, without
// any filtering.
type RawHandler interface {
	Handler
	Handle(ctx context.Context, w ResponseWriter, m *Message) error
}

type baseRawHandler struct {
	Handler
	rhf RawFunc
}

func NewRawHandler(name, desc string, f RawFunc) RawHandler {
	return &baseRawHandler{
		Handler: newBaseHandler(name, desc),
		rhf:     f,
	}
}

func (brh *baseRawHandler) Handle(ctx context.Context, w ResponseWriter, m *Message) error {
	return brh.rhf(ctx, w, m)
}

// BackgroundFunc
type BackgroundFunc func(ctx context.Context, w ResponseWriter)

// BackgroundHandler gets run when the bot starts listening. They are
// intended for publishing messages that are not in response to any
// specific incoming message.
type BackgroundHandler interface {
	Handler
	StartBackground(ctx context.Context, w ResponseWriter)
}

type baseBackgroundHandler struct {
	Handler
	bhf BackgroundFunc
}

func NewBackgroundHandler(name, desc string, f BackgroundFunc) BackgroundHandler {
	return &baseBackgroundHandler{
		Handler: newBaseHandler(name, desc),
		bhf:     f,
	}
}

func (bbh *baseBackgroundHandler) StartBackground(ctx context.Context, w ResponseWriter) {
	bbh.bhf(ctx, w)
}

// HeardFunc
type HeardFunc func(ctx context.Context, w ResponseWriter, m *Message, submatches [][]string) // Called once a message matches, and is passed any submatches from the regexp capture groups

// HearsHandler is a handler which responds to messages matching a specific
// pattern.
type HearsHandler interface {
	Handler
	Hears() *regexp.Regexp                                                          // Returns the regexp we want to hear
	Heard(ctx context.Context, w ResponseWriter, m *Message, submatches [][]string) // Called once a message matches, and is passed any submatches from the regexp capture groups
}

type baseHearsHandler struct {
	Handler
	rgxp *regexp.Regexp
	hhf  HeardFunc
}

func NewHearsHandler(name, desc string, rgxp *regexp.Regexp, f HeardFunc) HearsHandler {
	return &baseHearsHandler{
		Handler: newBaseHandler(name, desc),
		rgxp:    rgxp,
		hhf:     f,
	}
}

func (bhh *baseHearsHandler) Hears() *regexp.Regexp {
	return bhh.rgxp
}

func (bhh *baseHearsHandler) Heard(ctx context.Context, w ResponseWriter, m *Message, submatches [][]string) {
	bhh.hhf(ctx, w, m, submatches)
}

// CommandFunc
type CommandFunc func(ctx context.Context, w ResponseWriter, m *Message) error

// CommandHandler handlers are used to implement CLI style commands
type CommandHandler interface {
	Handler
	Command(ctx context.Context, w ResponseWriter, m *Message) error
}

type baseCommandHandler struct {
	Handler
	bcf CommandFunc
}

func NewCommandHandler(name, desc string, f CommandFunc) CommandHandler {
	return &baseCommandHandler{
		Handler: newBaseHandler(name, desc),
		bcf:     f,
	}
}

func (bch *baseCommandHandler) Command(ctx context.Context, w ResponseWriter, m *Message) error {
	return bch.bcf(ctx, w, m)
}

// CommandHandler handlers are used to implement CLI style commands
type HTTPHandler interface {
	Handler
	http.Handler
}

type baseHTTPHandler struct {
	Handler
	httph http.Handler
}

func NewHTTPHandler(name, desc string, h http.Handler) HTTPHandler {
	return &baseHTTPHandler{
		Handler: newBaseHandler(name, desc),
		httph:   h,
	}
}

func (bwh *baseHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bwh.httph.ServeHTTP(w, r)
}

func glogPanic() {
	err := recover()
	if err != nil && err != ErrNextCommand && err != flag.ErrHelp {
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
		h.StartBackground(ctx, w)
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
	if h != nil && glog.V(2) {
		glog.Infof("RUNNING %v %v\n", h, m.args)
	}
	defer glogPanic()
	var err error

	if m.args == nil {
		m.args, err = shellwords.Parse(m.Text)
		if err != nil {
			return err
		}
	}

	if len(m.args) == 0 {
		//nothing to do.
		return errors.New("command handler called with no possible arguments")
	}

	name := m.args[0]
	m.flagOut = &bytes.Buffer{}
	m.FlagSet = flag.NewFlagSet(name, flag.ContinueOnError)
	m.FlagSet.SetOutput(m.flagOut)

	err = h.Command(ctx, w, m)
	if err == flag.ErrHelp {
		fmt.Fprint(w, cmdUsage(h, name, nil).Error())
		return ErrSkipHears
	}

	return err
}
