// Package etcd implemets a hugot store over an etcd datastore
package etcd

import (
	"context"

	"github.com/coreos/etcd/clientv3"

	"github.com/golang/glog"
	"github.com/tcolgate/hugot/storage"
)

type Store struct {
	cli *clientv3.Client
}

func New(cli *clientv3.Client) *Store {
	return &Store{cli}
}

// Get retries a key from the store
func (s *Store) Get(key []string) (string, bool, error) {
	val, err := s.cli.Get(context.Background(), storage.PathToKey(key))
	if err != nil {
		glog.Info(err)
		return "", false, err
	}

	if val.Count == 0 {
		return "", false, nil
	}

	return string(val.Kvs[0].Value), true, err
}

// List all items under the provided prefix
func (s *Store) List(path []string) ([][]string, error) {
	var err error
	var paths [][]string
	keys, err := s.cli.Get(context.Background(), storage.PathToKey(path), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	for _, k := range keys.Kvs {
		paths = append(paths, storage.KeyToPath(string(k.Key)))
	}
	return paths, nil
}

// Set a key in the store
func (s *Store) Set(path []string, value string) error {
	_, err := s.cli.Put(context.Background(), storage.PathToKey(path), value)
	return err
}

// Unset a key in the store
func (s *Store) Unset(path []string) error {
	_, err := s.cli.Delete(context.Background(), storage.PathToKey(path))
	return err
}
