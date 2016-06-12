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
	"log"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/tcolgate/hugot/handler"
	"github.com/tcolgate/hugot/message"

	"github.com/nlopes/slack"
)

type Bot struct {
	slackToken string
	nick       string

	id   string
	icon string

	dirPat *regexp.Regexp
	api    *slack.Client

	debug bool

	Sender   chan *message.Message
	Receiver chan slack.RTMEvent

	cacheLock sync.Mutex
	userCache map[string]*slack.User
	chanCache map[string]*slack.Channel
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

func (b *Bot) Start() {
	if b.Token() == "" {
		log.Println("Slack Token must be set")
		os.Exit(1)
	}

	//b.debug = c.Bool("debug")

	//if c.String("log") != "-" {
	//		hupablelog.SetupLog(c.String("log"))
	//}

	for _, h := range handler.Handlers {
		log.Println("calling setup for %s: %v", h.Names()[0], h.Setup)
		err := h.Setup()
		if err != nil {
			log.Fatalf("Error in handler %s: %s", h.Names()[0], err.Error)
			os.Exit(1)
			return
		}
	}

	b.api = slack.New(b.Token())
	log.Println(b.api)
	b.api.SetDebug(b.debug)

	us, _ := b.api.GetUsers()

	for _, u := range us {
		if u.Name == b.nick {
			b.id = u.ID
			b.icon = u.Profile.Image72
		}
	}

	if b.id == "" {
		log.Fatalf("Could not locate bot's user ID")
	}

	b.dirPat = regexp.MustCompile(fmt.Sprintf("(?m)^(!|(@?%s|<@%s>)[:, ]?)(.*)", b.nick, b.id))

	b.run()
}

func (b *Bot) GetUser(id string) (*slack.User, error) {
	b.cacheLock.Lock()
	defer b.cacheLock.Unlock()

	if b.userCache == nil {
		b.userCache = make(map[string]*slack.User)
	}

	if u, ok := b.userCache[id]; ok {
		return u, nil
	}

	u, err := b.api.GetUserInfo(id)
	if err != nil {
		return nil, err
	}

	b.userCache[id] = u
	return u, nil
}

func (b *Bot) GetChannel(id string) (*slack.Channel, error) {
	b.cacheLock.Lock()
	defer b.cacheLock.Unlock()

	if b.chanCache == nil {
		b.chanCache = make(map[string]*slack.Channel)
	}

	if c, ok := b.chanCache[id]; ok {
		return c, nil
	}

	c, err := b.api.GetChannelInfo(id)
	if err != nil {
		return nil, err
	}

	b.chanCache[id] = c
	return c, nil
}

func (b *Bot) run() {
	b.Sender = make(chan *message.Message)

	wsAPI := b.api.NewRTM()
	b.Receiver = wsAPI.IncomingEvents

	go wsAPI.ManageConnection()
	go func(api *slack.Client, sender chan *message.Message) {
		for {
			select {
			case msg := <-sender:
				if msg.Text != "" && msg.ChannelID != "" {
					/*
						  // The RTM api doesn't support attachments, and that's
							// just no fun
							smsg := wsAPI.NewOutgoingMessage(msg.Text, msg.ChannelID)
							smsg.Attachments = attchs
							wsAPI.SendMessage(smsg)
					*/
					params := slack.NewPostMessageParameters()
					params.AsUser = false
					params.Attachments = msg.Attachments
					params.Username = b.nick
					params.IconURL = b.icon // permit overriding this
					//					params.Parse = "none"
					_, _, err := api.PostMessage(msg.ChannelID, msg.Text, params)
					if err != nil {
						log.Println("SEND ERROR: ", err.Error())
					}
				} else {
					log.Println("Attempt to send empty message")
				}
			}
		}
	}(b.api, b.Sender)

	for {
		select {
		case msg := <-b.Receiver:
			fmt.Print("Event Received: ")
			switch msg.Data.(type) {
			case slack.HelloEvent:
				// Ignore hello
			case *slack.MessageEvent:
				b.dispatch(msg.Data.(*slack.MessageEvent))
			case *slack.PresenceChangeEvent:
				a := msg.Data.(*slack.PresenceChangeEvent)
				fmt.Printf("Presence Change: %v\n", a)
			case slack.LatencyReport:
				a := msg.Data.(slack.LatencyReport)
				fmt.Printf("Current latency: %v\n", a.Value)
			default:
				fmt.Printf("Unexpected: %v\n", msg.Data)
			}
		}
	}
}

func (b *Bot) Debugf(s string, is ...interface{}) {
	if b.debug {
		log.Printf(s, is...)
	}
}

func (b *Bot) Send(m *message.Message) {
	b.Sender <- m
}

func (b *Bot) dispatch(ev *slack.MessageEvent) {
	var err error

	var private, tobot bool
	var c *slack.Channel
	log.Println(*ev)

	//log.Println("bot.id ", b.Id)
	//log.Println("ev.id ", ev.UserId)
	//log.Println("ev.username ", ev.Username)
	//log.Println("ev.msg.id ", ev.Msg.UserId)

	// We ignore message from our own ID
	if ev.User == b.id || ev.Username == b.nick {
		log.Println("ignoring messagr from us, ", *ev)
		return
	}

	u, err := b.GetUser(ev.User)
	if err != nil {
		log.Println("could not lookgup user ev.UserId")
		return
	}
	log.Println("u.name ", u.Name)

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
					log.Printf("%#v", (m))
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
