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
	"runtime/debug"

	"github.com/golang/glog"
	"github.com/tcolgate/hugot/adapter"
	"github.com/tcolgate/hugot/handler"
	"github.com/tcolgate/hugot/message"
)

func ListenAndServe(ctx context.Context, a adapter.Adapter, h handler.Handler) {
	if h == nil {
		h = handler.DefaultMux
	}

	if bh, ok := h.(handler.BackgroundHandler); ok {
		glog.Infof("Starting background handler %v\n", bh)
		go func(ctx context.Context, bh handler.BackgroundHandler) {
			defer func() {
				err := recover()
				if err != nil {
					glog.Error(err)
					glog.Error(string(debug.Stack()))
				}
			}()
			bh.BackgroundHandle(ctx, a)
		}(ctx, bh)
	}

	for {
		select {
		case m := <-a.Receive():
			glog.Infoln(m)
			go func(ctx context.Context, h handler.Handler, a adapter.Adapter, m *message.Message) {
				defer func() {
					err := recover()
					if err != nil {
						glog.Error(err)
						glog.Error(string(debug.Stack()))
					}
				}()

				processMessage(ctx, h, a, m)
			}(ctx, h, a, m)
		case <-ctx.Done():
			return
		}
	}
}

func processMessage(ctx context.Context, h handler.Handler, a adapter.Adapter, m *message.Message) {
	glog.Infof("Passing message %v to %v\n", m, h)
	h.Handle(ctx, a, m)
}

/*
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
					if glog.V(3) {
						glog.Infof("%#v", (m))
					}
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
