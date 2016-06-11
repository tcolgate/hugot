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
	"errors"
	"flag"
	"regexp"

	"github.com/tcolgate/hugot/message"
)

var (
	ErrIgnore       = errors.New("ignored message")
	ErrAskNicely    = errors.New("potentially dangerous must ask nicely")
	ErrUnAuthorized = errors.New("you are not orthorized to perform this action")
	ErrNeedsPrivacy = errors.New("potentially dangerous must ask nicely")
	ErrBadCLI       = errors.New("coul not process as command line")
)

type HearHandlerFunc func(send chan *message.Message, msg *message.Message)
type HearMap map[*regexp.Regexp]HearHandlerFunc

type Handler interface {
	Setup() error
	Names() []string  // A list of names/aliases for this command
	Describe() string // A list of names/aliases for this command
	Help() string     // A list of names/aliases for this command

	Start(send chan *message.Message) // Triggered when the bot is ready
	Handle(send chan *message.Message, m *message.Message) error
	Hears() HearMap
	BuildFlags() *flag.FlagSet
}

type SetupFunc func() error
type StartFunc func(chan *message.Message) error
type HandleFunc func(chan *message.Message, *message.Message) error
type BuildFlagsFunc func() *flag.FlagSet

type CLIHandler interface {
	Handler
	SetupFlags(*flag.FlagSet)
}
