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

package hugot

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/golang/glog"
)

func init() {
	DefaultMux = NewMux("defaultMux", "")
}

type Mux struct {
	name string
	desc string

	*sync.RWMutex
	hndlrs   []Handler                         // All the handlers
	rhndlrs  []RawHandler                      // Raw handlers
	bghndlrs []BackgroundHandler               // Long running background handlers
	hears    map[*regexp.Regexp][]HearsHandler // Hearing handlers
	cmds     map[string]CommandHandler         // Command handlers
}

var DefaultMux *Mux

func NewMux(name, desc string) *Mux {
	return &Mux{
		name:     name,
		desc:     desc,
		RWMutex:  &sync.RWMutex{},
		rhndlrs:  []RawHandler{},
		bghndlrs: []BackgroundHandler{},
		hears:    map[*regexp.Regexp][]HearsHandler{},
		cmds:     map[string]CommandHandler{},
	}
}

func (mx *Mux) BackgroundHandler(ctx context.Context, w Sender) {
	mx.RLock()
	defer mx.RUnlock()

	for _, h := range mx.bghndlrs {
		go runBGHandler(ctx, w, h)
	}
}

func (mx *Mux) Handle(ctx context.Context, w Sender, m *Message) error {
	mx.RLock()
	defer mx.RUnlock()

	for _, h := range mx.SelectHandlers(m) {
		go runOneHandler(ctx, h, m)
	}

	return nil
}

func Add(h Handler) error {
	return DefaultMux.Add(h)
}

func (mx *Mux) Add(h Handler) error {
	mx.Lock()
	defer mx.Unlock()
	name, _ := h.Describe()

	var used bool
	if h, ok := h.(RawHandler); ok {
		glog.Errorf("Registered raw handler %v", name)
		mx.rhndlrs = append(mx.rhndlrs, h)
		used = true
	}

	if h, ok := h.(BackgroundHandler); ok {
		glog.Errorf("Registered baclground handler %v", name)
		mx.bghndlrs = append(mx.bghndlrs, h)
		used = true
	}

	if h, ok := h.(CommandHandler); ok {
		glog.Errorf("Registered command handler %v", name)
		n := h.CommandName()
		mx.cmds[n] = h
		used = true
	}

	if h, ok := h.(HearsHandler); ok {
		glog.Errorf("Registered hears handler %v", name)
		r := h.Hears()
		mx.hears[r] = append(mx.hears[r], h)
		used = true
	}

	if !used {
		glog.Errorf("failed to register %v, not a recognised handler type", name)
		return fmt.Errorf("Don't know how to use %T as a handler", h)
	}

	mx.hndlrs = append(mx.hndlrs, h)

	return nil
}

func (mx *Mux) Describe() (string, string) {
	return mx.name, mx.desc
}

func (mx *Mux) SelectHandlers(m *Message) []Handler {
	hs := []Handler{}
	var cmd CommandHandler
	cmdStr := ""

	if tks := strings.Fields(m.Text); m.ToBot && len(tks) > 0 {
		// We should add the help handler here
		cmdStr = tks[0]
		if _, ok := mx.cmds[cmdStr]; ok {
			cmd = mx.cmds[cmdStr]
			hs = append(hs, Handler(cmd))
		}
	}

	// if this isn't a help request, we'll apply
	// all the hear handlers
	if cmdStr != "help" {
		for _, hhs := range mx.hears {
			for _, hh := range hhs {
				hs = append(hs, Handler(hh))
			}
		}
	}

	// We were sent a direct message, but we don't
	// have a matching command
	if m.ToBot && cmd == nil {
		// We should add the help handler here
	}

	// We run all raw message handlers
	for _, rh := range mx.rhndlrs {
		hs = append(hs, Handler(rh))
	}

	return hs
}

/*
func doCmd(h handler.Handler, msg *Message) {
}
*/

func (mx *Mux) CommandName() string {
	return "help"
}

func (mx *Mux) Command(ctx context.Context, w Sender, m *Message) error {
	return nil
}
