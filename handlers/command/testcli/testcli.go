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
	"log"
	"time"

	"context"

	"github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/bot"
	"github.com/tcolgate/hugot/handlers/command"
)

type testCli struct {
	command.Commander

	cs  command.Set
	wcs command.Set
}

// New adds new testcli command handler
func New() command.Commander {
	t := &testCli{
		cs: command.NewSet(
			command.New("hello", "but hello to what", helloCommand),
			command.New("world", "the whole thing", worldCommand)),
		wcs: command.NewSet(
			command.New("world", "deeper down the rabbit hole", world2Command)),
	}

	t.Commander = command.New("testcli", "test command line thing", t.Command)
	return t
}

func (t *testCli) Command(ctx context.Context, w hugot.ResponseWriter, m *command.Message) error {
	if err := m.Parse(); err != nil {
		return err
	}

	log.Printf("I'm in here %#v", *m)
	return t.cs.Command(ctx, w, m)
}

func helloCommand(ctx context.Context, w hugot.ResponseWriter, m *command.Message) error {
	a := m.Bool("a", false, "use an attachment")
	if err := m.Parse(); err != nil {
		return err
	}

	if !*a {
		fmt.Print(w, "Hello")
		return nil
	}
	r := m.Reply("")
	r.Attachments = []hugot.Attachment{
		{
			Text:  "Hello",
			Color: "good",
		},
	}
	w.Send(ctx, r)
	return nil
}

func worldCommand(ctx context.Context, w hugot.ResponseWriter, m *command.Message) error {
	if err := m.Parse(); err != nil {
		return err
	}
	return nil
}

func world2Command(ctx context.Context, w hugot.ResponseWriter, m *command.Message) error {
	_ = m.String("arg", "", "A string argument")
	_ = m.Int("num", 0, "An int argument")
	_ = m.Duration("time", 1*time.Hour, "A duration argument")
	_ = m.Bool("v", false, "verbose")
	if err := m.Parse(); err != nil {
		return err
	}

	fmt.Fprint(w, "Deeper!")
	return nil
}

func Register() {
	bot.Command(New())
}
