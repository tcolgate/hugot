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
	"context"
	"regexp"
	"time"

	"github.com/tcolgate/hugot"
)

// y=ｰ( ﾟдﾟ)･∵.
// From Will P

func init() {
	hugot.Add(New())
}

type tableflip struct {
}

func New() hugot.Handler {
	return &tableflip{}
}

func (*tableflip) Describe() (string, string) {
	return "tableflip", "stress reliever, just say the word, watch the tables go flying"
}

var tableflipRegexp = regexp.MustCompile(`(^| *)tableflip($| *)`)

func (*tableflip) Hears() *regexp.Regexp {
	return tableflipRegexp
}

// We'll be horrid and use some globals
var flipState bool
var lastFlip time.Time

func (*tableflip) Heard(ctx context.Context, s hugot.Sender, m *hugot.Message, submatches [][]string) {
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
				s.Send(ctx, m.Reply(unFlip))
			}
		}()

		switch fs := tableflipRegexp.FindAllString(m.Text, 5); len(fs) {
		case 1:
			s.Send(ctx, m.Reply(flip))
		case 2:
			s.Send(ctx, m.Reply(doubleFlip))
		case 3:
			s.Send(ctx, m.Reply(tripleFlip))
		default:
			s.Send(ctx, m.Reply(flipOff))
			flipState = false
		}
		return
	}

	flipState = false
	s.Send(ctx, m.Reply(unFlip))
}
