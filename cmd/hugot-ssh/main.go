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
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"context"

	"github.com/prometheus/client_golang/prometheus"
	bot "github.com/tcolgate/hugot"
	"github.com/tcolgate/hugot/adapters/ssh"

	"github.com/tcolgate/hugot"

	cssh "golang.org/x/crypto/ssh"

	// Add some handlers
	"github.com/tcolgate/hugot/handlers/ping"
	"github.com/tcolgate/hugot/handlers/tableflip"
	"github.com/tcolgate/hugot/handlers/testcli"
	"github.com/tcolgate/hugot/handlers/testweb"
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
	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	config := &cssh.ServerConfig{
		PublicKeyCallback: func(conn cssh.ConnMetadata, pkey cssh.PublicKey) (*cssh.Permissions, error) {
			/*
				pkey, err := models.SearchPublicKeyByContent(strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key))))
				if err != nil {
					// handle error
					return nil, err
				}
				return &cssh.Permissions{Extensions: map[string]string{"key-id": com.ToStr(pkey.ID)}}, nil
			*/
			return nil, nil
		}}

	privateBytes, err := ioutil.ReadFile("host_rsa_key")
	if err != nil {
		panic("Failed to load private key")
	}

	private, err := cssh.ParsePrivateKey(privateBytes)
	if err != nil {
		panic("Failed to parse private key")
	}

	config.AddHostKey(private)

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	listener, err := net.Listen("tcp", "0.0.0.0:2022")
	if err != nil {
		panic("failed to listen for connection")
	}

	a := ssh.New(*nick, listener, config)

	hugot.Handle(ping.New())
	hugot.Handle(testcli.New())
	hugot.Handle(tableflip.New())
	hugot.Handle(testweb.New())

	hugot.HandleBackground(hugot.NewBackgroundHandler("test bg", "testing bg", bgHandler))
	hugot.HandleHTTP(hugot.NewWebHookHandler("test", "test http", httpHandler))

	u, _ := url.Parse("http://localhost:8080")
	hugot.SetURL(u)

	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(":8081", nil)

	bot.ListenAndServe(ctx, a, nil)
	cancel()

	<-ctx.Done()

	//delay to check we get the output
	<-time.After(time.Second * 1)
}
