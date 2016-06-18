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

package handler

import (
	"context"
	"sync"

	"github.com/tcolgate/hugot/adapter"
	"github.com/tcolgate/hugot/message"
)

func init() {
	DefaultMux = NewMux()
}

var DefaultMux *Mux

func NewMux() *Mux {
	return &Mux{
		RWMutex:  &sync.RWMutex{},
		handlers: make(map[string]Handler),
	}
}

type Mux struct {
	*sync.RWMutex
	handlers map[string]Handler
}

func (mx *Mux) Handle(ctx context.Context, w adapter.Sender, m *message.Message) {
	hs, _ := mx.Handlers(m)
	for _, h := range hs {
		go h.Handle(ctx, w, m)
	}
}

func (mx *Mux) Handlers(m *message.Message) ([]Handler, []string) {
	mx.RLock()
	defer mx.RUnlock()
	ns := []string{}
	hs := []Handler{}
	for n, h := range mx.handlers {
		ns = append(ns, n)
		hs = append(hs, h)
	}
	return hs, ns
}

func Add(s []string, h Handler) error {
	return DefaultMux.Add(s, h)
}

func (mx *Mux) Add(s []string, h Handler) error {
	mx.Lock()
	defer mx.Unlock()

	mx.handlers[s[0]] = h

	return nil
}
