package command

import "errors"

var (
	// ErrSkipHears suggests that a handler has dealt with a command,
	// and than any subsequent Hears handlers should be skipped.
	ErrSkipHears = errors.New("skip hear messages")

	// ErrUnknownCommand is returned by a command mux if the command did
	// not match any of it's registered handlers.
	ErrUnknownCommand = errors.New("unknown command")

	// ErrBadCLI implies that we could not process this message as a
	// command line. E.g. due to potentially mismatched quoting or bad
	// escaping.
	ErrBadCLI = errors.New("could not process as command line")
)

// errUsage indicates that Command handler was used incorrectly. The
// string returned is a usage message generated by a call to -help
// for this command
type errUsage string

// Error implements the Error interface for an ErrUsage.
func (e errUsage) Error() string {
	return string(e)
}

// ErrUsage indicates that Command handler was used incorrectly. The
// string returned is a usage message generated by a call to -help
// for this command
func ErrUsage(s string) error {
	return errUsage(s)
}
