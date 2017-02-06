package hugottest

import (
	"context"

	"github.com/tcolgate/hugot"
)

// ResponseRecorder will send all messages a handler writes to it
// out of the provided channel.
type ResponseRecorder struct {
	MessagesOut chan hugot.Message

	defchan string
	defto   string
}

// Send is called to send a message to the recorder's channel.
func (rr *ResponseRecorder) Send(ctx context.Context, m *hugot.Message) {
	if m.Channel == "" {
		m.Channel = rr.defchan
	}
	rr.MessagesOut <- *m
}

// Wrtie implement io.Write, but sending data written to it as a single
// hugot.Message.
func (rr *ResponseRecorder) Write(bs []byte) (int, error) {
	nmsg := hugot.Message{
		Channel: rr.defchan,
		To:      rr.defto,
	}
	nmsg.Text = string(bs)
	rr.Send(context.TODO(), &nmsg)
	return len(bs), nil
}

// SetChannel is used to decide which channel data sent use Write
// will be sent on.
func (rr *ResponseRecorder) SetChannel(c string) {
	rr.defchan = c
}

// SetTo tells the record who to send data sent with Write to.
func (rr *ResponseRecorder) SetTo(to string) {
	rr.defto = to
}
