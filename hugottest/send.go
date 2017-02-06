package hugottest

import "github.com/tcolgate/hugot"

// MessagePlayer can be used to pass a channel of messages into a bot.
type MessagePlayer struct {
	MessagesIn chan *hugot.Message
}

// Receive will return a mesasge from the player's channel.
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
