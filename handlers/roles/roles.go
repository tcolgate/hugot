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

// Package roles is intended to provide Roles Based access controls
// for users and channels. This is a work in progress, and is likely
// to chnage.
package roles

import (
	"context"
	"io"

	"github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/bot"
	"github.com/tcolgate/hugot/handlers/command"
	"github.com/tcolgate/hugot/handlers/mux"
	"github.com/tcolgate/hugot/storage"
)

// Handler implements support for user roles
type Handler struct {
	up hugot.Handler
	cs command.Set
}

// New creates a new roles handler.
func New(up hugot.Handler, cs command.Set, s storage.Storer) *Handler {
	cs.MustAdd(&manager{})

	return &Handler{cs: cs, up: up}
}

// Describe implements the Describer interface for the alias handler
func (h *Handler) Describe() (string, string) {
	return h.up.Describe()
}

// Help implements the command.Helper interfaace for the alias handler
func (h *Handler) Help(w io.Writer) error {
	if hh, ok := h.up.(mux.Helper); ok {
		return hh.Help(w)
	}
	return nil
}

type rolesCtxKeyType int

const rolesCtxKey = rolesCtxKeyType(1)

// ProcessMessage adds any roles the user has to the context
func (h *Handler) ProcessMessage(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	nctx := context.WithValue(ctx, rolesCtxKey, map[string]struct{}{"admin": struct{}{}})
	return h.up.ProcessMessage(nctx, w, m)
}

// FromContext retrieves a set of roles from the current context
func FromContext(ctx context.Context) map[string]struct{} {
	ri := ctx.Value(rolesCtxKey)
	roles := ri.(map[string]struct{})
	if roles == nil {
		roles = map[string]struct{}{}
	}
	return roles
}

// Check verifies the the current user has the requested role
func Check(ctx context.Context, role string) bool {
	_, ok := FromContext(ctx)[role]
	return ok
}

// CheckAny checks to see if the current user has any one of the
// request roles.
func CheckAny(ctx context.Context, want []string) bool {
	roles := FromContext(ctx)

	if len(want) == 0 {
		return true
	}

	for _, r := range want {
		if _, ok := roles[r]; ok {
			return true
		}
	}
	return false
}

// CheckAll checks that the user has all of the request roles.
func CheckAll(ctx context.Context, want []string) bool {
	roles := FromContext(ctx)

	for _, r := range want {
		if _, ok := roles[r]; !ok {
			return false
		}
	}
	return true
}

type manager struct {
}

func (am *manager) Describe() (string, string) {
	return "roles", "manage roles"
}

func (am *manager) Command(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message, args []string) error {
	return nil
}
func (am *manager) CommandSetup(root *command.Command) error {
	root.Use = "roles"
	root.Short = "manager roles"
	root.Run = am.Command
	return nil
}

// Register installs this handler on  bot.DefaultBot
func Register() {
	bot.DefaultBot.Mux.ToBot = New(bot.DefaultBot.Mux.ToBot, bot.DefaultBot.Commands, bot.DefaultBot.Store)
}
