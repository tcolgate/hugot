package hugottest

import (
	"context"

	"github.com/tcolgate/hugot"
)

type ResponseRecorder struct {
	MessagesOut chan hugot.Message

	defchan string
	defto   string
}

func (rr *ResponseRecorder) Send(ctx context.Context, m *hugot.Message) {
	if m.Channel == "" {
		m.Channel = rr.defchan
	}
	rr.MessagesOut <- *m
}

func (rr *ResponseRecorder) Write(bs []byte) (int, error) {
	nmsg := hugot.Message{
		Channel: rr.defchan,
		To:      rr.defto,
	}
	nmsg.Text = string(bs)
	rr.Send(context.TODO(), &nmsg)
	return len(bs), nil
}

func (rr *ResponseRecorder) SetChannel(c string) {
	rr.defchan = c
}

func (rr *ResponseRecorder) SetTo(to string) {
	rr.defto = to
}

func (rr *ResponseRecorder) SetSender(a hugot.Sender) {
	// Not sure if this is usefault
}
