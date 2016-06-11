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
	"log"
	"os"

	bot "github.com/tcolgate/hugot"
	_ "github.com/tcolgate/hugot/handler/help"
	_ "github.com/tcolgate/hugot/handler/ping"
	_ "github.com/tcolgate/hugot/handler/tableflip"
	_ "github.com/tcolgate/hugot/handler/testcli"
)

var slackToken = flag.String("token", os.Getenv("SLACK_TOKEN"), "Slack API Token")
var nick = flag.String("nick", "hugot", "Bot nick")

func main() {
	flag.Parse()

	bot, err := bot.New(
		bot.Token(*slackToken),
		bot.Nick(*nick),
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println(bot)
	bot.Start()
}
