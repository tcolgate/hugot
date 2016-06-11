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

import "github.com/tcolgate/hugot/message"

var Handlers = map[string]Handler{}

func MustRegister(h Handler) {
	if len(h.Names()) == 0 {
		panic("attempt to register handler with no names")
	}
	Handlers[h.Names()[0]] = h
}

func StartAll(send chan *message.Message) {
	for _, h := range Handlers {
		go h.Start(send)
	}
}
