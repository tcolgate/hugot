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

// Package mattermost implements an adapter for http://mm.com
package mattermost

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"context"

	"github.com/golang/glog"
	"github.com/tcolgate/hugot"

	mm "github.com/mattermost/mattermost-server/model"
)

type mma struct {
	email string

	client *mm.Client4
	user   *mm.User
	team   *mm.Team
	cache  *cache

	id   string
	icon string

	dirPat      *regexp.Regexp
	api         *mm.Client4
	initialLoad *mm.InitialLoad

	ws *mm.WebSocketClient

	sender chan *hugot.Message
}

// New creates a new adapter that communicates with Mattermost
func New(apiurl, team, email, password string) (hugot.Adapter, error) {
	c := mma{client: mm.NewAPIv4Client(apiurl)}

	user, resp := c.client.Login(email, password)
	if resp.Error != nil {
		return nil, resp.Error
	}

	c.user = user

	teams, resp := c.client.GetAllTeams("", 0, 10000)
	glog.Infof("Teams: %#v", resp)
	if resp.Error != nil {
		return nil, resp.Error
	}

	for _, t := range teams {
		if t.Name == team {
			c.team = t
			break
		}
	}
	c.cache = newCache(c.client, c.team)

	if c.team == nil {
		return nil, fmt.Errorf("Could not find team %s", team)
	}

	pat := fmt.Sprintf("(?m)^(!|(@?%s)[:,]? )(.*)", c.user.Username)
	c.dirPat = regexp.MustCompile(pat)

	wsurl, _ := url.Parse(apiurl)
	wsurl.Scheme = "ws"

	var resperr *mm.AppError
	c.ws, resperr = mm.NewWebSocketClient4(wsurl.String(), c.client.AuthToken)
	if resperr != nil {
		return nil, resperr
	}

	c.ws.Listen()

	return &c, nil
}

func (s *mma) Send(ctx context.Context, m *hugot.Message) {
	post := &mm.Post{}
	ch, err := s.cache.GetChannelByName(m.Channel)
	if err != nil {
		glog.Errorf("could not look up channel, %v", err)
		ch = &mm.Channel{Id: m.Channel}
	}

	post.ChannelId = ch.Id
	post.Message = m.Text
	var attchs []*mm.SlackAttachment
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
		flds := []*mm.SlackAttachmentField{}
		for _, f := range a.Fields {
			flds = append(flds, &mm.SlackAttachmentField{
				Title: f.Title,
				Value: f.Value,
				Short: f.Short,
			},
			)
		}
		attchs = append(attchs,
			&mm.SlackAttachment{
				Fallback:   a.Fallback,
				Pretext:    a.Pretext,
				Text:       a.Text,
				Title:      a.Title,
				TitleLink:  a.TitleLink,
				ImageURL:   a.ImageURL,
				ThumbURL:   a.ThumbURL,
				Color:      a.Color,
				AuthorName: a.AuthorName,
				AuthorLink: a.AuthorLink,
				AuthorIcon: a.AuthorIcon,
				Fields:     flds,
			})
	}

	if len(attchs) > 0 {
		post.AddProp("attachments", attchs)
	}

	if _, resp := s.client.CreatePost(post); resp.Error != nil {
		glog.Errorf("error creating post, %v", resp.Error)
	}
}

func (s *mma) Receive() <-chan *hugot.Message {
	out := make(chan *hugot.Message, 1)
	for {
		select {
		case m := <-s.ws.EventChannel:
			if glog.V(3) {
				glog.Infof("mattermost event: %#v\n", m)
			}
			switch m.Event {
			case mm.WEBSOCKET_EVENT_POSTED:
				p := mm.PostFromJson(strings.NewReader(m.Data["post"].(string)))
				if p == nil || p.UserId == s.user.Id {
					continue
				}
				out <- s.mmMsgToHugot(m)
				return out
			default:
				glog.Infof("unknown event: %#v\n", m)
			}
		}
	}
}

func (s *mma) mmMsgToHugot(me *mm.WebSocketEvent) *hugot.Message {
	var private, tobot bool
	if glog.V(3) {
		glog.Infof("mattermost message: %#v\n", *me)
	}

	p := mm.PostFromJson(strings.NewReader(me.Data["post"].(string)))

	ct, ok := me.Data["channel_type"]
	if !ok {
		glog.Infoln("channel_type not set")
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
		p.Message = strings.Trim(dirMatch[3], " ")
	}

	ch, err := s.cache.GetChannel(p.ChannelId)
	if err != nil {
		glog.Errorf("could not resolve incoming channel name")
		ch = &mm.Channel{Id: p.ChannelId}
	}

	user, err := s.cache.GetUser(p.UserId)
	if err != nil {
		glog.Errorf("could not resolve incoming user name")
		user = &mm.User{Id: p.UserId, Username: p.UserId}
		return nil
	}

	m := hugot.Message{
		Channel: ch.Id,
		UserID:  user.Id,
		From:    user.Username,
		To:      "",
		Private: private,
		ToBot:   tobot,
		Text:    p.Message,
	}

	if glog.V(3) {
		glog.Infof("hugot message: %#v\n", m)
	}

	return &m
}
