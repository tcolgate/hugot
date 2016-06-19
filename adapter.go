package hugot

import (
	"context"
)

type Receiver interface {
	Receive() <-chan *Message
}

type Sender interface {
	Send(ctx context.Context, m *Message)
}

type User string
type Channel string

type Adapter interface {
	Sender
	Receiver
}
