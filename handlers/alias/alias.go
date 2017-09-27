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
	"github.com/tcolgate/hugot/bot"
	"github.com/tcolgate/hugot/handlers/command"
	"github.com/tcolgate/hugot/handlers/mux"
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

// Help implements a command.Help hanndler for the alias handler
func (h *Alias) Help(w io.Writer) error {
	if hh, ok := h.up.(mux.Helper); ok {
		return hh.Help(w)
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

func (am *aliasManager) CommandSetup(root *command.Command) error {
	root.Use = "alias"
	root.Short = "manager aliases"

	var aCtx aliasContext
	aCtx.s = am.s
	aCtx.g = root.Flags().BoolP("global", "g", false, "Create alias globally for all users on all channels")
	aCtx.c = root.Flags().BoolP("channel", "c", false, "Create alias for current channel only")
	aCtx.u = root.Flags().BoolP("user", "u", false, "Create alias private for your user only")
	aCtx.cu = root.Flags().BoolP("channel-user", "C", false, "Create alias private for your user, only on this channel")
	aCtx.d = root.Flags().BoolP("delete", "d", false, "Delete an alias")

	root.Run = aCtx.Command

	return nil
}

type aliasContext struct {
	s storage.Storer

	g  *bool
	c  *bool
	u  *bool
	cu *bool
	d  *bool
}

func (am *aliasContext) Command(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message, args []string) error {
	var store storage.Storer
	switch {
	case !*am.g && !*am.c && !*am.u && !*am.cu:
		if len(args) > 0 {
			return errors.New("to set an alias, select a scope")
		}
		return am.listCmd(w, m)
	case *am.g && !*am.c && !*am.u && !*am.cu:
		store = scoped.New(am.s, scope.Global, m.Channel, m.From)
	case !*am.g && *am.c && !*am.u && !*am.cu:
		store = scoped.New(am.s, scope.Channel, m.Channel, m.From)
	case !*am.g && !*am.c && *am.u && !*am.cu:
		store = scoped.New(am.s, scope.User, m.Channel, m.From)
	case !*am.g && !*am.c && !*am.u && *am.cu:
		store = scoped.New(am.s, scope.ChannelUser, m.Channel, m.From)
	default:
		return fmt.Errorf("Specify exactly one of -g, -c, -cu or -u")
	}

	if *am.d {
		if len(args) != 1 {
			return errors.New("delete requires one alias name to delete")
		}
		return store.Unset(args)
	}

	if len(args) < 2 {
		return errors.New("you must provide an alias name and expansion")
	}

	strs := []string{}
	for _, str := range args[1:] {
		strs = append(strs, fmt.Sprintf("%q", str))
	}

	return store.Set([]string{args[0]}, strings.Join(strs, " "))
}

func (am *aliasContext) listCmd(w hugot.ResponseWriter, m *hugot.Message) error {
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

// Register installs this handler on  bot.DefaultBot
func Register() {
	bot.DefaultBot.Mux.ToBot = New(bot.DefaultBot.Mux.ToBot, bot.DefaultBot.Commands, bot.DefaultBot.Store)
}
