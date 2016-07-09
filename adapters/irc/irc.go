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
// github.com/fluffle/goirc/client
package irc

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/fluffle/goirc/client"
	iglog "github.com/fluffle/goirc/logging/glog"
	"github.com/golang/glog"
	"github.com/tcolgate/hugot"
)

type irc struct {
	cfg      *client.Config
	defChans []string

	c chan *hugot.Message

	start sync.Once
	*client.Conn
}

// New creates a new adapter that communicates with an IRC server using
// github.com/thoj/go-ircevent
func New(c *client.Config, chans ...string) hugot.Adapter {
	a := &irc{
		c,
		chans,
		make(chan *hugot.Message),
		sync.Once{},
		nil,
	}

	return a
}

func (i *irc) Send(ctx context.Context, m *hugot.Message) {
	i.Start()
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
		i.Privmsg(m.Channel, l)
	}
}

func (i *irc) Receive() <-chan *hugot.Message {
	i.Start()
	return i.c
}

func (i *irc) Start() {
	i.start.Do(func() {
		go i.run()
	})
}

func (i *irc) run() {
	iglog.Init()
	for {
		glog.Info("In here")
		i.Conn = client.Client(i.cfg)

		disconnected := make(chan struct{})
		i.HandleFunc(client.DISCONNECTED, func(c *client.Conn, l *client.Line) {
			close(disconnected)
		})

		// Connect to an IRC server.
		if err := i.ConnectTo(i.cfg.Server); err != nil {
			glog.Errorf("could not connect to server, %v", err)
			<-time.After(5 * time.Second)
			continue
		}

		i.HandleFunc(client.PRIVMSG, func(conn *client.Conn, l *client.Line) {
			i.c <- i.eventToHugot(l)
		})

		i.HandleFunc(client.CONNECTED, func(conn *client.Conn, l *client.Line) {
			for _, c := range i.defChans {
				i.Join(c)
			}
		})

		// Wait for disconnection.
		<-disconnected
	}

}

func (i *irc) eventToHugot(l *client.Line) *hugot.Message {
	txt := l.Text()
	nick := i.Me().Nick
	tobot := false
	priv := false
	channel := l.Target()

	if l.Public() {
		// Check if the message was sent @bot, if so, set it as to us
		// and strip the leading politeness
		dir := regexp.MustCompile(fmt.Sprintf("^%s[:, ]+(.*)", nick))
		dirMatch := dir.FindStringSubmatch(txt)
		if glog.V(3) {
			glog.Infof("Match %#v", dirMatch)
		}

		if len(dirMatch) > 1 {
			tobot = true
			txt = strings.Trim(dirMatch[1], " ")
		}
	} else {
		tobot = true
		priv = true
	}

	return &hugot.Message{
		Channel: channel,
		From:    l.Nick,
		To:      nick,
		Text:    txt,
		ToBot:   tobot,
		Private: priv,
	}
}
