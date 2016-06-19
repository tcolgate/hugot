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

	"github.com/golang/glog"
	"github.com/tcolgate/hugot"
)

func init() {
	hugot.Add([]string{"ping"}, New())
}

type ping struct {
}

func New() hugot.Handler {
	return &ping{}
}

func (*ping) Handle(ctx context.Context, s hugot.Sender, m *hugot.Message) error {
	glog.Info("Got ping ", *m)
	s.Send(ctx, m.Reply("PONG!"))
	return nil
}
