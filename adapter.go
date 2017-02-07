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

// Adapter can be used to communicate with an external chat system such as
// slack or IRC.
type Adapter interface {
	Sender
	Receiver
}
