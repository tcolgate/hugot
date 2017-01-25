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

package alias

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/handlers/command"
)

// AliasHandler implements alias support for use by Mux
type AliasHandler struct {
	up hugot.Handler
	cs command.Set
}

// New creates a new alias handler and registers as a command on
// the Mux
func New(up hugot.Handler, cs command.Set, s hugot.Storer) hugot.Handler {
	cs.MustAdd(&aliasManager{})

	return &AliasHandler{cs: cs, up: up}
}

func (h *AliasHandler) Describe() (string, string) {
	return h.up.Describe()
}

// ProcessMessage attempets to execute the message using the Mux, if that fails,
// it looks for aliases in the properties of the message. If a suitable alias
// is found, that is executed. It also adds an alias manager command for
// managing the set of aliases
func (h *AliasHandler) ProcessMessage(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	err := h.up.ProcessMessage(ctx, w, m)
	if err == command.ErrUnknownCommand {
		return h.execAlias(ctx, w, m)
	}
	return err
}

func (h *AliasHandler) execAlias(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	parts := strings.SplitN(m.Text, " ", 2)
	if len(parts) != 1 && len(parts) != 2 {
		return command.ErrUnknownCommand
	}

	/*
		v, ok, _ := m.Properties().Lookup(parts[0])
	*/
	aliases := map[string]string{"x": "ping"}
	v, ok := aliases[parts[0]]
	if !ok {
		return command.ErrUnknownCommand
	}

	m.Text = v
	if len(parts) == 2 {
		m.Text = fmt.Sprintf("%s %s", m.Text, parts[1])
	}
	return h.up.ProcessMessage(ctx, w, m)
}

// aliasManager
type aliasManager struct {
}

func (am *aliasManager) Describe() (string, string) {
	return "alias", "manage command aliases"
}

func (am *aliasManager) Command(ctx context.Context, w hugot.ResponseWriter, m *command.Message) error {
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

	fmt.Fprintf(w, "err = %v", err)

	ls, _ := m.Store.List("")
	for i, l := range ls {
		fmt.Fprintf(w, "store[%d] = %s", i, l)
	}

	return nil
}
