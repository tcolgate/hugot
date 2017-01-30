package command

import (
	"fmt"

	"github.com/tcolgate/hugot"
)

// ScopeFlags can be used to add a set of cli flags to a message to aide
// the user in selecting a scope.
type ScopeFlags struct {
	g  *bool
	c  *bool
	u  *bool
	cu *bool
}

// AddScopeFlags populate a set of flags to let a user select a scope
func AddScopeFlags(m *Message) *ScopeFlags {
	return &ScopeFlags{
		g:  m.Bool("g", false, "Create alias globally for all users on all channels"),
		c:  m.Bool("c", false, "Create alias for current channel only"),
		u:  m.Bool("u", false, "Create alias private for your user only"),
		cu: m.Bool("cu", false, "Create alias private for your user, only on this channel"),
	}
}

// Scope find which scope the user has selected
func (sf *ScopeFlags) Scope() (hugot.Scope, error) {
	switch {
	case *sf.g && !*sf.c && !*sf.u && !*sf.cu:
		return hugot.ScopeGlobal, nil
	case !*sf.g && *sf.c && !*sf.u && !*sf.cu:
		return hugot.ScopeChannel, nil
	case !*sf.g && !*sf.c && *sf.u && !*sf.cu:
		return hugot.ScopeUser, nil
	case !*sf.g && !*sf.c && !*sf.u && *sf.cu:
		return hugot.ScopeChannelUser, nil
	default:
		return hugot.ScopeUnknown, fmt.Errorf("Specify exactly one of -g, -c, -cu or -u")
	}
}

// Global has the global scope been selected
func (sf *ScopeFlags) Global() bool {
	return *sf.g
}

// Channel test if the the channel scope been selected
func (sf *ScopeFlags) Channel() bool {
	return *sf.c
}

// ChannelUser test if the channeluser scope been selected
func (sf *ScopeFlags) ChannelUser() bool {
	return *sf.cu
}

// User est if the user scope been selected
func (sf *ScopeFlags) User() bool {
	return *sf.u
}
