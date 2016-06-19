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
	"regexp"
	"sync"
)

func init() {
	DefaultMux = NewMux()
}

var DefaultMux *Mux

func NewMux() *Mux {
	return &Mux{
		RWMutex:  &sync.RWMutex{},
		handlers: make(map[string]Handler),
	}
}

type Mux struct {
	*sync.RWMutex
	handlers map[string]Handler
}

func (mx *Mux) BackgroundHandler(ctx context.Context, w Sender) {
	mx.RLock()
	defer mx.RUnlock()

	for _, h := range mx.handlers {
		if bh, ok := h.(BackgroundHandler); ok {
			go bh.BackgroundHandle(ctx, w)
		}
	}
}

func (mx *Mux) Handle(ctx context.Context, w Sender, m *Message) error {
	mx.RLock()
	defer mx.RUnlock()

	hs, _ := mx.Handlers(m)
	for _, h := range hs {
		go h.Handle(ctx, w, m)
	}
	return nil
}

func (mx *Mux) Hears() []*regexp.Regexp {
	mx.RLock()
	defer mx.RUnlock()

	hrs := []*regexp.Regexp{}
	for _, h := range mx.handlers {
		if hh, ok := h.(HearsHandler); ok {
			hrs = append(hrs, hh.Hears()...)
		}
	}
	return hrs
}

func (mx *Mux) Handlers(m *Message) ([]Handler, []string) {
	mx.RLock()
	defer mx.RUnlock()

	ns := []string{}
	hs := []Handler{}
	for n, h := range mx.handlers {
		ns = append(ns, n)
		hs = append(hs, h)
	}
	return hs, ns
}

func Add(s []string, h Handler) error {
	return DefaultMux.Add(s, h)
}

func (mx *Mux) Add(s []string, h Handler) error {
	mx.Lock()
	defer mx.Unlock()

	mx.handlers[s[0]] = h

	return nil
}

/*
	if m.Private {
		b.Debugf("Handling private from %v: %v", m.From.Name, m.Text)
	} else {
		b.Debugf("Handling in %v from %v: %v", m.Channel.Name, m.From.Name, m.Text)
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
					go func(h handler.Handler, msg *Message) {
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

						err = h.Handle(b.Sender, &m)

						return
					}(h, &m)
					run = true
					break
				}
			}
		}

		// If this is not a call for help we'll check all the Hear patterns
		// We have to ignore any not directly send to us on private channels.
		// All  sent via the API (not the websocket API), will show us as
		// being from the user we are chatting with EVEN IF WE SENT THEM.
		if hrs := h.Hears(); cmd != "help" && hrs != nil {
			go func(h handler.Handler, hrs handler.HearMap, msg *Message) {
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
*/
