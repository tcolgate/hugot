package scoped

import (
	"github.com/tcolgate/hugot/scope"
	"github.com/tcolgate/hugot/storage"
	"github.com/tcolgate/hugot/storage/prefix"
)

// Store implements a store which
type Store struct {
	storage.Storer
}

// New returns a new Storer that prefixes the passed in Store with a key
// defined by the scope as described by the scope and Message
func New(base storage.Storer, s scope.Scope, channel, user string) *Store {
	key := s.Key(channel, user)
	return &Store{prefix.New(base, []string{key})}
}

// Get retries a key from the store
func (s *Store) Get(key []string) (string, bool, error) {
	return s.Storer.Get(key)
}

// List all items under the provided prefix
func (s *Store) List(key []string) ([][]string, error) {
	return s.Storer.List(key)
}

// Set a key in the store
func (s *Store) Set(key []string, value string) error {
	return s.Storer.Set(key, value)
}

// Unset a key in the store
func (s *Store) Unset(key []string) error {
	return s.Storer.Unset(key)
}
