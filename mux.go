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
	cmds     *CommandMux                       // Command handlers
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
		cmds:     NewCommandMux(nil),
	}
	mx.AddCommandHandler(&muxHelp{mx})
	return mx
}

func (mx *Mux) BackgroundHandler(ctx context.Context, w ResponseWriter) {
	mx.RLock()
	defer mx.RUnlock()

	for _, h := range mx.bghndlrs {
		go RunBackgroundHandler(ctx, h, w)
	}
}

func (mx *Mux) Handle(ctx context.Context, w ResponseWriter, m *Message) error {
	mx.RLock()
	defer mx.RUnlock()

	err := RunCommandHandler(ctx, mx.cmds, w, m)

	// If it wasn't a known command, run the hear commands
	if err == ErrIgnored {
		for _, hhs := range mx.hears {
			for _, hh := range hhs {
				mc := *m
				go RunHearsHandler(ctx, hh, w, &mc)
			}
		}
	}

	// We run all raw message handlers
	for _, rh := range mx.rhndlrs {
		mc := *m
		go rh.Handle(ctx, w, &mc)
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
	//	name, _ := h.Describe()

	if !used {
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
	//	name, _ := h.Describe()

	if h, ok := h.(RawHandler); ok {
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
	//name, _ := h.Describe()

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

func AddCommandHandler(h CommandHandler) *CommandMux {
	return DefaultMux.AddCommandHandler(h)
}

func (mx *Mux) AddCommandHandler(h CommandHandler) *CommandMux {
	mx.Lock()
	defer mx.Unlock()

	return mx.cmds.AddCommandHandler(h)
}

func (mx *Mux) Describe() (string, string) {
	return mx.name, mx.desc
}

var (
	// ErrIgnore
	ErrIgnored = errors.New("the handler ignored the message")

	// ErrNextCommand is returned if the command wishes the message
	// to be passed to one of the SubCommands.
	ErrNextCommand = errors.New("pass this to the next command")
)

type CommandMux struct {
	base    CommandHandler
	subCmds map[string]CommandHandler // Command handlers
}

func NewCommandMux(base CommandHandler) *CommandMux {
	mx := &CommandMux{
		base:    base,
		subCmds: map[string]CommandHandler{},
	}
	return mx
}

func (cx *CommandMux) AddCommandHandler(c CommandHandler) *CommandMux {
	n, _ := c.Describe()

	subMux := NewCommandMux(c)
	cx.subCmds[n] = subMux
	return subMux
}

// not sure what to do here.
func (cx *CommandMux) Describe() (string, string) {
	if cx.base != nil {
		return cx.base.Describe()
	}
	return "", ""
}

func (cx *CommandMux) Command(ctx context.Context, w ResponseWriter, m *Message) error {
	var err error
	if cx.base != nil {
		err = cx.base.Command(ctx, w, m)
	} else {
		err = ErrNextCommand
	}

	if err != ErrNextCommand {
		return err
	}

	glog.Infof("IN COMMANDMUX COMMAND %v, %v, %v", err, m.args, cx.subCmds)
	if len(m.args) == 0 {
		return fmt.Errorf("handler requested next command, but no args remain")
	}

	subs := cx.subCmds
	for {
		if cmd, ok := subs[m.args[0]]; ok {
			err = cmd.Command(ctx, w, m)
			if err != ErrNextCommand {
				return err
			}
			var subh CommandWithSubsHandler
			var ok bool
			if subh, ok = cmd.(CommandWithSubsHandler); !ok {
				glog.Infof("NOT A SUBS HANDLER")
				return nil
			}
			subs = subh.SubCommands()
		} else {
			err = ErrUnknownCommand
			break
		}
	}

	return err
}

func (cx *CommandMux) SubCommands() map[string]CommandHandler {
	return cx.subCmds
}
