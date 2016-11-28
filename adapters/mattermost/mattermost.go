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

// Package mattermost implements an adapter for http://mm.com using
// github.com/nlopes/mm
package mm

import (
	"fmt"
	"regexp"
	"strings"

	"context"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tcolgate/hugot"

	mm "github.com/mattermost/platform/model"
)

var (
	mmLatency = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "mm_latency_reports_millisecond",
		Help: "Latency of Slack in milliseconds.",
	})
	mmMessagesTx = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mm_messages_sent_total",
		Help: "Number of mm messages sent.",
	})
	mmMessagesRx = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mm_messages_received_total",
		Help: "Number of mm messages received.",
	})
)

type mma struct {
	email string

	client *mm.Client
	user   *mm.User
	team   *mm.Team

	id   string
	icon string

	dirPat      *regexp.Regexp
	api         *mm.Client
	initialLoad *mm.InitialLoad

	ws *mm.WebSocketClient

	sender chan *hugot.Message
}

// New creates a new adapter that communicates with Mattermost
func New(url, team, email, password string) (hugot.Adapter, error) {
	c := mma{client: mm.NewClient(url)}

	lr, err := c.client.Login(email, password)
	if err != nil {
		return nil, err
	}

	c.user = lr.Data.(*mm.User)

	ilr, err := c.client.GetInitialLoad()
	if err != nil {
		return nil, err
	}

	c.initialLoad = ilr.Data.(*mm.InitialLoad)
	for _, t := range c.initialLoad.Teams {
		if t.Name == team {
			c.team = t
			break
		}
	}

	if c.team == nil {
		return nil, fmt.Errorf("Could not find team %s", team)
	}

	c.client.SetTeamId(c.team.Id)

	pat := fmt.Sprintf("^@%s[:,]? (.*)", c.user.Username)
	c.dirPat = regexp.MustCompile(pat)
	c.ws, err = mm.NewWebSocketClient("ws://localhost:8065", c.client.AuthToken)
	if err != nil {
		return nil, err
	}

	c.ws.Listen()

	return &c, nil
}

func (s *mma) Send(ctx context.Context, m *hugot.Message) {
	post := &mm.Post{}
	post.ChannelId = m.Channel
	post.Message = m.Text
	var attchs []map[string]interface{}
	for _, a := range m.Attachments {
		switch a.Color {
		case "good":
			a.Color = "#00ff00"
		case "warning":
			a.Color = "#ff1010"
		case "danger":
			a.Color = "#ff0000"
		}
		if a.Fallback == "" {
			a.Fallback = a.Text
		}
		flds := []map[string]interface{}{}
		for _, f := range a.Fields {
			flds = append(flds, map[string]interface{}{
				"title": f.Title,
				"value": f.Value,
				"short": f.Short,
			})
		}
		attchs = append(attchs,
			map[string]interface{}{
				"fallback":       a.Fallback,
				"pretext":        a.Pretext,
				"text":           a.Text,
				"title":          a.Title,
				"title_link":     a.TitleLink,
				"image_url":      a.ImageURL,
				"thumb_url":      a.ThumbURL,
				"color":          a.Color,
				"author_name":    a.AuthorName,
				"author_subname": a.AuthorSubname,
				"author_link":    a.AuthorLink,
				"author_icon":    a.AuthorIcon,
				"fields":         flds,
			})
	}

	if len(attchs) > 0 {
		post.Type = mm.POST_SLACK_ATTACHMENT
		post.Props = make(map[string]interface{})
		post.Props["attachments"] = attchs
	}

	if _, err := s.client.CreatePost(post); err != nil {
		glog.Infoln(err.Error())
	}
}

func (s *mma) Receive() <-chan *hugot.Message {
	out := make(chan *hugot.Message, 1)
	go func() {
		for {
			select {
			case m := <-s.ws.EventChannel:
				switch m.Event {
				case mm.WEBSOCKET_EVENT_POSTED:
					p := mm.PostFromJson(strings.NewReader(m.Data["post"].(string)))
					if p == nil || p.UserId == s.user.Id {
						continue
					}
					out <- s.mmMsgToHugot(m)
				default:
				}
			}
		}
	}()
	return out
}

func (s *mma) mmMsgToHugot(me *mm.WebSocketEvent) *hugot.Message {
	var private, tobot bool
	if glog.V(3) {
		glog.Infof("message: %#v\n", *me)
	}

	p := mm.PostFromJson(strings.NewReader(me.Data["post"].(string)))

	var uname string

	uname = p.UserId
	if uname == "" {
		glog.Infoln("could not resolve username")
		return nil
	}

	ct, ok := me.Data["channel_type"]
	if !ok {
		return nil
	}

	switch ct.(string) {
	case "D":
		{ // One on one,
			private = true
			tobot = true
		}
	case "P":
		{ // private group chat
			private = true
		}
	case "O":
	default:
		glog.Errorf("cannot determine channel type for %s", p.ChannelId)
		return nil
	}

	// Check if the message was sent @bot, if so, set it as to us
	// and strip the leading politeness
	dirMatch := s.dirPat.FindStringSubmatch(p.Message)
	if len(dirMatch) > 1 && len(dirMatch[1]) > 0 {
		tobot = true
		p.Message = strings.Trim(dirMatch[1], " ")
	}

	m := hugot.Message{
		Channel: p.ChannelId,
		From:    uname,
		To:      "",
		UserID:  p.UserId,
		Private: private,
		ToBot:   tobot,
		Text:    p.Message,
	}

	return &m
}
