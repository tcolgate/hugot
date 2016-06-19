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
	"regexp"
)

var (
	ErrIgnore       = errors.New("ignored message")
	ErrAskNicely    = errors.New("potentially dangerous must ask nicely")
	ErrUnAuthorized = errors.New("you are not orthorized to perform this action")
	ErrNeedsPrivacy = errors.New("potentially dangerous must ask nicely")
)

type SetupFunc func() error
type StartFunc func(chan *Message) error
type HandleFunc func(ctx context.Context, s Sender, m *Message)

type Handler interface {
	Handle(ctx context.Context, s Sender, m *Message) error
}

type BackgroundHandler interface {
	BackgroundHandle(ctx context.Context, s Sender)
}

type HearsHandler interface {
	Hears() []*regexp.Regexp
}

type CommandHandler interface {
}

//type HearHandlerFunc func(send chan *Message, msg *Message)
//type HearMap map[*regexp.Regexp]HearHandlerFunc
//	Setup() error
//	Describe() string // A list of names/aliases for this command
//	Names() []string  // A list of names/aliases for this command
//	Help() string     // A list of names/aliases for this command
