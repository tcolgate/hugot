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

// Package hugot provides a simple interface for building extensible
// chat bots in an idiomatic go style. It is heavily influenced by
// net/http, and uses an internal message format that is compatible
// with Slack messages.
//
// Note: This package requires go1.7
//
// Adapters
//
// Adapters are used to integrate with external chat systems. Currently
// the following adapters exist:
//   slack - github.com/tcolgate/hugot/adapters/slack - for https://slack.com/
//   mattermost - github.com/tcolgate/hugot/adapters/mattermost - for https://www.mattermost.org/
//   irc - github.com/tcolgate/hugot/adapters/irc - simple irc adapter
//   shell - github.com/tcolgate/hugot/adapters/shell - simple readline based adapter
//   ssh - github.com/tcolgate/hugot/adapters/ssh - Toy implementation of unauth'd ssh interface
//
// Examples of using these adapters can be found in github.com/tcolgate/hugot/cmd
//
// Handlers
//
// Handlers process messages. There are a several built in handler types:
//
// - Plain Handlers will execute for every message sent to them.
//
// - Background handlers, are started when a bot is started. They do not
// receive messages but can send them. They are intended to implement long
// lived background tasks that react to external inputs.
//
// - WebHook handlers can be used to implement web hooks by adding the bot to a
// http.ServeMux. A URL is built from the name of the handler.
//
// In addition to these basic handlers some more complex handlers are supplied.
//
// - Hears Handlers will execute for any message which matches a given regular
// expression.
//
// - Command Handlers act as command line tools. Message are attempted to be
// processed as a command line. Quoted text is handle as a single argument. The
// passed message can be used as a flag.FlagSet.
//
// - A Mux. The Mux will multiplex message across a set of handlers. In addition,
// a top level "help" Command handler is added to provide help on usage of the
// various handlers added to the Mux.
//
// WARNING: The API is still subject to change.
package hugot
