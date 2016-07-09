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
	"crypto/tls"
	"flag"
	"strings"

	"golang.org/x/net/context"

	// Add some handlers
	"github.com/fluffle/goirc/client"
	"github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/adapters/irc"
	"github.com/tcolgate/hugot/handlers/ping"
	"github.com/tcolgate/hugot/handlers/tableflip"
)

var (
	nick    = flag.String("nick", "hugot-prom", "Bot nick")
	user    = flag.String("irc.user", "hugot-prom", "IRC username")
	pass    = flag.String("irc.pass", "", "IRC password")
	server  = flag.String("irc.server", "chat.freenode.net:6697", "Server to connect to")
	ircchan = flag.String("irc.channel", "#hugot", "Channel to listen in")
	useSSL  = flag.Bool("irc.usessl", true, "Use SSL to connect")
)

func main() {
	flag.Parse()

	c := client.NewConfig(*nick)
	c.Server = *server
	c.SSL = *useSSL
	c.Pass = *pass
	c.SSLConfig = &tls.Config{ServerName: strings.Split(*server, ":")[0]}

	a := irc.New(c, *ircchan)

	hugot.Add(ping.New())
	hugot.Add(tableflip.New())

	hugot.ListenAndServe(context.Background(), a, nil)
}
