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

package minion

import (
	"errors"
	"flag"
	"fmt"
	"regexp"
	"strings"

	"github.com/golang/glog"
	"github.com/tcolgate/hugot/handler"
	"github.com/tcolgate/hugot/message"
	"github.com/tcolgate/hugot/slackcache"

	"github.com/nlopes/slack"
)

type Bot struct {
	slackToken string
	nick       string

	id   string
	icon string

	dirPat *regexp.Regexp
	api    *slack.Client
	*slackcache.Cache

	Sender   chan *message.Message
	Receiver chan slack.RTMEvent
}

type option func(*Bot) error

func New(opts ...option) (b *Bot, err error) {
	b = &Bot{}
	for _, opt := range opts {
		if err = opt(b); err != nil {
			return nil, err
		}
	}

	if b.Token() == "" {
		return nil, errors.New("SlackToken must be set")
	}

	return b, nil
}

func Token(s string) option {
	return func(p *Bot) error {
		p.slackToken = s
		return nil
	}
}

func (b *Bot) Token() string {
	return b.slackToken
}

func Nick(s string) option {
	return func(p *Bot) error {
		p.nick = s
		return nil
	}
}

func (b *Bot) Nick() string {
	return b.nick
}

func (b *Bot) Start() error {
	if b.Token() == "" {
		return errors.New("Slack Token must be set")
	}

	for _, h := range handler.Handlers {
		go func(h handler.Handler) {
			defer func() {
				if r := recover(); r != nil {
					glog.Infof("%s Setup() paniced: %v", h.Names()[0], r)
				}
			}()
			glog.Infoln("calling setup for %s: %v", h.Names()[0], h.Setup)
			err := h.Setup()
			if err != nil {
				glog.Fatalf("Error in handler %s: %s", h.Names()[0], err.Error)
			}
		}(h)
	}

	b.api = slack.New(b.Token())
	if glog.V(3) {
		b.api.SetDebug(true)
	}
	b.Cache = slackcache.New(b.api)

	us, _ := b.api.GetUsers()

	for _, u := range us {
		if u.Name == b.nick {
			b.id = u.ID
			b.icon = u.Profile.Image72
		}
	}

	if b.id == "" {
		return errors.New("Could not locate bot's user ID")
	}

	b.dirPat = regexp.MustCompile(fmt.Sprintf("(?m)^(!|(@?%s|<@%s>)[:, ]?)(.*)", b.nick, b.id))

	b.run()
	return nil
}

func (b *Bot) run() {
	b.Sender = make(chan *message.Message)

	wsAPI := b.api.NewRTM()
	b.Receiver = wsAPI.IncomingEvents

	for _, h := range handler.Handlers {
		go h.Start(b.Sender)
	}

	go wsAPI.ManageConnection()
	go func(api *slack.Client, sender chan *message.Message) {
		for {
			select {
			case msg := <-sender:
				if (msg.Text != "" || len(msg.Attachments) > 0) && msg.ChannelID != "" {
					params := slack.NewPostMessageParameters()
					params.AsUser = false
					params.Attachments = msg.Attachments
					params.Username = b.nick
					params.IconURL = b.icon // permit overriding this
					//					params.Parse = "none"
					_, _, err := api.PostMessage(msg.ChannelID, msg.Text, params)
					if err != nil {
						glog.Infoln("SEND ERROR: ", err.Error())
					}
				} else {
					glog.Infoln("Attempt to send empty message")
				}
			}
		}
	}(b.api, b.Sender)

	for {
		select {
		case msg := <-b.Receiver:
			b.Debugf("event received: %+v", msg.Data)
			switch msg.Data.(type) {
			case slack.HelloEvent:
			case *slack.PresenceChangeEvent:
			case slack.LatencyReport:
			case *slack.MessageEvent:
				b.dispatch(msg.Data.(*slack.MessageEvent))
			default:
				glog.Infof("Unexpected: %v\n", msg.Data)
			}
		}
	}
}

func (b *Bot) Debugf(s string, is ...interface{}) {
	if glog.V(3) {
		glog.Infof(s, is...)
	}
}

func (b *Bot) Send(m *message.Message) {
	b.Sender <- m
}

func (b *Bot) dispatch(ev *slack.MessageEvent) {
	var err error

	var private, tobot bool
	var c *slack.Channel
	glog.Infoln(*ev)

	// We ignore message from our own ID
	if ev.User == b.id || ev.Username == b.nick {
		glog.Infoln("ignoring messagr from us, ", *ev)
		return
	}

	u, err := b.GetUser(ev.User)
	if err != nil {
		glog.Infoln("could not lookgup user ev.UserId")
		return
	}
	glog.Infoln("u.name ", u.Name)

	txt := ev.Msg.Text

	switch {
	case strings.HasPrefix(ev.Channel, "D"):
		{ // One on one,
			private = true
			tobot = true
		}
	case strings.HasPrefix(ev.Channel, "G"):
		{ // private group chat
			private = true
		}
	case strings.HasPrefix(ev.Channel, "C"):
		{
			c, err = b.GetChannel(ev.Channel)
			if err != nil {
				b.Debugf("cannot resolve channel name %s", ev.Channel)
			}
			private = false
		}
	default:
		b.Debugf("cannot determine channel type for %s", ev.Channel)
		return
	}

	// Check if the message was sent @bot, if so, set it as to us
	// and strip the leading politeness
	dirMatch := b.dirPat.FindStringSubmatch(txt)
	if len(dirMatch) > 1 && len(dirMatch[1]) > 0 {
		tobot = true
		txt = strings.Trim(dirMatch[3], " ")
	}

	m := message.Message{
		Event:     ev,
		ChannelID: ev.Channel,
		Channel:   c,
		From:      u,
		Private:   private,
		ToBot:     tobot,
		Text:      txt,
	}

	if m.Private {
		b.Debugf("Handling private message from %v: %v", m.From.Name, m.Text)
	} else {
		b.Debugf("Handling message in %v from %v: %v", m.Channel.Name, m.From.Name, m.Text)
	}

	run := false
	var cmd string
	tokens := strings.Fields(m.Text)
	if len(tokens) > 0 {
		cmd = tokens[0]
	}

	var cmds []string
	for _, h := range handler.Handlers {
		if m.ToBot {
			names := h.Names()
			cmds = append(cmds, names[0])
			for _, n := range names {
				if n == cmd {
					go func(h handler.Handler, msg *message.Message) {
						var err error
						defer func() {
							if r := recover(); r != nil {
								b.Send(m.Replyf("Handler paniced, %v", r))
								return
							}

							switch err {
							case nil, handler.ErrIgnore:
							case handler.ErrAskNicely:
								b.Send(m.Reply("You should ask Nicely"))
							case handler.ErrUnAuthorized:
								b.Send(m.Reply("You are not authorized to do that"))
							case handler.ErrNeedsPrivacy:
								b.Send(m.Reply("You should ask that in private"))
							default:
								b.Send(m.Replyf("error, %v", err.Error()))
							}
						}()

						m.FlagSet = flag.NewFlagSet(cmd, flag.ContinueOnError)
						err = h.Handle(b.Sender, &m)

						return
					}(h, &m)
					run = true
					break
				}
			}
		}

		// If this is not a call for help we'll check all the Hear patterns
		// We have to ignore any message not directly send to us on private channels.
		// All messages sent via the API (not the websocket API), will show us as
		// being from the user we are chatting with EVEN IF WE SENT THEM.
		if hrs := h.Hears(); cmd != "help" && hrs != nil {
			go func(h handler.Handler, hrs handler.HearMap, msg *message.Message) {
				for hr, f := range hrs {
					glog.Infof("%#v", (m))
					if hr.MatchString(m.Text) {
						f(b.Sender, msg)
					}
				}
			}(h, hrs, &m)
		}
	}

	if m.ToBot && !run {
		cmdList := strings.Join(cmds, ",")
		b.Send(m.Replyf("Unknown command '%s', known commands are: %s", cmd, cmdList))
	}
}
