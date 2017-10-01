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

// Package ping provides a handler that replies to any message sent
package ping

import (
	"context"
	"fmt"

	"github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/bot"
	"github.com/tcolgate/hugot/handlers/command"
)

// New creates a new ping command that responds with a Pong
func New() *command.Handler {
	return command.NewFunc(func(root *command.Command) error {
		root.Use = "ping"
		root.Short = "confirms the bot is running"
		root.Run = func(ctx context.Context, w hugot.ResponseWriter, msg *hugot.Message, args []string) error {
			fmt.Fprintf(w, "PONG!")
			return nil
		}
		return nil
	})
}

// Register installs this handler on  bot.DefaultBot
func Register() {
	bot.Command(New())
}
