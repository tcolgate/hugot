package hugot

import "github.com/tcolgate/hugot/storers/memory"

// Storer is an interface describing our KV store
// requirements
type Storer interface {
	Get(key string) (string, bool, error)
	List(key string) ([]string, error)
	Set(key string, value string) error
	Unset(key string) error
}

// PrefixStore wraps a Storer, and transparenttly
// appends and removes a prefix
type PrefixStore struct {
	pfx  string
	base Storer
}

var (
	DefaultStore = memory.New()
)

// NewPrefixedStore creates a store than preprends your
// provided prefix to store keys (with a # separator)
func NewPrefixedStore(s Storer, pfx string) PrefixStore {
	return PrefixStore{
		pfx:  string(append([]byte(pfx), []byte("#")...)),
		base: s,
	}
}

// Get retrieves a key from the store.
func (p PrefixStore) Get(key string) (string, bool, error) {
	return p.base.Get(string(append([]byte(p.pfx), key...)))
}

// List lists all the keys with the given prefix
func (p PrefixStore) List(key string) ([]string, error) {
	return p.base.List(string(append([]byte(p.pfx), key...)))
}

// Set sets  a key in the store.
func (p PrefixStore) Set(key string, value string) error {
	return p.base.Set(string(append([]byte(p.pfx), key...)), value)
}

// Unset removes a key from the store
func (p PrefixStore) Unset(key string) error {
	return p.base.Unset(string(append([]byte(p.pfx), key...)))
}
