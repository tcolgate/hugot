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
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/golang/glog"
	"github.com/nlopes/slack"
	"github.com/tcolgate/hugot/storage"
)

// Message describes a Message from or to a user. It is intended to
// provided a resonable lowest common denominator for modern chat systems.
// It takes the Slack message format to provide that minimum but makes no
// assumption about support for any markup.
// If used within a command handler, the message can also be used as a flag.FlagSet
// for adding and processing the message as a CLI command.
type Message struct {
	To      string
	From    string
	Channel string

	UserID string // Verified user identitify within the source adapter

	Text        string // A plain text message
	Attachments []Attachment

	Private bool
	ToBot   bool

	Store storage.Storer
}

// Copy is used to provide a deep copy of a message
func (m *Message) Copy() *Message {
	nm := *m
	copy(nm.Attachments, m.Attachments)
	return &nm
}

// Attachment represents a rich message attachment and is directly
// modeled on the Slack attachments API
type Attachment slack.Attachment

// Reply returns a messsage with Text tx and the From and To fields switched
func (m *Message) Reply(txt string) *Message {
	out := *m
	out.Text = txt

	out.From = ""
	out.To = m.From

	return &out
}

// Replyf returns message with txt set to the fmt.Printf style formatting,
// and the from/to fields switched.
func (m *Message) Replyf(s string, is ...interface{}) *Message {
	return m.Reply(fmt.Sprintf(s, is...))
}

func stringFromTemplate(tmpls *template.Template, name string, data interface{}) string {
	if tmpls.Lookup(name) != nil {
		out := bytes.Buffer{}
		err := tmpls.ExecuteTemplate(&out, name, data)
		if err != nil {
			glog.Infof("error expanding %v template, %v", name, err.Error())
			return ""
		}
		return out.String()
	}
	return ""
}

// AttachmentFieldFromTemplates builds an attachment field from a set of templates
// templates can include "field_title","field_value", and "field_short". "field_short"
// should expand to "true" or "false".
func AttachmentFieldFromTemplates(tmpls *template.Template, data interface{}) (slack.AttachmentField, error) {
	fieldTitle := stringFromTemplate(tmpls, "field_title", data)
	fieldValue := stringFromTemplate(tmpls, "field_title", data)
	short := false
	fieldShortStr := stringFromTemplate(tmpls, "field_short", data)
	if fieldShortStr == "true" {
		short = true
	}

	return slack.AttachmentField{
		Title: fieldTitle,
		Value: fieldValue,
		Short: short,
	}, nil
}

// AttachmentFromTemplates builds an attachment from aset of templates. Template can include
// the following individual template items: title, title_link, color,
// author_name, author_link, author_icon, image_url,
// thumb_url, pretext, text, fallback, fields_json.
// fields_json will be parsed as json, and any fields found will be appended to those in the fields
// arguments
func AttachmentFromTemplates(tmpls *template.Template, data interface{}, fields ...slack.AttachmentField) (Attachment, error) {
	title := stringFromTemplate(tmpls, "title", data)
	titleLink := stringFromTemplate(tmpls, "title_link", data)
	color := stringFromTemplate(tmpls, "color", data)
	authorName := stringFromTemplate(tmpls, "author_name", data)
	authorLink := stringFromTemplate(tmpls, "author_link", data)
	authorIcon := stringFromTemplate(tmpls, "author_icon", data)
	imageURL := stringFromTemplate(tmpls, "image_url", data)
	thumbURL := stringFromTemplate(tmpls, "thumb_url", data)
	pretext := stringFromTemplate(tmpls, "pretext", data)
	text := stringFromTemplate(tmpls, "text", data)
	fallback := stringFromTemplate(tmpls, "fallback", data)

	fieldsJSON := stringFromTemplate(tmpls, "fields_json", data)
	var jfields []slack.AttachmentField
	if fieldsJSON != "" {
		err := json.Unmarshal([]byte(fieldsJSON), jfields)
		if err != nil {
			glog.Infof("error parsing fields as json, ", err.Error())
		}
	}

	fields = append(fields, jfields...)

	return Attachment{
		Title:      title,
		TitleLink:  titleLink,
		AuthorName: authorName,
		AuthorIcon: authorIcon,
		AuthorLink: authorLink,
		ThumbURL:   thumbURL,
		ImageURL:   imageURL,
		Color:      color,
		Pretext:    pretext,
		Text:       text,
		Fallback:   fallback,
		Fields:     fields,
		MarkdownIn: []string{"pretext", "text", "fields"},
	}, nil
}
