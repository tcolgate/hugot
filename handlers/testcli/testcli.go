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

	"golang.org/x/net/context"

	"github.com/tcolgate/hugot"
)

func init() {
	hugot.AddCommandHandler(New())
}

func New() hugot.CommandHandler {
	mux := hugot.NewCommandMux(&testcli{})
	mux.AddCommandHandler(&testcliHello{})
	wmux := mux.AddCommandHandler(&testcliWorld{})
	wmux.AddCommandHandler(&testcliWorld2{})
	return mux
}

type testcli struct {
}

func (*testcli) Describe() (string, string) {
	return "testcli", "does many many interesting things"
}

func (*testcli) Command(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	t := m.String("arg", "", "A string argument")
	i := m.Int("num", 0, "An int argument")
	d := m.Duration("time", 1*time.Hour, "A duration argument")
	if err := m.Parse(); err != nil {
		return err
	}

	fmt.Fprintf(w, "testclivals: (\"%v\",%v,%v)  args: %#v", *t, *i, *d, m.Args())
	return hugot.ErrNextCommand
}

type testcliHello struct {
}

func (*testcliHello) Describe() (string, string) {
	return "hello", "does many many interesting things"
}

func (*testcliHello) Command(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	if err := m.Parse(); err != nil {
		return err
	}

	fmt.Print(w, "Hello")
	return nil
}

type testcliWorld struct {
}

func (*testcliWorld) Describe() (string, string) {
	return "world", "does many many interesting things"
}

func (*testcliWorld) Command(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	if err := m.Parse(); err != nil {
		return err
	}

	return hugot.ErrNextCommand
}

type testcliWorld2 struct {
}

func (*testcliWorld2) Describe() (string, string) {
	return "world", "does many many interesting things"
}

func (*testcliWorld2) Command(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	if err := m.Parse(); err != nil {
		return err
	}

	fmt.Fprint(w, "Deeper!")
	return nil
}
