package hugot

import (
	"fmt"

	"github.com/tcolgate/hugot/storage"
)

// Scope represents a scope for a property
type Scope int

const (
	// ScopeUnknown  invalid scope
	ScopeUnknown Scope = iota
	// ScopeGlobal applies to all users
	ScopeGlobal
	// ScopeChannel applies to all users in the current channel
	ScopeChannel
	// ScopeUser applues to one user
	ScopeUser
	// ScopeChannelUser applies to one user in and only in the current channel
	ScopeChannelUser
)

var scopeOrder = []Scope{
	ScopeUnknown,
	ScopeChannelUser,
	ScopeUser,
	ScopeChannel,
	ScopeGlobal,
}

func (s Scope) keyFmt() {
}

// PropertyStore is used to store key values pairs that are
// dependent on a scope
type PropertyStore struct {
	store storage.Storer
	m     *Message
}

// NewPropertyStore uese the provided store to store properties,
// under a prefix pfx
func NewPropertyStore(s storage.Storer, m *Message) PropertyStore {
	return PropertyStore{s, m}
}

func propertyKey(m *Message, s Scope, k string) string {
	switch s {
	case ScopeGlobal:
		return fmt.Sprintf("global.%q", k)
	case ScopeChannel:
		return fmt.Sprintf("channel(%q).%q", m.Channel, k)
	case ScopeChannelUser:
		return fmt.Sprintf("channel(%q).user(%q).%q", m.Channel, m.From, k)
	case ScopeUser:
		return fmt.Sprintf("user(%q).%q", m.From, k)
	default:
		return fmt.Sprintf("%d().%q", s, k)
	}
}

// Set sets a property for the given scope, using the channel and
// and user details in the message provided in Message
func (ps PropertyStore) Set(s Scope, k, v string) error {
	return ps.store.Set([]string{propertyKey(ps.m, s, k)}, v)
}

// Unset sets a property in a given scope, using the channel and
// and user details in the message provided in Message
func (ps PropertyStore) Unset(s Scope, k string) error {
	return ps.store.Unset([]string{propertyKey(ps.m, s, k)})
}

// Lookup looks up a property. Scopes are searched in the following order:
//	 ScopeChanneUser
//	 ScopeUser
//	 ScopeChannel
//	 ScopeGlobal
func (ps PropertyStore) Lookup(k string) (string, bool, error) {
	for _, s := range scopeOrder {
		v, ok, err := ps.LookupInScope(s, k)
		if err != nil {
			return "", false, err
		}
		if ok {
			return v, ok, err
		}
	}
	return "", false, nil
}

// LookupInScope looks up the property for the given message, in the reuested scope
func (ps PropertyStore) LookupInScope(s Scope, k string) (string, bool, error) {
	bs, ok, err := ps.store.Get([]string{propertyKey(ps.m, s, k)})
	return string(bs), ok, err
}

// LookupAll searches all scopes for property k
func (ps PropertyStore) LookupAll(k string) (map[Scope]string, error) {
	res := map[Scope]string{}
	for _, s := range scopeOrder {
		v, ok, err := ps.LookupInScope(s, k)
		if err != nil {
			return nil, err
		}
		if ok {
			res[s] = v
		}
	}
	return res, nil
}
