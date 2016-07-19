package hugottest

import (
	"sync"

	"github.com/tcolgate/hugot"
)

type MessagePlayer struct {
	sync.Mutex
	Messages []*hugot.Message
}

func (mp *MessagePlayer) Receive() <-chan *hugot.Message {
	c := make(chan *hugot.Message)

	go func() {
		defer close(c)
		for {
			mp.Lock()
			if len(mp.Messages) == 0 {
				mp.Unlock()
				return
			}

			m := mp.Messages[0]
			mp.Messages = mp.Messages[1:]
			mp.Unlock()

			c <- m
		}
	}()
	return c
}
