// Copyright (c) 2017 Tristan Colgate-McFarlane
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

//Package help provides a help command, to allow users to get instructions
//on how to use other commands, and find out what commands are available.
package help

import (
	"bytes"
	"io"

	"github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/handlers/command"

	"context"
)

// Helper describes an interface to be implemented by handlers that wish
// to provide help to users.
type Helper interface {
	// Help should write any help text, in a pleasantly formatted fashion, to
	// the provided writer.
	Help(ctx context.Context, w io.Writer, m *command.Message) error
}

// Handler implements the help handler.
type Handler struct {
	nh Helper
}

// New creates a new help handler. It should be pass an implementation of a
// Helper to serve the root of the help information.
func New(nh Helper) *Handler {
	h := &Handler{nh}
	return h
}

// Describe returns the name and description of the handler.
func (h *Handler) Describe() (string, string) {
	return "help", "provides description of handler usage"
}

// Command implements the user facing help command handler.
func (h *Handler) Command(ctx context.Context, w hugot.ResponseWriter, m *command.Message) error {
	out := &bytes.Buffer{}
	m.Parse()
	if err := h.nh.Help(ctx, out, m); err != nil {
		w.Send(ctx, m.Reply(err.Error()))
	} else {
		w.Send(ctx, m.Reply(string(out.Bytes())))
	}
	return command.ErrSkipHears
}
