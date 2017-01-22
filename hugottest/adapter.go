package hugottest

import "github.com/tcolgate/hugot"

// Adapter provides a test adapter you can preload
// with messages, and retrieve messages from.
type Adapter struct {
	*ResponseRecorder
	*MessagePlayer
}

// NewAdapter creates a new Adapter, preloaded with the
// provided set of messages
func NewAdapter(in chan *hugot.Message, out chan hugot.Message) *Adapter {
	return &Adapter{
		&ResponseRecorder{MessagesOut: out},
		&MessagePlayer{MessagesIn: in},
	}
}
