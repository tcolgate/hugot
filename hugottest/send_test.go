package hugottest

import (
	"testing"

	"github.com/tcolgate/hugot"
)

func TestMessagePlayer(t *testing.T) {
	rp := MessagePlayer{
		Messages: []*hugot.Message{
			&hugot.Message{Text: "message1"},
			&hugot.Message{Text: "message2"},
		},
	}

	m := <-rp.Receive()
	if m == nil {
		t.Fatalf("did not get a message")
	}

	if m.Text != "message1" {
		t.Fatalf("expected \"message1\", got %#v", m.Text)
	}
}
