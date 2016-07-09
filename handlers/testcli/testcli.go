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

func New() hugot.CommandHandler {
	wcs := hugot.NewCommandSet()
	wcs.AddCommandHandler(hugot.NewCommandHandler("world", "deeper down the rabbit hole", world2Command, nil))

	cs := hugot.NewCommandSet()
	cs.AddCommandHandler(hugot.NewCommandHandler("hello", "but hello to what", helloCommand, nil))
	cs.AddCommandHandler(hugot.NewCommandHandler("world", "the whole thing", worldCommand, wcs))

	return hugot.NewCommandHandler("testcli", "test command line thing", nil, cs)
}

func helloCommand(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	if err := m.Parse(); err != nil {
		return err
	}

	fmt.Print(w, "Hello")
	return nil
}

func worldCommand(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	if err := m.Parse(); err != nil {
		return err
	}
	return hugot.ErrNextCommand(ctx)
}

func world2Command(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
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
