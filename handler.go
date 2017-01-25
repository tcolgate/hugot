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
	"flag"
	"fmt"
	"io"
	"runtime/debug"

	"context"

	"github.com/golang/glog"
)

// Describer returns the name and description of a handler. This
// is used to identify the handler within Command and HTTP Muxs,
// and to provide a descriptive name for the handler in help text.
type Describer interface {
	Describe() (string, string)
}

// HandlerFunc describes a function that can be used as a hugot handler.
type HandlerFunc func(ctx context.Context, w ResponseWriter, m *Message) error

// Handler is a handler with no actual functionality
type Handler interface {
	Describer
	ProcessMessage(ctx context.Context, w ResponseWriter, m *Message) error
}

// NewNullResponseWriter creates a ResponseWriter that discards all
// message sent to it.
func NewNullResponseWriter(m Message) ResponseWriter {
	return newResponseWriter(nullSender{}, m, "null")
}

// ResponseWriter is used to Send messages back to a user.
type ResponseWriter interface {
	Sender
	io.Writer

	SetChannel(c string) // Forces messages to a certain channel
	SetTo(to string)     // Forces messages to a certain user
	SetSender(a Sender)  // Forces messages to a different sender or adapter

	Copy() ResponseWriter // Returns a copy of this response writer
}

type responseWriter struct {
	snd Sender
	msg Message
	an  string
}

func newResponseWriter(s Sender, m Message, an string) ResponseWriter {
	return &responseWriter{s, m, an}
}

// ResponseWriterFromContext constructs a ResponseWriter from the adapter
// stored in the context. A destination Channel/User must be set to send
// messages..
func ResponseWriterFromContext(ctx context.Context) (ResponseWriter, bool) {
	s, ok := SenderFromContext(ctx)
	if !ok {
		return nil, false
	}
	an := fmt.Sprintf("%T", s)
	return newResponseWriter(s, Message{}, an), true
}

// Write implements the io.Writer interface. All writes create a single
// new message that is then sent to the ResoneWriter's current adapter
func (w *responseWriter) Write(bs []byte) (int, error) {
	nmsg := w.msg
	nmsg.Text = string(bs)
	w.Send(context.TODO(), &nmsg)
	return len(bs), nil
}

// SetChannel sets the outbound channel for message sent via this writer
func (w *responseWriter) SetChannel(s string) {
	w.msg.Channel = s
}

// SetChannel sets the target user for message sent via this writer
func (w *responseWriter) SetTo(s string) {
	w.msg.To = s
}

// SetSender sets the target adapter for sender sent via this writer
func (w *responseWriter) SetSender(s Sender) {
	w.snd = s
}

// Send implements the Sender interface
func (w *responseWriter) Send(ctx context.Context, m *Message) {
	messagesTx.WithLabelValues(w.an, m.Channel, m.From).Inc()
	w.snd.Send(ctx, m)
}

// Copy returns a copy of this response writer
func (w *responseWriter) Copy() ResponseWriter {
	return &responseWriter{w.snd, Message{}, w.an}
}

// nullSender is a sender which discards anything sent to it, this is
// useful for the help handler.
type nullSender struct {
}

// Send implements Send, and discards the message
func (nullSender) Send(ctx context.Context, m *Message) {
}

func glogPanic() {
	err := recover()
	if err != nil && err != flag.ErrHelp {
		glog.Error(err)
		glog.Error(string(debug.Stack()))
	}
}
