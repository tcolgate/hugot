package redis

import (
	"github.com/tcolgate/hugot/storage"
	redis "gopkg.in/redis.v5"
)

type Store struct {
	cli *redis.Client
}

func New(opts *redis.Options) *Store {
	return &Store{
		cli: redis.NewClient(opts),
	}
}

// Get retries a key from the store
func (s *Store) Get(key []string) (string, bool, error) {
	exists, err := s.cli.Exists(storage.PathToKey(key)).Result()
	if err != redis.Nil && err != nil {
		return "", false, err
	}
	if !exists {
		return "", false, nil
	}

	val, err := s.cli.Get(storage.PathToKey(key)).Result()
	if err != redis.Nil && err != nil {
		return "", false, err
	}

	return val, true, nil
}

// List all items under the provided prefix
func (s *Store) List(path []string) ([][]string, error) {
	var paths [][]string
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = s.cli.Scan(cursor, storage.PathToKey(path)+"/*", 100).Result()
		if err != redis.Nil && err != nil {
			return nil, err
		}
		for _, key := range keys {
			paths = append(paths, storage.KeyToPath(key))
		}
		if cursor == 0 {
			break
		}
	}

	return paths, nil
}

// Set a key in the store
func (s *Store) Set(path []string, value string) error {
	err := s.cli.Set(storage.PathToKey(path), value, 0).Err()
	if err != redis.Nil && err != nil {
		return err
	}
	return nil
}

// Unset a key in the store
func (s *Store) Unset(path []string) error {
	err := s.cli.Del(storage.PathToKey(path)).Err()
	if err != redis.Nil && err != nil {
		return err
	}
	return nil
}
