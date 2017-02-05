package memory

import (
	"strings"
	"sync"

	"github.com/tcolgate/hugot/storage"
)

// Store implements a simple memory store over a map
// It is safe for concurrent access
type Store struct {
	sync.RWMutex
	data map[string]string
}

// New creates a new memory store baked by go map
func New() *Store {
	return &Store{
		sync.RWMutex{},
		make(map[string]string),
	}
}

// Get retries a key from the store
func (s *Store) Get(path []string) (string, bool, error) {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[storage.PathToKey(path)]
	return v, ok, nil
}

// List all items under the provided prefix
func (s *Store) List(path []string) ([][]string, error) {
	s.RLock()
	defer s.RUnlock()

	pfx := storage.PathToKey(path)

	ks := [][]string{}
	for k := range s.data {
		if strings.HasPrefix(string(k), pfx) {
			ks = append(ks, storage.KeyToPath(k))
		}
	}
	return ks, nil
}

// Set a key in the store
func (s *Store) Set(path []string, value string) error {
	s.Lock()
	defer s.Unlock()

	s.data[storage.PathToKey(path)] = value

	return nil
}

// Unset a key in the store
func (s *Store) Unset(path []string) error {
	s.Lock()
	defer s.Unlock()

	delete(s.data, storage.PathToKey(path))
	return nil
}
