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

package message

import (
	"flag"
	"fmt"

	"github.com/nlopes/slack"
)

type Message struct {
	Event     *slack.MessageEvent
	From      *slack.User
	ChannelID string
	Channel   *slack.Channel

	Text        string
	Attachments []slack.Attachment

	Private bool
	ToBot   bool

	*flag.FlagSet
}

func (m *Message) Reply(txt string) *Message {
	out := *m
	out.Text = txt

	if !m.Private && m.ToBot {
		out.Text = fmt.Sprintf("@%s: %s", m.From.Name, txt)
	}

	out.Event = nil
	out.From = nil

	return &out
}

func (m *Message) Replyf(s string, is ...interface{}) *Message {
	return m.Reply(fmt.Sprintf(s, is...))
}
