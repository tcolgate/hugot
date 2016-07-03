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

// Package tableflip provides an exacmple Hears handler that will
// flip tables on behalf of embittered users.
package tableflip

import (
	"fmt"
	"regexp"
	"time"

	"golang.org/x/net/context"

	"github.com/tcolgate/hugot"
)

// y=ｰ( ﾟдﾟ)･∵.
// From Will P

func init() {
	hugot.AddHearsHandler(New())
}

type tableflip struct {
}

func New() hugot.HearsHandler {
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

func (*tableflip) Heard(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message, submatches [][]string) {
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
				fmt.Fprint(w, unFlip)
			}
		}()

		switch fs := tableflipRegexp.FindAllString(m.Text, 5); len(fs) {
		case 1:
			fmt.Fprint(w, flip)
		case 2:
			fmt.Fprint(w, doubleFlip)
		case 3:
			fmt.Fprint(w, tripleFlip)
		default:
			fmt.Fprint(w, flipOff)
			flipState = false
		}
		return
	}

	flipState = false
	fmt.Fprint(w, unFlip)
}
