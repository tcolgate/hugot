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
	"context"
	"fmt"
	"net/http"

	"github.com/tcolgate/hugot"
)

// New create an instance of the example webhook handler.
func New() *TestWeb {

	return &TestWeb{
		hugot.NewWebHookHandler("testweb", "says hello from the world wide web", handleWeb),
	}
}

// TestWeb is a simple example of a handler that gets traffic from a web hook.
type TestWeb struct {
	hugot.WebHookHandler
}

// Describe returns the name and description of the hook.
func (t *TestWeb) Describe() (string, string) {
	return "testweb", "Get url of testweb"
}

// Command implements a simple command that returns the URL of the hook to the users.
func (t *TestWeb) Command(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	fmt.Fprintf(w, "url is %s", t.URL())
	return nil
}

func handleWeb(w http.ResponseWriter, r *http.Request) {
	rw, ok := hugot.ResponseWriterFromContext(r.Context())
	if !ok {
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, "Hello world")
	fmt.Fprintf(rw, "Hello from world wide web")
}
