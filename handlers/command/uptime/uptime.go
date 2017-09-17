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

// Package uptime provides a handler that replies to any message sent
package uptime

import (
	"fmt"
	"time"

	"github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/bot"
	"github.com/tcolgate/hugot/handlers/command"
)

var start time.Time

func init() {
	start = time.Now()
}

// New creaates a new command that responds with the uptime of the bot.
func New() *command.Handler {
	return command.NewFunc(func(root *command.Command) error {
		root.Use = "uptime"
		root.Short = "report the uptime of the bot"
		root.Run = func(cmd *command.Command, w hugot.ResponseWriter, m *hugot.Message, args []string) error {
			fmt.Fprintf(w, "I've been running for %s", time.Since(start))

			return nil
		}
		return nil
	})
}

// Register installs this handler on  bot.DefaultBot
func Register() {
	bot.Command(New())
}
