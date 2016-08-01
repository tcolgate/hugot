package memory

import "sync"

type memStore struct {
	sync.RWMutex
	data map[string][]byte
}

func (m *memStore) Get(key []byte) ([]byte, bool, error) {
	m.RLock()
	defer m.RUnlock()

	v, ok := m.data[string(key)]
	return v, ok, nil
}

func (m *memStore) Set(key []byte, value []byte) error {
	m.Lock()
	defer m.Unlock()

	m.data[string(key)] = value
	return nil
}

func (m *memStore) Unset(key []byte) error {
	m.Lock()
	defer m.Unlock()

	delete(m.data, string(key))
	return nil
}

func New() *memStore {
	return &memStore{
		sync.RWMutex{},
		make(map[string][]byte),
	}
}
