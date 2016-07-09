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

package hugot_test

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	"github.com/tcolgate/hugot"
)

func ExampleMessage_Parse(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	v := m.Bool("v", false, "be verbose")
	dur := m.Duration("d", 1*time.Hour, "how long to do thing for")
	baz := m.String("baz", "some argument", "another agument")

	if err := m.Parse(); err != nil {
		return err
	}

	fmt.Fprintf(w, "Got args %#v %#v %#v", *v, *dur, *baz)

	return nil
}
