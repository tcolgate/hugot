// Package storage describes an interface for storing of hugot
// handler data in a Key/Value store.
package storage

import (
	"fmt"
	"net/url"
	"strings"
)

// Storer is an interface describing our KV store
// requirements
type Storer interface {
	Get(key []string) (string, bool, error)
	List(key []string) ([][]string, error)
	Set(key []string, value string) error
	Unset(key []string) error
}

// PathToKey takes a storage path and translate it to a flat
// key usable in a KV store.
func PathToKey(path []string) string {
	if len(path) == 0 {
		return ""
	}
	str := url.QueryEscape(path[0])
	for i := range path[1:] {
		str += "/" + url.QueryEscape(path[1+i])
	}
	return str
}

// KeyToPath takes string from PathToKey and reverts it to a
// structured key.
func KeyToPath(key string) []string {
	parts := strings.Split(key, "/")
	var path []string
	for i := range parts {
		ki, err := url.QueryUnescape(parts[i])
		if err != nil {
			panic(fmt.Errorf("invalid key in storage"))
		}

		path = append(path, ki)
	}
	return path
}
