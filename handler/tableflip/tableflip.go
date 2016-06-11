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

package handlers

import (
	"regexp"
	"time"

	"github.com/tcolgate/hugot/handler"
	"github.com/tcolgate/hugot/message"
)

// y=ｰ( ﾟдﾟ)･∵.
// From Will P

func init() {
	handler.MustRegister(New())
}

func New() handler.Handler {
	h, _ := handler.New(
		handler.Hears(handler.HearMap{
			tableflipRegexp: heardFlip,
		}),
		handler.Description("stress reliever"),
		handler.Help("just say the word, watch the tables go flying"),
	)
	return h
}

var tableflipRegexp = regexp.MustCompile(`(^| *)tableflip($| *)`)

// We'll be horrid and use some globals
var flipState bool
var lastFlip time.Time

func heardFlip(s chan *message.Message, m *message.Message) {
	flip := `(╯°□°）╯︵ ┻━┻`
	unFlip := `┬━┬ ノ( ゜-゜ノ)`
	doubleFlip := "┻━┻ ︵¯\\(ツ)/¯ ︵ ┻━┻"
	tripleFlip := "(╯°□°）╯¸.·´¯`·.¸¸.·´¯`·.¸¸.·´¯ ┻━┻"
	flipOff := "ಠ︵ಠ凸"

	if !flipState {
		flipState = true

		go func() {
			five, _ := time.ParseDuration("30s")
			time.Sleep(five)
			if flipState == true {
				flipState = false
				s <- m.Reply(unFlip)
			}
		}()

		switch fs := tableflipRegexp.FindAllString(m.Text, 5); len(fs) {
		case 1:
			s <- m.Reply(flip)
		case 2:
			s <- m.Reply(doubleFlip)
		case 3:
			s <- m.Reply(tripleFlip)
		default:
			s <- m.Reply(flipOff)
			flipState = false
		}
		return
	}

	flipState = false
	s <- m.Reply(unFlip)
}
