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

// Package slack implements an adapter for http://slack.com using
// github.com/slack-go/slack
package slack

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"context"

	"github.com/golang/glog"
	"github.com/tcolgate/hugot"

	client "github.com/slack-go/slack"
)

type slack struct {
	appToken string
	botToken string
	nick     string

	id   string
	icon string

	dirPat   *regexp.Regexp
	api      *client.Client
	info     client.Info
	users    []client.User
	channels []client.Channel
	*cache

	sender   chan *hugot.Message
	receiver chan client.RTMEvent
}

// New creates a new adapter that communicates with the Slack messaging
// API. A slack API token, and coresponding bot username must be provided
func New(token, nick string) (hugot.Adapter, error) {
	s := slack{nick: nick}
	if token == "" {
		return nil, errors.New("Slack Token must be set")
	}
	if s.nick == "" {
		return nil, errors.New("Slack Token must be set")
	}

	s.api = client.New(token)
	s.cache = newCache(s.api)

	s.info = client.Info{}
	s.users, _ = s.api.GetUsers()
	s.channels, _ = s.api.GetChannels(false)

	for _, u := range s.users {
		if u.Name == nick {
			s.id = u.ID
			s.icon = u.Profile.Image72
			break
		}
	}

	if s.id == "" {
		return nil, errors.New("Could not locate bot's user ID")
	}

	s.dirPat = regexp.MustCompile(fmt.Sprintf("(?m)^(!|(@?%s|<@%s>)[:,]? )(.*)", s.nick, s.id))

	// We use RTM to recieve, but the regular slack API to send
	// RTM does not support formatted message parsing
	wsAPI := s.api.NewRTM()
	s.receiver = wsAPI.IncomingEvents
	s.sender = make(chan *hugot.Message)

	go wsAPI.ManageConnection()

	return &s, nil
}

func (s *slack) Send(ctx context.Context, m *hugot.Message) {
	if (m.Text != "" || len(m.Attachments) > 0) && m.Channel != "" {
		var err error
		chanout := ""
		c, err := s.GetChannel(m.Channel)
		if err != nil {
			glog.Errorf("unresolvable channel, %#v", m.Channel)
			chanout = m.Channel
		} else {
			chanout = c.Name
		}
		if glog.V(3) {
			glog.Infof("sending, %#v to %#v", *m, chanout)
		}

		attchs := []client.Attachment{}
		for _, a := range m.Attachments {
			attchs = append(attchs, client.Attachment(a))
		}
		_, _, err = s.api.PostMessage(
			m.Channel,
			client.MsgOptionText(m.Text, false),
			client.MsgOptionAsUser(false),
			client.MsgOptionIconURL(s.icon),
			client.MsgOptionUsername(s.nick),
			client.MsgOptionAttachments(attchs...))
		if err != nil {
			glog.Errorf("error sending, %#v", err.Error())
		}
	} else {
		glog.Infoln("Attempt to send empty message")
	}
}

func (s *slack) Receive() <-chan *hugot.Message {
	out := make(chan *hugot.Message, 1)
	for {
		select {
		case m := <-s.receiver:
			switch ev := m.Data.(type) {
			case *client.HelloEvent:
			case *client.UserTypingEvent:
			case *client.ReconnectUrlEvent:
			case *client.ConnectingEvent:
				glog.Infof("Connecting")
			case *client.ConnectedEvent:
				glog.Infof("Connected")
			case *client.PresenceChangeEvent:
			case *client.LatencyReport:
				glog.Infof("Latency: %v", ev.Value)
			case *client.MessageEvent:
				m := s.slackMsgToHugot(ev)
				if m == nil {
					continue
				}
				out <- m
				return out
			default:
				glog.Infof("Unexpected: %T %v\n", m.Data, m.Data)
			}
		}
	}
}

func (s *slack) slackMsgToHugot(me *client.MessageEvent) *hugot.Message {
	var private, tobot bool
	if glog.V(3) {
		glog.Infof("message: %#v\n", *me)
	}

	var uname string

	if me.Username == "" {
		u, err := s.GetUser(me.User)
		if err == nil {
			uname = u.Name
		}
	} else {
		uname = me.Username
	}

	if uname == "" {
		glog.Infoln("could not resolve username")
		return nil
	}

	cname := me.Channel
	c := s.info.GetChannelByID(me.Channel)
	if c != nil {
		cname = c.Name
	}

	// ignore from self
	if me.User == s.id || uname == s.nick {
		return nil
	}

	txt := me.Msg.Text

	switch {
	case strings.HasPrefix(me.Channel, "D"):
		{ // One on one,
			private = true
			tobot = true
		}
	case strings.HasPrefix(me.Channel, "G"):
		{ // private group chat
			private = true
		}
	case strings.HasPrefix(me.Channel, "C"):
		{
			private = false
		}
	default:
		glog.Errorf("cannot determine channel type for %s", me.Channel)
		return nil
	}

	// Check if the message was sent @bot, if so, set it as to us
	// and strip the leading politeness
	dirMatch := s.dirPat.FindStringSubmatch(txt)
	if len(dirMatch) > 1 && len(dirMatch[1]) > 0 {
		tobot = true
		txt = strings.Trim(dirMatch[3], " ")
	}

	m := hugot.Message{
		Channel: cname,
		From:    uname,
		To:      "",
		UserID:  me.User,
		Private: private,
		ToBot:   tobot,
		Text:    txt,
	}

	if m.Private {
		glog.Infof("Handling private message from %v: %v", m.From, m.Text)
	} else {
		glog.Infof("Handling message in %v from %v: %v", m.Channel, m.From, m.Text)
	}

	return &m
}
