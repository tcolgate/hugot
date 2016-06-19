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
	"runtime/debug"

	"github.com/golang/glog"
)

func ListenAndServe(ctx context.Context, a Adapter, h Handler) {
	if h == nil {
		h = DefaultMux
	}

	if bh, ok := h.(BackgroundHandler); ok {
		glog.Infof("Starting background %v\n", bh)
		go func(ctx context.Context, bh BackgroundHandler) {
			defer func() {
				err := recover()
				if err != nil {
					glog.Error(err)
					glog.Error(string(debug.Stack()))
				}
			}()
			bh.BackgroundHandle(ctx, a)
		}(ctx, bh)
	}

	for {
		select {
		case m := <-a.Receive():
			glog.Infoln(m)
			m.Adapter = a
			go func(ctx context.Context, h Handler, a Adapter, m *Message) {
				defer func() {
					err := recover()
					if err != nil {
						glog.Error(err)
						glog.Error(string(debug.Stack()))
					}
				}()

				processMessage(ctx, h, a, m)
			}(ctx, h, a, m)
		case <-ctx.Done():
			return
		}
	}
}

func processMessage(ctx context.Context, h Handler, a Adapter, m *Message) {
	glog.Infof("Passing %v to %v\n", m, h)
	h.Handle(ctx, a, m)
}
