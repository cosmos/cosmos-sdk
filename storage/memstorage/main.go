/*
package memstorage provides a simple in-memory key store designed for
use in test cases, particularly to isolate them from the filesystem,
concurrency, and cleanup issues.
*/
package memstorage

import (
	"github.com/pkg/errors"
	keys "github.com/tendermint/go-keys"
)

type data struct {
	info keys.KeyInfo
	key  []byte
}

type MemStore map[string]data

// New creates an instance of file-based key storage with tight permissions
func New() MemStore {
	return MemStore{}
}

// assertStorage just makes sure we implement the Storage interface
func (s MemStore) assertStorage() keys.Storage {
	return s
}

// Put adds the given key, returns an error if it another key
// is already stored under this name
func (s MemStore) Put(name string, key []byte, info keys.KeyInfo) error {
	if _, ok := s[name]; ok {
		return errors.Errorf("Key named '%s' already exists", name)
	}
	s[name] = data{info, key}
	return nil
}

// Get returns the key stored under the name, or returns an error if not present
func (s MemStore) Get(name string) ([]byte, keys.KeyInfo, error) {
	var err error
	d, ok := s[name]
	if !ok {
		err = errors.Errorf("Key named '%s' doesn't exist", name)
	}
	return d.key, d.info, err
}

// List returns the public info of all keys in the MemStore in unsorted order
func (s MemStore) List() ([]keys.KeyInfo, error) {
	res := make([]keys.KeyInfo, len(s))
	i := 0
	for _, d := range s {
		res[i] = d.info
		i++
	}
	return res, nil
}

// Delete removes the named key from the MemStore, raising an error if it
// wasn't present yet.
func (s MemStore) Delete(name string) error {
	_, ok := s[name]
	if !ok {
		return errors.Errorf("Key named '%s' doesn't exist", name)
	}
	delete(s, name)
	return nil
}
