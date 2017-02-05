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
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"context"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	bot "github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/adapters/shell"
	"github.com/tcolgate/hugot/storage/redis"

	"github.com/tcolgate/hugot"

	// Add some handlers

	"github.com/tcolgate/hugot/handlers/alias"
	"github.com/tcolgate/hugot/handlers/command"
	"github.com/tcolgate/hugot/handlers/command/ping"
	"github.com/tcolgate/hugot/handlers/command/testcli"
	"github.com/tcolgate/hugot/handlers/command/uptime"
	"github.com/tcolgate/hugot/handlers/hears/tableflip"
	"github.com/tcolgate/hugot/handlers/help"
	"github.com/tcolgate/hugot/handlers/mux"
	"github.com/tcolgate/hugot/handlers/roles"
	goredis "gopkg.in/redis.v5"
)

var nick = flag.String("nick", "minion", "Bot nick")

func bgHandler(ctx context.Context, w hugot.ResponseWriter) {
	fmt.Fprint(w, "Starting backgroud")
	<-ctx.Done()
	fmt.Fprint(w, "Stopping backgroud")
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%#v", *r)
	w.Write([]byte("hello world"))
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	a, err := shell.New(*nick)
	if err != nil {
		glog.Fatal(err)
	}

	mux.Background(hugot.NewBackgroundHandler("test bg", "testing bg", bgHandler))
	mux.HandleHTTP(hugot.NewWebHookHandler("test", "test http", httpHandler))

	mux.DefaultMux.Hears(tableflip.New())

	command.DefaultSet.MustAdd(testcli.New())
	command.DefaultSet.MustAdd(uptime.New())
	command.DefaultSet.MustAdd(ping.New())
	command.DefaultSet.MustAdd(help.New(mux.DefaultMux))

	redisOpts, err := goredis.ParseURL("redis://localhost:6379")
	if err != nil {
		glog.Fatal(err)
	}

	hugot.DefaultStore = redis.New(redisOpts)
	hugot.DefaultHandler = alias.New(hugot.DefaultHandler, command.DefaultSet, hugot.DefaultStore)
	hugot.DefaultHandler = roles.New(hugot.DefaultHandler, command.DefaultSet, hugot.DefaultStore)

	u, _ := url.Parse("http://localhost:8080")
	mux.SetURL(u)

	go bot.ListenAndServe(ctx, hugot.DefaultHandler, a)
	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(":8081", nil)

	a.Main()

	cancel()

	<-ctx.Done()

	//delay to check we get the output
	<-time.After(time.Second * 1)
}
