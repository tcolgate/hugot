// Package etcd implemets a hugot store over an etcd datastore
package etcd

import (
	"time"

	redis "gopkg.in/redis.v5"

	"github.com/golang/glog"
	"github.com/tcolgate/hugot/storage"
)

type Store struct {
}

func New(opts *redis.Options) *Store {
	return &Store{}
}

// Get retries a key from the store
func (s *Store) Get(key []string) (string, bool, error) {
	val, err := s.cli.Get(storage.PathToKey(key)).Result()
	if err != nil {
		glog.Info(err)
		return val, false, err
	}

	return val, true, err
}

// List all items under the provided prefix
func (s *Store) List(path []string) ([][]string, error) {
	var paths [][]string
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = s.cli.Scan(cursor, storage.PathToKey(path)+"/*", 100).Result()
		if err != nil {
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
	return s.cli.Set(storage.PathToKey(path), value, 1000000*time.Hour).Err()
}

// Unset a key in the store
func (s *Store) Unset(path []string) error {
	return s.cli.Del(storage.PathToKey(path)).Err()
}
