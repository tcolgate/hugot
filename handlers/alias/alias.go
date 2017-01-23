// Copyright (c) 2016 Tristan Colgate-McFarlane
//
// This file is part of
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
// along with   If not, see <http://www.gnu.org/licenses/>.

package hugot

import (
	"context"
	"fmt"

	"github.com/golang/glog"
)

// AliasHandler implements alias support for use by Mux
type AliasHandler struct {
	c int
}

func NewAliasHandler(s Storer) *AliasHandler {
	return &AliasHandler{}
}

func (*AliasHandler) Describe() (string, string) {
	return "alias", "manage command aliases"
}

func (ah *AliasHandler) ProcessMessage(ctx context.Context, w ResponseWriter, m *Message) error {
	return ah.Command(ctx, w, m)
}

func (ah *AliasHandler) Command(ctx context.Context, w ResponseWriter, m *Message) error {
	g := m.Bool("g", false, "Create alias globally for all users on all channels")
	c := m.Bool("c", false, "Create alias for current channel only")
	u := m.Bool("u", false, "Create alias private for your user only")
	cu := m.Bool("cu", false, "Create alias private for your user, only on this channel")
	d := m.Bool("d", false, "Delete an alias")
	if err := m.Parse(); err != nil {
		return err
	}
	glog.Infof("%v%v%v%v", g, c, cu, u, d)

	val, ok, err := m.Properties().Lookup("thing")
	fmt.Fprintf(w, "set val = %v, ok = %v, err = %v", val, ok, err)

	err = m.Properties().Set(ScopeChannelUser, "thing", fmt.Sprintf("value%d", ah.c))
	err = m.Properties().Set(ScopeChannelUser, fmt.Sprintf("thing%d", ah.c), "value")
	fmt.Fprintf(w, "err = %v", err)
	ah.c++

	ls, _ := m.Store.List("")
	for i, l := range ls {
		fmt.Fprintf(w, "store[%d] = %s", i, l)
	}

	return nil
}

func (ah *AliasHandler) FindAlias(m *Message, str string) (string, bool) {
	v, ok, _ := m.Properties().Lookup(str)
	return v, ok
}
