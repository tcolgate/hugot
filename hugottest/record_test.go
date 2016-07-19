package hugottest

import "testing"

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
}
