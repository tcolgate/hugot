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
)

const adapterKey key = 0

// NewAdapterContext creates a context for passing an adapter. This is
// mainly used by web handlers.
func NewAdapterContext(ctx context.Context, a Adapter) context.Context {
	return context.WithValue(ctx, adapterKey, a)
}

// AdapterFromContext returns the Adapter stored in a context.
func AdapterFromContext(ctx context.Context) (Adapter, bool) {
	a, ok := ctx.Value(adapterKey).(Adapter)
	return a, ok
}

// SenderFromContext can be used to retrieve a valid sender from
// a context. This is mostly useful in WebHook handlers for sneding
// messages back to the inbound Adapter.
func SenderFromContext(ctx context.Context) (Sender, bool) {
	a, ok := ctx.Value(adapterKey).(Adapter)
	return a, ok
}

// Receiver cam ne used to receive messages
type Receiver interface {
	Receive() <-chan *Message // Receive returns a channel that can be used to read one message, nil indicated there will be no more messages
}

// Sender can be used to send messages
type Sender interface {
	Send(ctx context.Context, m *Message)
}

// TextOnly is an interface to hint to handlers that the adapter
// they are talking to is a text only handler, to help adjust
// output.
type TextOnly interface {
	Sender
	IsTextOnly()
}

// IsTextOnly returns true if the sender only support text.
func IsTextOnly(s Sender) bool {
	_, ok := s.(TextOnly)
	return ok
}

// User represents a user within the chat sytems. The adapter is responsible
// for translating the string User to and from it's external representation
type User string

// Channel represents discussion channel, such as an IRC channel or
// Slack channel. The Adapter is responsible for translating between a human
// name for the channel, and any internal representation
type Channel string

// Adapter can be used to communicate with an external chat system such as
// slack or IRC.
type Adapter interface {
	Sender
	Receiver
}
