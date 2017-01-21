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

package hugot

import (
	"context"
	"fmt"

	"github.com/golang/glog"
)

// ListenAndServe runs the handler h, passing all messages to/from
// the provided adapter. The context may be used to gracefully shut
// down the server.
func ListenAndServe(ctx context.Context, h Handler, a Adapter, as ...Adapter) {
	if h == nil {
		h = DefaultMux
	}

	ctx = NewAdapterContext(ctx, a)

	an := fmt.Sprintf("%T", a)
	if bh, ok := h.(BackgroundHandler); ok {
		runBackgroundHandler(ctx, bh, newResponseWriter(a, Message{}, an))
	}

	if wh, ok := h.(WebHookHandler); ok {
		wh.SetAdapter(a)
	}

	type smrw struct {
		w ResponseWriter
		m *Message
	}
	mrws := make(chan smrw)

	for _, a := range append(as, a) {
		go func(a Adapter) {
			an := fmt.Sprintf("%T", a)
			for {
				select {
				case m := <-a.Receive():
					rw := newResponseWriter(a, *m, an)
					mrws <- smrw{rw, m}
				case <-ctx.Done():
					return
				}
			}
		}(a)
	}

	for {
		select {
		case mrw := <-mrws:
			if glog.V(3) {
				glog.Infof("Message: %#v", *mrw.m)
			}
			messagesRx.WithLabelValues(an, mrw.m.Channel, mrw.m.From).Inc()

			go h.ProcessMessage(ctx, mrw.w, mrw.m)

		case <-ctx.Done():
			return
		}
	}
}
