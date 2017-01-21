package memory

import (
	"bytes"
	"sync"
)

// MemStore implements a simple memory store over a map
// It is safe for concurrent access
type MemStore struct {
	sync.RWMutex
	data map[string][]byte
}

// Get retries a key from the store
func (m *MemStore) Get(key []byte) ([]byte, bool, error) {
	m.RLock()
	defer m.RUnlock()

	v, ok := m.data[string(key)]
	return v, ok, nil
}

// List all items under the provided prefix
func (m *MemStore) List(key []byte) ([][]byte, error) {
	m.RLock()
	defer m.RUnlock()

	ks := [][]byte{}
	for k := range m.data {
		if bytes.HasPrefix([]byte(k), key) {
			ks = append(ks, []byte(k[len(key):]))
		}
	}
	return ks, nil
}

// Set a key in the store
func (m *MemStore) Set(key []byte, value []byte) error {
	m.Lock()
	defer m.Unlock()

	m.data[string(key)] = value
	return nil
}

// Unset a key in the store
func (m *MemStore) Unset(key []byte) error {
	m.Lock()
	defer m.Unlock()

	delete(m.data, string(key))
	return nil
}

// New creates a new memory store baked by go map
func New() *MemStore {
	return &MemStore{
		sync.RWMutex{},
		make(map[string][]byte),
	}
}
