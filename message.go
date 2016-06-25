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
	"log"

	"github.com/mattn/go-shellwords"
	"github.com/nlopes/slack"
)

type Attachment slack.Attachment

type Message struct {
	Event   *slack.MessageEvent
	To      string
	From    string
	Channel string

	Text        string
	Attachments []Attachment

	Private bool
	ToBot   bool

	*flag.FlagSet

	args    []string
	flagOut *bytes.Buffer
}

var ErrBadCLI = errors.New("coul not process as command line")

func (m *Message) Reply(txt string) *Message {
	out := *m
	out.Text = txt

	out.Event = nil
	out.From = ""
	out.To = m.From

	return &out
}

func (m *Message) Replyf(s string, is ...interface{}) *Message {
	return m.Reply(fmt.Sprintf(s, is...))
}

func (m *Message) NewFlagSet() error {
	return nil
}

func (m *Message) Parse() error {
	args, err := shellwords.Parse(m.Text)
	if err != nil {
		return err
	}
	log.Println(args)

	return m.FlagSet.Parse(args[1:])
}
