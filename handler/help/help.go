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

package handlers

import (
	"fmt"
	"strings"

	"github.com/tcolgate/hugot/handler"
	"github.com/tcolgate/hugot/message"
)

func init() {
	handler.MustRegister(New())
}

func New() handler.Handler {
	h, _ := handler.New(
		handler.Description("provide help"),
		handler.Help("help [command]"),
		handler.Handle(Handle),
	)
	return h
}

func Handle(reply chan *message.Message, m *message.Message) error {
	if !m.Private {
		return handler.ErrNeedsPrivacy
	}

	args := strings.Fields(m.Text)

	var msg string

	// Help for a specific handler
	if len(args) > 1 {
		found := false
		var cmds []string

		for _, h := range handler.Handlers {
			cmds = append(cmds, h.Names()[0])
			if args[1] == h.Names()[0] {
				found = true
				msg = fmt.Sprintf("help for %s:\n", args[1])
				msg += h.Help()
				break
			}
		}
		if !found {
			msg = fmt.Sprintf("No such command '', vailable commands are: %s\n", strings.Join(cmds, ","))
		}
	} else {
		//Help summary
		msg = "Available Commands: \n"
		for _, h := range handler.Handlers {
			msg += fmt.Sprintf(" %s - %s\n", h.Names()[0], h.Describe())
		}
	}

	reply <- m.Reply(msg)
	return nil
}
