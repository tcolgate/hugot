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

	"golang.org/x/net/context"

	"github.com/golang/glog"
	bot "github.com/tcolgate/hugot"
	irc "github.com/tcolgate/hugot/adapters/irc"
	irce "github.com/thoj/go-ircevent"

	// Add some handlers
	_ "github.com/tcolgate/hugot/handlers/ping"
	_ "github.com/tcolgate/hugot/handlers/tableflip"
	_ "github.com/tcolgate/hugot/handlers/testcli"
)

var (
	nick    = flag.String("nick", "hugottest", "Bot nick")
	user    = flag.String("irc.user", "", "IRC username")
	pass    = flag.String("irc.pass", "", "IRC password")
	server  = flag.String("irc.server", "chat.freenode.net:6697", "Server to connect to")
	ircchan = flag.String("irc.channel", "#hugottest", "Channel to listen in")
	useSSL  = flag.Bool("irc.usessl", true, "Use SSL to connect")
)

func main() {
	flag.Parse()

	c := irce.IRC(*nick, *user)
	c.UseTLS = true
	c.Password = *pass

	err := c.Connect(*server)
	if err != nil {
		glog.Fatal(err)
	}
	c.Join(*ircchan)
	defer c.Quit()

	ctx := context.Background()
	a, err := irc.New(c)

	go bot.ListenAndServe(ctx, a, nil)

	c.Loop()

}
