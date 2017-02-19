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
	"github.com/tcolgate/hugot/handlers/help"
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

func (h *Handler) Describe() (string, string) {
	return h.up.Describe()
}

func (h *Handler) Help(ctx context.Context, w io.Writer, m *command.Message) error {
	if hh, ok := h.up.(help.Helper); ok {
		return hh.Help(ctx, w, m)
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

func FromContext(ctx context.Context) map[string]struct{} {
	ri := ctx.Value(rolesCtxKey)
	roles := ri.(map[string]struct{})
	if roles == nil {
		roles = map[string]struct{}{}
	}
	return roles
}

func Check(ctx context.Context, role string) bool {
	_, ok := FromContext(ctx)[role]
	return ok
}

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

func (am *manager) Command(ctx context.Context, w hugot.ResponseWriter, m *command.Message) error {

	return nil
}

func Register() {
	bot.DefaultBot.Mux.ToBot = New(bot.DefaultBot.Mux.ToBot, bot.DefaultBot.Commands, bot.DefaultBot.Store)
}
