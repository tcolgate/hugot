package memory

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
)

// MemStore implements a simple memory store over a map
// It is safe for concurrent access
type MemStore struct {
	sync.RWMutex
	data map[string]string
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
func (m *MemStore) Get(key []string) (string, bool, error) {
	m.RLock()
	defer m.RUnlock()

	v, ok := m.data[keyToPath(key)]
	return v, ok, nil
}

// List all items under the provided prefix
func (m *MemStore) List(key []string) ([][]string, error) {
	m.RLock()
	defer m.RUnlock()

	pfx := keyToPath(key)

	ks := [][]string{}
	for k := range m.data {
		if strings.HasPrefix(string(k), pfx) {
			ks = append(ks, pathToKey(string(k[len(key):])))
		}
	}
	return ks, nil
}

// Set a key in the store
func (m *MemStore) Set(key []string, value string) error {
	m.Lock()
	defer m.Unlock()

	m.data[keyToPath(key)] = value
	return nil
}

// Unset a key in the store
func (m *MemStore) Unset(key []string) error {
	m.Lock()
	defer m.Unlock()

	delete(m.data, keyToPath(key))
	return nil
}

// New creates a new memory store baked by go map
func New() *MemStore {
	return &MemStore{
		sync.RWMutex{},
		make(map[string]string),
	}
}
