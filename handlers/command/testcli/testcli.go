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

// Package testcli provides an example Command handler with nested
// command handling.
package testcli

import (
	"fmt"
	"time"

	"github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/bot"
	"github.com/tcolgate/hugot/handlers/command"
)

type testCli struct {
	attach *bool
	e      *string
}

// New adds new testcli command handler
func New() *command.Handler {
	return command.New(&testCli{})
}

func (t *testCli) CommandSetup(root *command.Command) error {
	tctx := *t
	tctx.attach = root.Flags().BoolP("attach", "a", false, "use an attachment")
	tctx.e = root.PersistentFlags().StringP("environment", "e", "staging", "where?")

	root.Use = "testcli"
	root.Short = "test command line thing"
	root.Run = tctx.Command

	wctx := &worldCtx{}
	world := &command.Command{
		Use:   "world",
		Short: "do something else",
		Run:   wctx.Command,
	}
	wctx.d = world.Flags().DurationP("duration", "d", 5*time.Second, "this long")

	root.AddCommand(world)

	return nil
}

type testCliCtx struct {
	attach *bool
}

func (t *testCli) Command(cmd *command.Command, w hugot.ResponseWriter, m *hugot.Message, args []string) error {
	fmt.Fprintf(w, "I'm in here %#v", *t.attach)
	return nil
}

type worldCtx struct {
	d *time.Duration
}

func (wc *worldCtx) Command(cmd *command.Command, w hugot.ResponseWriter, m *hugot.Message, args []string) error {
	fmt.Fprintf(w, "in here: %#v", *wc.d)
	return nil
}

// Register installs this handler on  bot.DefaultBot
func Register() {
	bot.Command(New())
}
