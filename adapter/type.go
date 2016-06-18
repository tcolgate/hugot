package adapter

import (
	"context"

	"github.com/tcolgate/hugot/message"
)

type Receiver interface {
	Receive() <-chan *message.Message
}

type Sender interface {
	Send(ctx context.Context, m *message.Message)
}

type User string
type Channel string

type Adapter interface {
	Sender
	Receiver
}
