package command

import (
	"bytes"
	"strings"

	shellwords "github.com/mattn/go-shellwords"
)

// Copy can be used to copy a command.Mesage
func (m *Message) Copy() *Message {
	nm := Message{Message: m.Message.Copy()}
	nm.args = nil
	nm.FlagSet = nil
	nm.FlagOut = &bytes.Buffer{}
	return &nm
}

// Args returns the arguments parsed from the message
func (m *Message) Args() []string {
	if m.args == nil {
		var err error
		m.args, err = shellwords.Parse(m.Text)
		if err != nil {
			m.args = strings.Split(m.Text, " ")
		}

	}
	return m.args
}

// SetArgs sets the arguments to a specific set of values
func (m *Message) SetArgs(args []string) {
	m.args = args
}

// Parse process any Args for this message in line with any flags that have
// been added to the message.
func (m *Message) Parse() error {
	var err error
	if m.args == nil {
		m.args, err = shellwords.Parse(m.Text)
	}
	if err != nil {
		return ErrBadCLI
	}

	if len(m.args) > 0 {
		err = m.FlagSet.Parse(m.args[1:])
		m.args = m.FlagSet.Args()
	}
	return err
}
