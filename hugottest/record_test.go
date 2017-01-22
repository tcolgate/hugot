package hugottest

import (
	"context"
	"testing"

	"github.com/tcolgate/hugot"
)

func TestResponseRecorder(t *testing.T) {
	rr := &ResponseRecorder{MessagesOut: make(chan hugot.Message, 1)}
	rr.SetChannel("test")
	rr.Write([]byte("Hello World"))

	var m hugot.Message
	select {
	case m = <-rr.MessagesOut:
	default:
		t.Fatalf("Expected to be able to read 1 message, got blocked")
	}

	if m.Text != "Hello World" {
		t.Fatalf("Expected \"Hello World\" message, got %#v", m.Text)
	}

	rr.SetChannel("test2")
	rr.Send(context.Background(), &hugot.Message{Text: "another message"})

	select {
	case m = <-rr.MessagesOut:
	default:
		t.Fatalf("Expected to be able to read 1 message, got blocked")
	}

	if m.Channel != "test2" {
		t.Fatalf("Expected channel \"test2\" in message, got %#v", m.Channel)
	}

	if m.Text != "another message" {
		t.Fatalf("Expected \"another message\" message, got %#v", m.Text)
	}
}
