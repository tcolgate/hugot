package hugottest

import (
	"context"
	"testing"

	"github.com/tcolgate/hugot"
)

func TestResponseRecorder(t *testing.T) {
	rr := &ResponseRecorder{}
	rr.SetChannel("test")
	rr.Write([]byte("Hello World"))

	if len(rr.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(rr.Messages))
	}

	if rr.Messages[0].Text != "Hello World" {
		t.Fatalf("Expected \"Hello World\" message, got %#v", rr.Messages[0].Text)
	}

	rr.SetChannel("test2")
	rr.Send(context.Background(), &hugot.Message{Text: "another message"})

	if len(rr.Messages) != 2 {
		t.Fatalf("Expected 2 message, got %d", len(rr.Messages))
	}

	if rr.Messages[1].Channel != "test2" {
		t.Fatalf("Expected channel \"test2\" in message, got %#v", rr.Messages[1].Channel)
	}

	if rr.Messages[1].Text != "another message" {
		t.Fatalf("Expected \"another message\" message, got %#v", rr.Messages[1].Text)
	}
}
