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
	"errors"
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
	mx := &Mux{
		name:     name,
		desc:     desc,
		RWMutex:  &sync.RWMutex{},
		rhndlrs:  []RawHandler{},
		bghndlrs: []BackgroundHandler{},
		hears:    map[*regexp.Regexp][]HearsHandler{},
		cmds:     map[string]CommandHandler{},
	}
	mx.AddCommandHandler(&muxHelp{mx})
	return mx
}

func (mx *Mux) BackgroundHandler(ctx context.Context, w ResponseWriter) {
	mx.RLock()
	defer mx.RUnlock()

	for _, h := range mx.bghndlrs {
		go runBGHandler(ctx, w, h)
	}
}

func (mx *Mux) Handle(ctx context.Context, w ResponseWriter, m *Message) error {
	mx.RLock()
	defer mx.RUnlock()

	for _, h := range mx.SelectHandlers(m) {
		go runOneHandler(ctx, w, h, m)
	}

	return nil
}

func Add(h Handler) error {
	return DefaultMux.Add(h)
}

// Add a generic handler with potentially multiple
func (mx *Mux) Add(h Handler) error {
	var used bool
	if h, ok := h.(RawHandler); ok {
		mx.AddRawHandler(h)
		used = true
	}

	if h, ok := h.(BackgroundHandler); ok {
		mx.AddBackgroundHandler(h)
		used = true
	}

	if h, ok := h.(CommandHandler); ok {
		mx.AddCommandHandler(h)
		used = true
	}

	if h, ok := h.(HearsHandler); ok {
		mx.AddHearsHandler(h)
		used = true
	}

	mx.Lock()
	defer mx.Unlock()
	name, _ := h.Describe()

	if !used {
		glog.Errorf("failed to register %v, not a recognised handler type", name)
		return fmt.Errorf("Don't know how to use %T as a handler", h)
	}

	mx.hndlrs = append(mx.hndlrs, h)

	return nil
}

func AddRawHandler(h RawHandler) error {
	return DefaultMux.AddRawHandler(h)
}

func (mx *Mux) AddRawHandler(h RawHandler) error {
	mx.Lock()
	defer mx.Unlock()
	name, _ := h.Describe()

	if h, ok := h.(RawHandler); ok {
		glog.Errorf("Registered raw handler %v", name)
		mx.rhndlrs = append(mx.rhndlrs, h)
	}

	return nil
}

func AddBackgroundHandler(h BackgroundHandler) error {
	return DefaultMux.AddBackgroundHandler(h)
}

func (mx *Mux) AddBackgroundHandler(h BackgroundHandler) error {
	mx.Lock()
	defer mx.Unlock()
	name, _ := h.Describe()

	glog.Errorf("Registered baclground handler %v", name)
	mx.bghndlrs = append(mx.bghndlrs, h)

	return nil
}

func AddHearsHandler(h HearsHandler) error {
	return DefaultMux.AddHearsHandler(h)
}

func (mx *Mux) AddHearsHandler(h HearsHandler) error {
	mx.Lock()
	defer mx.Unlock()
	name, _ := h.Describe()

	glog.Errorf("Registered hears handler %v", name)
	r := h.Hears()
	mx.hears[r] = append(mx.hears[r], h)

	return nil
}

func AddCommandHandler(h CommandHandler) error {
	return DefaultMux.AddCommandHandler(h)
}

func (mx *Mux) AddCommandHandler(h CommandHandler) error {
	mx.Lock()
	defer mx.Unlock()
	name, _ := h.Describe()

	glog.Errorf("Registered command handler %v", name)
	n := name
	mx.cmds[n] = h

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

// ErrNextCommand is returned if the command wishes the message
// to be passed to one of the SubCommands.
var ErrNextCommand = errors.New("pass this to the next command")

type CommandMux struct {
	name string
	desc string
	*sync.RWMutex
	subCmds map[string]CommandHandler // Command handlers
}

// NewCommandMux returns a CommandMux using the provided
// CommandHandler as the base handler
func NewCommandMux(name string, description string) *CommandMux {
	mx := &CommandMux{
		subCmds: map[string]CommandHandler{},
	}
	return mx
}

func (cx *CommandMux) AddSubCommand(c CommandHandler) {
	cx.Lock()
	defer cx.Unlock()

	n, _ := c.Describe()
	cx.subCmds[n] = c
}

func (*CommandMux) Command(ctx context.Context, h CommandHandler, m *Message) {
}

func (*CommandMux) SubCommands() map[string]CommandHandler {
	return nil
}
