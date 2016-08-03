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

	"context"

	"github.com/tcolgate/hugot"
)

var start time.Time

func init() {
	start = time.Now()
}

func New() hugot.CommandHandler {
	return hugot.NewCommandHandler(
		"uptime",
		"report the uptime of the bot",
		uptimeCommand,
		nil)
}

func uptimeCommand(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	if err := m.Parse(); err != nil {
		return err
	}

	fmt.Fprintf(w, "I've been running for %s", time.Since(start))

	return nil
}
