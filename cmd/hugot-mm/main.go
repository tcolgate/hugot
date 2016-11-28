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

package main

import (
	"flag"
	"net/http"
	"net/url"

	"context"

	"github.com/golang/glog"
	bot "github.com/tcolgate/hugot"

	// Add some handlers
	"github.com/tcolgate/hugot"
	mm "github.com/tcolgate/hugot/adapters/mattermost"
	"github.com/tcolgate/hugot/handlers/ping"
	"github.com/tcolgate/hugot/handlers/tableflip"
	"github.com/tcolgate/hugot/handlers/testcli"
	"github.com/tcolgate/hugot/handlers/testweb"
)

var mmurl = flag.String("url", "http://localhost:8065", "url")
var team = flag.String("team", "team-t", "team name")
var mail = flag.String("email", "hugot@test.net", "Bot mail")
var pass = flag.String("pass", "hugot", "Bot pass")

func main() {
	flag.Parse()

	ctx := context.Background()
	a, err := mm.New(*mmurl, *team, *mail, *pass)
	if err != nil {
		glog.Fatal(err)
	}

	hugot.Handle(ping.New())
	hugot.Handle(tableflip.New())
	hugot.Handle(testcli.New())

	wh := testweb.New()
	hugot.HandleHTTP(wh)

	u, _ := url.Parse("http://localhost:8080")
	hugot.SetURL(u)

	glog.Infof("webhook at %s", wh.URL())

	go http.ListenAndServe(":8080", nil)
	bot.ListenAndServe(ctx, nil, a)
}
