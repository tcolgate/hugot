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

// Package irc implements a simple adapter for IRC using
// github.com/thoj/go-ircevent
package irc

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/context"

	"github.com/golang/glog"
	"github.com/tcolgate/hugot"
	irce "github.com/thoj/go-ircevent"
)

// New creates a new adapter that communicates with an IRC server using
// github.com/thoj/go-ircevent
func New(i *irce.Connection) (hugot.Adapter, error) {
	a := &irc{i, make(chan *hugot.Message), regexp.MustCompile(fmt.Sprintf("^%s[:, ]?(.*)", i.GetNick()))}

	i.AddCallback("PRIVMSG", a.gotEvent)
	return a, nil
}

type irc struct {
	*irce.Connection
	c   chan *hugot.Message
	dir *regexp.Regexp
}

func (i *irc) gotEvent(e *irce.Event) {
	go func() {
		if glog.V(3) {
			glog.Infof("Got %#v", *e)
		}
		i.c <- i.eventToHugot(e)
	}()
}

func (irc *irc) Send(ctx context.Context, m *hugot.Message) {
	if m.Private {
		if m.Channel == "" {
			if m.To != "" {
				m.Channel = m.To
			} else {
				m.Channel = m.From
			}
		}
	}
	if glog.V(3) {
		glog.Infof("Sending %#v", *m)
	}
	for _, l := range strings.Split(m.Text, "\n") {
		irc.Privmsg(m.Channel, l)
	}
}

func (irc *irc) Receive() <-chan *hugot.Message {
	return irc.c
}

func (irc *irc) eventToHugot(e *irce.Event) *hugot.Message {
	txt := e.Message()
	tobot := false
	priv := false

	channel := strings.Split(e.Raw, " ")[2] // either GetNick() or #mychannel

	// Check if the message was sent @bot, if so, set it as to us
	// and strip the leading politeness
	dirMatch := irc.dir.FindStringSubmatch(txt)
	if glog.V(3) {
		glog.Infof("Match %#v", dirMatch)
	}

	if len(dirMatch) > 1 {
		tobot = true
		txt = strings.Trim(dirMatch[1], " ")
	}

	if channel == irc.GetNick() {
		channel = strings.Split(e.Raw[1:], "!")[0] // Starting from the [1]st char, split at "!", store as string
		tobot = true
		priv = true
	}

	return &hugot.Message{
		Channel: channel,
		From:    e.Nick,
		Text:    txt,
		ToBot:   tobot,
		Private: priv,
	}
}
