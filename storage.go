package hugot

import "github.com/tcolgate/hugot/storers/memory"

// Storer is an interface describing our KV store
// requirements
type Storer interface {
	Get(key []byte) ([]byte, bool, error)
	List(key []byte) ([][]byte, error)
	Set(key []byte, value []byte) error
	Unset(key []byte) error
}

// PrefixStore wraps a Storer, and transparenttly
// appends and removes a prefix
type PrefixStore struct {
	pfx  []byte
	base Storer
}

var (
	DefaultStore = memory.New()
)

// NewPrefixedStore creates a store than preprends your
// provided prefix to store keys (with a # separator)
func NewPrefixedStore(s Storer, pfx []byte) PrefixStore {
	return PrefixStore{
		pfx:  append(pfx, []byte("#")...),
		base: s,
	}
}

// Get retrieves a key from the store.
func (p PrefixStore) Get(key []byte) ([]byte, bool, error) {
	return p.base.Get(append(p.pfx, key...))
}

// List lists all the keys with the given prefix
func (p PrefixStore) List(key []byte) ([][]byte, error) {
	return p.base.List(append(p.pfx, key...))
}

// Set sets  a key in the store.
func (p PrefixStore) Set(key []byte, value []byte) error {
	return p.base.Set(append(p.pfx, key...), value)
}

// Unset removes a key from the store
func (p PrefixStore) Unset(key []byte) error {
	return p.base.Unset(append(p.pfx, key...))
}
