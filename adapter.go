package hugot

import (
	"context"
)

// Receiver cam ne used to receive messages
type Receiver interface {
	Receive() <-chan *Message // Receive returns a channel that can be used to read one message, nil indicated there will be no more messages
}

// Sender can be used to send messages
type Sender interface {
	Send(ctx context.Context, m *Message)
}

type User string
type Channel string

// SenderReceiver is used to both send a receive messages
type SenderReceiver interface {
	Sender
	Receiver
}
