package state

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk"
)

//----------------------------------------

// MemKVStore is a simple implementation of sdk.SimpleDB.
// It is only intended for quick testing, not to be used
// in production or with large data stores.
type MemKVStore struct {
	m map[string][]byte
}

var _ sdk.SimpleDB = NewMemKVStore()

// NewMemKVStore initializes a MemKVStore
func NewMemKVStore() *MemKVStore {
	return &MemKVStore{
		m: make(map[string][]byte, 0),
	}
}

func (m *MemKVStore) Set(key []byte, value []byte) {
	m.m[string(key)] = value
}

func (m *MemKVStore) Get(key []byte) (value []byte) {
	return m.m[string(key)]
}

func (m *MemKVStore) Has(key []byte) (has bool) {
	_, ok := m.m[string(key)]
	return ok
}

func (m *MemKVStore) Remove(key []byte) (value []byte) {
	val := m.m[string(key)]
	delete(m.m, string(key))
	return val
}

func (m *MemKVStore) List(start, end []byte, limit int) []sdk.Model {
	keys := m.keysInRange(start, end)
	if limit > 0 && len(keys) > 0 {
		if limit > len(keys) {
			limit = len(keys)
		}
		keys = keys[:limit]
	}

	res := make([]sdk.Model, len(keys))
	for i, k := range keys {
		res[i] = sdk.Model{
			Key:   []byte(k),
			Value: m.m[k],
		}
	}
	return res
}

// First iterates through all keys to find the one that matches
func (m *MemKVStore) First(start, end []byte) sdk.Model {
	key := ""
	for _, k := range m.keysInRange(start, end) {
		if key == "" || k < key {
			key = k
		}
	}
	if key == "" {
		return sdk.Model{}
	}
	return sdk.Model{
		Key:   []byte(key),
		Value: m.m[key],
	}
}

func (m *MemKVStore) Last(start, end []byte) sdk.Model {
	key := ""
	for _, k := range m.keysInRange(start, end) {
		if key == "" || k > key {
			key = k
		}
	}
	if key == "" {
		return sdk.Model{}
	}
	return sdk.Model{
		Key:   []byte(key),
		Value: m.m[key],
	}
}

func (m *MemKVStore) Discard() {
	m.m = make(map[string][]byte, 0)
}

func (m *MemKVStore) Checkpoint() sdk.SimpleDB {
	return NewMemKVCache(m)
}

func (m *MemKVStore) Commit(sub sdk.SimpleDB) error {
	cache, ok := sub.(*MemKVCache)
	if !ok {
		return ErrNotASubTransaction()
	}

	// see if it points to us
	ref, ok := cache.store.(*MemKVStore)
	if !ok || ref != m {
		return ErrNotASubTransaction()
	}

	// apply the cached data to us
	cache.applyCache()
	return nil
}

func (m *MemKVStore) keysInRange(start, end []byte) (res []string) {
	s, e := string(start), string(end)
	for k := range m.m {
		afterStart := s == "" || k >= s
		beforeEnd := e == "" || k < e
		if afterStart && beforeEnd {
			res = append(res, k)
		}
	}
	sort.Strings(res)
	return
}
