package storage

import "github.com/tcolgate/hugot/storage/memory"

// Storer is an interface describing our KV store
// requirements
type Storer interface {
	Get(key []string) (string, bool, error)
	List(key []string) ([][]string, error)
	Set(key []string, value string) error
	Unset(key []string) error
}

var (
	//DefaultStore is the default storage used by all handlers.
	DefaultStore = memory.New()
)

// PrefixStore wraps a Storer, and transparenttly
// appends and removes a prefix
type PrefixStore struct {
	sep  string
	pfx  []string
	base Storer
}

// NewPrefixedStore creates a store than preprends your
// provided prefix to store keys (with a # separator)
func NewPrefixedStore(s Storer, pfx []string) PrefixStore {
	return PrefixStore{
		sep:  ".",
		pfx:  append([]string(pfx), []string{"."}...),
		base: s,
	}
}

// Get retrieves a key from the store.
func (p PrefixStore) Get(key []string) (string, bool, error) {
	return p.base.Get(append(p.pfx, key...))
}

// List lists all the keys with the given prefix
func (p PrefixStore) List(key []string) ([][]string, error) {
	return p.base.List(append(p.pfx, key...))
}

// Set sets  a key in the store.
func (p PrefixStore) Set(key []string, value string) error {
	return p.base.Set(append(p.pfx, key...), value)
}

// Unset removes a key from the store
func (p PrefixStore) Unset(key []string) error {
	return p.base.Unset(append(p.pfx, key...))
}
