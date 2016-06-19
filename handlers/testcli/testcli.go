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
	"time"

	"github.com/tcolgate/hugot/handler"
	"github.com/tcolgate/hugot/message"
)

func init() {
	handler.MustRegister(New())
}

func New() handler.Handler {
	h, _ := handler.New(
		handler.Description("all being well, says PONG"),
		handler.Handle(Handle),
	)
	return h
}

func Handle(reply chan *message.Message, m *message.Message) error {
	t := m.String("arg", "", "A string argument")
	i := m.Int("num", 0, "An int argument")
	d := m.Duration("time", 1*time.Hour, "A duration argument")
	if err := m.Parse(); err != nil {
		return err
	}

	reply <- m.Replyf("testclivals: (\"%v\",%v,%v)  args: %#v", *t, *i, *d, m.Args())
	return nil
}
