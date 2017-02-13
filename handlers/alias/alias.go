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

// Package alias imeplements user customizable aliases for commands. The user
// can set their own aliases, per channel, or system wide.
package alias

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/handlers/command"
	"github.com/tcolgate/hugot/handlers/help"
	"github.com/tcolgate/hugot/scope"
	"github.com/tcolgate/hugot/storage"
	"github.com/tcolgate/hugot/storage/prefix"
	"github.com/tcolgate/hugot/storage/properties"
	"github.com/tcolgate/hugot/storage/scoped"
)

// Alias implements alias support for use by Mux
type Alias struct {
	up hugot.Handler
	cs command.Set
	s  storage.Storer
}

// New creates a new alias handler and registers the alias command
// with the the Mux, to permit users to manage their aliases.
func New(up hugot.Handler, cs command.Set, s storage.Storer) hugot.Handler {
	store := prefix.New(s, []string{"aliases"})
	cs.MustAdd(&aliasManager{store})

	return &Alias{
		cs: cs,
		up: up,
		s:  store,
	}
}

// Describe returns the name and description of the handler.
func (h *Alias) Describe() (string, string) {
	return h.up.Describe()
}

// ProcessMessage attempets to execute the message using the Mux, if that fails,
// it looks for aliases in the properties of the message. If a suitable alias
// is found, that is executed. It also adds an alias manager command for
// managing the set of aliases
func (h *Alias) ProcessMessage(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	err := h.up.ProcessMessage(ctx, w, m)
	if err == command.ErrUnknownCommand {
		return h.execAlias(ctx, w, m)
	}
	return err
}

func (h *Alias) Help(ctx context.Context, w io.Writer, m *command.Message) error {
	if hh, ok := h.up.(help.Helper); ok {
		return hh.Help(ctx, w, m)
	}
	return nil
}

func (h *Alias) execAlias(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	props := properties.NewPropertyStore(h.s, m)

	parts := strings.SplitN(m.Text, " ", 2)
	if len(parts) != 1 && len(parts) != 2 {
		return command.ErrUnknownCommand
	}

	v, ok, _ := props.Get([]string{parts[0]})
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
	s storage.Storer
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

	var store storage.Storer
	switch {
	case !*g && !*c && !*u && !*cu:
		if len(m.Args()) > 0 {
			return errors.New("to set an alias, select a scope")
		}
		return am.listCmd(ctx, w, m)
	case *g && !*c && !*u && !*cu:
		store = scoped.New(am.s, scope.Global, m.Channel, m.From)
	case !*g && *c && !*u && !*cu:
		store = scoped.New(am.s, scope.Channel, m.Channel, m.From)
	case !*g && !*c && *u && !*cu:
		store = scoped.New(am.s, scope.User, m.Channel, m.From)
	case !*g && !*c && !*u && *cu:
		store = scoped.New(am.s, scope.ChannelUser, m.Channel, m.From)
	default:
		return fmt.Errorf("Specify exactly one of -g, -c, -cu or -u")
	}

	if *d {
		if len(m.Args()) != 1 {
			return errors.New("delete requires one alias name to delete")
		}
		return store.Unset([]string{m.Arg(0)})
	}

	if len(m.Args()) < 2 {
		return errors.New("you must provide an alias name and expansion")
	}

	strs := []string{}
	for _, str := range m.Args()[1:] {
		strs = append(strs, fmt.Sprintf("%q", str))
	}

	return store.Set([]string{m.Arg(0)}, strings.Join(strs, " "))
}

func (am *aliasManager) listCmd(ctx context.Context, w hugot.ResponseWriter, m *command.Message) error {
	out := &bytes.Buffer{}
	for _, s := range scope.Order {
		store := scoped.New(am.s, s, m.Channel, m.From)
		aliases, err := store.List([]string{})
		if err != nil {
			return err
		}

		if len(aliases) > 0 {
			fmt.Fprintf(out, "Aliases %s\n", s.Describe(m.Channel, m.From))
			tw := new(tabwriter.Writer)
			tw.Init(out, 0, 8, 1, '\t', 0)
			for _, k := range aliases {
				a, ok, err := store.Get(k)
				if !ok || err != nil {
					continue
				}
				fmt.Fprintf(tw, "  %s\t -> %s\n", k[0], a)
			}
			tw.Flush()
		} else {
			fmt.Fprintf(out, "No aliases %s\n", s.Describe(m.Channel, m.From))
		}
	}
	io.Copy(w, out)

	return nil
}
