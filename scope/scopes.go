// Package scope describes hugot's mesage scoping. This can be used
// for various purpose.
// - Store data that is user or channel specific (see handlers/alias)
// - Define which users have rights to perform cert actions (see handlers/roles)
package scope

import "fmt"

// Scope represents a scope for a property
type Scope int

const (
	// Unknown  invalid scope
	Unknown Scope = iota
	// Global applies to all users
	Global
	// Channel applies to all users in the current channel
	Channel
	// User applues to one user
	User
	// ChannelUser applies to one user in and only in the current channel
	ChannelUser
)

// String implements String
func (s Scope) String() string {
	switch s {
	case Global:
		return "Global"
	case Channel:
		return "Channel"
	case ChannelUser:
		return "Channel+User"
	case User:
		return "User"
	default:
		return fmt.Sprintf("Scope(%d)", s)
	}
}

// Order is a predefined order to search scopes
var Order = []Scope{
	ChannelUser,
	User,
	Channel,
	Global,
}

// Describe provides a human readable description of
// a scope.
func (s Scope) Describe(channel, user string) string {
	switch s {
	case Global:
		return fmt.Sprintf("set globally")
	case Channel:
		return fmt.Sprintf("for channel %s", channel)
	case ChannelUser:
		return fmt.Sprintf("for user %s in channel %s", user, channel)
	case User:
		return fmt.Sprintf("for user %s", user)
	default:
		return fmt.Sprintf("in unknown scope Scope(%d)", s)
	}
}

// Key returns the key for a the scope
func (s Scope) Key(channel, user string) string {
	switch s {
	case Global:
		return fmt.Sprintf("global")
	case Channel:
		return fmt.Sprintf("channel(%q)", channel)
	case ChannelUser:
		return fmt.Sprintf("channelUser(%q,%q)", channel, user)
	case User:
		return fmt.Sprintf("user(%q)", user)
	default:
		return fmt.Sprintf("scope(%d)", s)
	}
}
