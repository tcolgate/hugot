package memory

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
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

func keyToPath(key []string) string {
	if len(key) == 0 {
		return ""
	}
	str := url.QueryEscape(key[0])
	for i := range key[1:] {
		str += "/" + url.QueryEscape(key[1+i])
	}
	return str
}

func pathToKey(path string) []string {
	parts := strings.Split(path, "/")
	key := []string{}
	for i := range parts {
		ki, err := url.QueryUnescape(parts[i])
		if err != nil {
			panic(fmt.Errorf("invalid key path path"))
		}

		key = append(key, ki)
	}
	return key
}

// Get retries a key from the store
func (s *Store) Get(key []string) (string, bool, error) {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[keyToPath(key)]
	return v, ok, nil
}

// List all items under the provided prefix
func (s *Store) List(key []string) ([][]string, error) {
	s.RLock()
	defer s.RUnlock()

	pfx := keyToPath(key)

	ks := [][]string{}
	for k := range s.data {
		if strings.HasPrefix(string(k), pfx) {
			ks = append(ks, pathToKey(string(k[len(key):])))
		}
	}
	return ks, nil
}

// Set a key in the store
func (s *Store) Set(key []string, value string) error {
	s.Lock()
	defer s.Unlock()

	s.data[keyToPath(key)] = value

	return nil
}

// Unset a key in the store
func (s *Store) Unset(key []string) error {
	s.Lock()
	defer s.Unlock()

	delete(s.data, keyToPath(key))
	return nil
}
