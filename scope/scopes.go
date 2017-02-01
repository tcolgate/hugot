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

// Order is a predefined order to search scopes
var Order = []Scope{
	ChannelUser,
	User,
	Channel,
	Global,
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
