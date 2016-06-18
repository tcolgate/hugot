package adapter

import "github.com/tcolgate/hugot/message"

type Receiver interface {
	Receive() <-chan *message.Message
}

type User string
type Channel string

type Adapter interface {
	message.Sender
	Receiver
}
