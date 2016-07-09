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

// Package testweb provides an example webhook handler
package testweb

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"github.com/tcolgate/hugot"
)

func New() hugot.WebHookHandler {
	return hugot.NewWebHookHandler("testweb", "does things", handleWeb)
}

func handleWeb(ctx context.Context, hw hugot.ResponseWriter, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world")

	hw.SetChannel("bottest")
	fmt.Fprint(hw, "Hello from world wide web")
}
