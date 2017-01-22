package hugottest

import "github.com/tcolgate/hugot"

type MessagePlayer struct {
	MessagesIn chan *hugot.Message
}

func (mp *MessagePlayer) Receive() <-chan *hugot.Message {
	c := make(chan *hugot.Message)

	go func() {
		defer close(c)
		for {
			m := <-mp.MessagesIn
			c <- m
		}
	}()
	return c
}
