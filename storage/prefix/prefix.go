package prefix

import "github.com/tcolgate/hugot/storage"

// Store wraps a Storer, and transparenttly
// appends and removes a prefix
type Store struct {
	sep  string
	pfx  []string
	base storage.Storer
}

// New creates a store than preprends your
// provided prefix to store keys (with a # separator)
func New(s storage.Storer, pfx []string) Store {
	return Store{
		sep:  ".",
		pfx:  append([]string(pfx), []string{"."}...),
		base: s,
	}
}

// Get retrieves a key from the store.
func (p Store) Get(key []string) (string, bool, error) {
	return p.base.Get(append(p.pfx, key...))
}

// List lists all the keys with the given prefix
func (p Store) List(key []string) ([][]string, error) {
	return p.base.List(append(p.pfx, key...))
}

// Set sets  a key in the store.
func (p Store) Set(key []string, value string) error {
	return p.base.Set(append(p.pfx, key...), value)
}

// Unset removes a key from the store
func (p Store) Unset(key []string) error {
	return p.base.Unset(append(p.pfx, key...))
}
