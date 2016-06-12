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
	"log"
	"runtime"
	"strings"

	"github.com/tcolgate/hugot/message"
)

type base struct {
	names       []string
	description string
	help        string
	hears       HearMap

	setup  SetupFunc
	start  StartFunc
	handle HandleFunc
}

type option func(*base) error

func New(opts ...option) (Handler, error) {
	var err error
	b := &base{names: []string{defaultName()}}
	log.Println(b)
	for _, opt := range opts {
		if err = opt(b); err != nil {
			return nil, err
		}
	}

	return b, nil
}

func Name(s string) option {
	return func(p *base) error {
		p.names = append(p.names, s)
		return nil
	}
}

func (b *base) Names() []string {
	return b.names
}

func Description(s string) option {
	return func(p *base) error {
		p.description = s
		return nil
	}
}

func (b *base) Describe() string {
	return b.description
}

func Help(s string) option {
	return func(p *base) error {
		p.help = s
		return nil
	}
}

func (b *base) Help() string {
	return b.help
}

func Hears(hm HearMap) option {
	return func(p *base) error {
		p.hears = hm
		return nil
	}
}

func (b *base) Hears() HearMap {
	return b.hears
}

func Setup(f SetupFunc) option {
	return func(p *base) error {
		p.setup = f
		return nil
	}
}

func (h *base) Setup() error {
	if h.setup != nil {
		return h.setup()
	}
	return nil
}

func Start(f StartFunc) option {
	return func(p *base) error {
		p.start = f
		return nil
	}
}

func (h *base) Start(send chan *message.Message) {
	if h.start != nil {
		h.start(send)
	}
}

func Handle(f HandleFunc) option {
	return func(p *base) error {
		p.handle = f
		return nil
	}
}

func (h *base) Handle(send chan *message.Message, m *message.Message) error {
	if h.handle != nil {
		return h.handle(send, m)
	}
	return nil
}

func defaultName() string {
	pc, _, _, _ := runtime.Caller(2)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), "/")
	pl := len(parts)

	return strings.Split(parts[pl-1], ".")[0]
}
