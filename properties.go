package hugot

import (
	"github.com/tcolgate/hugot/scope"
	"github.com/tcolgate/hugot/storage"
	"github.com/tcolgate/hugot/storage/scoped"
)

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

// Set sets a property for the given scope, using the channel and
// and user details in the message provided in Message
func (ps PropertyStore) Set(s scope.Scope, k []string, v string) error {
	store := scoped.New(ps.store, s, ps.m.Channel, ps.m.From)
	return store.Set(k, v)
}

// Unset sets a property in a given scope, using the channel and
// and user details in the message provided in Message
func (ps PropertyStore) Unset(s scope.Scope, k []string) error {
	store := scoped.New(ps.store, s, ps.m.Channel, ps.m.From)
	return store.Unset(k)
}

// Get looks up a property. Scopes are searched in the following order:
//	 ChanneUser
//	 User
//	 Channel
//	 Global
func (ps PropertyStore) Get(k []string) (string, bool, error) {
	for _, s := range scope.Order {
		v, ok, err := ps.GetInScope(s, k)
		if err != nil {
			return "", false, err
		}
		if ok {
			return v, ok, err
		}
	}
	return "", false, nil
}

// GetInScope looks up the property for the given message, in the reuested scope
func (ps PropertyStore) GetInScope(s scope.Scope, k []string) (string, bool, error) {
	store := scoped.New(ps.store, s, ps.m.Channel, ps.m.From)
	return store.Get(k)
}

// LookupAll searches all scopes for property k
func (ps PropertyStore) LookupAll(k []string) (map[scope.Scope]string, error) {
	res := map[scope.Scope]string{}
	for _, s := range scope.Order {
		v, ok, err := ps.GetInScope(s, k)
		if err != nil {
			return nil, err
		}
		if ok {
			res[s] = v
		}
	}
	return res, nil
}
