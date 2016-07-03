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

// Package shell implements a simple adapter that provides a readline
// style shell adapter for debugging purposes
package shell

import (
	"log"
	"os"

	"golang.org/x/net/context"

	"github.com/tcolgate/hugot"

	"github.com/chzyer/readline"
)

type shell struct {
	nick string
	user string
	rch  chan *hugot.Message
	sch  chan *hugot.Message
}

func New(nick string) (*shell, error) {
	rch := make(chan *hugot.Message)
	sch := make(chan *hugot.Message)
	return &shell{nick, os.Getenv("USER"), rch, sch}, nil
}

func (s *shell) Send(ctx context.Context, m *hugot.Message) {
	s.sch <- m
}

func (s *shell) Receive() <-chan *hugot.Message {
	return s.rch
}

func (s *shell) Main() {
	rl, err := readline.NewEx(&readline.Config{
		UniqueEditLine: true,
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	rl.ResetHistory()
	log.SetOutput(rl.Stderr())

	rl.SetPrompt(s.user + "> ")

	done := make(chan struct{})
	go func() {
		for {
			select {
			case m := <-s.sch:
				log.Printf("%s: %s", s.nick, m.Text)
			case <-done:
				break
			}
		}
		done <- struct{}{}
	}()

	for {
		ln, err := rl.Readline()
		if err != nil {
			break
		}

		s.rch <- &hugot.Message{Text: ln, ToBot: true, From: s.user}
	}

	rl.Clean()
	done <- struct{}{}
}
