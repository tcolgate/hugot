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
	"context"
	"time"

	"github.com/tcolgate/hugot"
)

func init() {
	hugot.AddCommandHandler(New())
}

type testcli struct {
	sub *hugot.Mux
}

func New() hugot.CommandHandler {
	sub := hugot.NewMux("testcli-cmds", "somethig somethig")
	sub.AddCommandHandler(&testcliHello{})
	sub.AddCommandHandler(&testcliWorld{})
	return &testcli{sub}
}

func (*testcli) Describe() (string, string) {
	return "testcli", "does many many interesting things"
}

func (*testcli) CommandName() string {
	return "testcli"
}

func (*testcli) Command(ctx context.Context, s hugot.Sender, m *hugot.Message) error {
	t := m.String("arg", "", "A string argument")
	i := m.Int("num", 0, "An int argument")
	d := m.Duration("time", 1*time.Hour, "A duration argument")
	if err := m.Parse(); err != nil {
		return err
	}

	s.Send(ctx, m.Replyf("testclivals: (\"%v\",%v,%v)  args: %#v", *t, *i, *d, m.Args()))
	return nil
}

type testcliHello struct {
}

func (*testcliHello) Describe() (string, string) {
	return "hello", "does many many interesting things"
}

func (*testcliHello) CommandName() string {
	return "hello"
}

func (*testcliHello) Command(ctx context.Context, s hugot.Sender, m *hugot.Message) error {
	if err := m.Parse(); err != nil {
		return err
	}

	s.Send(ctx, m.Replyf("Hello"))
	return nil
}

type testcliWorld struct {
}

func (*testcliWorld) Describe() (string, string) {
	return "world", "does many many interesting things"
}

func (*testcliWorld) CommandName() string {
	return "world"
}

func (*testcliWorld) Command(ctx context.Context, s hugot.Sender, m *hugot.Message) error {
	if err := m.Parse(); err != nil {
		return err
	}

	s.Send(ctx, m.Replyf("World!"))
	return nil
}
