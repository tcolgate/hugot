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
	"fmt"

	"github.com/nlopes/slack"
)

// Message describes a Message from or to a user. It is intended to
// provided a resonable lowest common denominator for modern chat systems.
// It takes the Slack message format to provide that minimum but makes no
// assumption about support for any markup.
// If used within a command handler, the message can also be used as a flag.FlagSet
// for adding and processing the message as a CLI command.
type Message struct {
	To      string
	From    string
	Channel string

	UserID string // Verified user identitify within the source adapter

	Text        string // A plain text message
	Attachments []Attachment

	Private bool
	ToBot   bool

	Store Storer
}

// Copy is used to provide a deep copy of a message
func (m *Message) Copy() *Message {
	nm := *m
	copy(nm.Attachments, m.Attachments)
	return &nm
}

// Attachment represents a rich message attachment and is directly
// modeled on the Slack attachments API
type Attachment slack.Attachment

// Reply returns a messsage with Text tx and the From and To fields switched
func (m *Message) Reply(txt string) *Message {
	out := *m
	out.Text = txt

	out.From = ""
	out.To = m.From

	return &out
}

// Replyf returns message with txt set to the fmt.Printf style formatting,
// and the from/to fields switched.
func (m *Message) Replyf(s string, is ...interface{}) *Message {
	return m.Reply(fmt.Sprintf(s, is...))
}

// Properties are used to associate scoped key/value data
// with a message
func (m *Message) Properties() PropertyStore {
	return NewPropertyStore(NewPrefixedStore(m.Store, []string{"props"}), m)
}
