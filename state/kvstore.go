package state

import (
	"sort"

	"github.com/tendermint/go-wire/data"
)

// KVStore is a simple interface to get/set data
type KVStore interface {
	Set(key, value []byte)
	Get(key []byte) (value []byte)
}

//----------------------------------------

// Model grabs together key and value to allow easier return values
type Model struct {
	Key   data.Bytes
	Value data.Bytes
}

// SimpleDB allows us to do some basic range queries on a db
type SimpleDB interface {
	KVStore

	Has(key []byte) (has bool)
	Remove(key []byte) (value []byte) // returns old value if there was one

	// Start is inclusive, End is exclusive...
	// Thus List ([]byte{12, 13}, []byte{12, 14}) will return anything with
	// the prefix []byte{12, 13}
	List(start, end []byte, limit int) []Model
	First(start, end []byte) Model
	Last(start, end []byte) Model

	// Checkpoint returns the same state, but where writes
	// are buffered and don't affect the parent
	Checkpoint() SimpleDB

	// Commit will take all changes from the checkpoint and write
	// them to the parent.
	// Returns an error if this is not a child of this one
	Commit(SimpleDB) error

	// Discard will remove reference to this
	Discard()
}

//----------------------------------------

// MemKVStore is a simple implementation of SimpleDB.
// It is only intended for quick testing, not to be used
// in production or with large data stores.
type MemKVStore struct {
	m map[string][]byte
}

var _ SimpleDB = NewMemKVStore()

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

func (m *MemKVStore) List(start, end []byte, limit int) []Model {
	keys := m.keysInRange(start, end)
	if limit > 0 && len(keys) > 0 {
		if limit > len(keys) {
			limit = len(keys)
		}
		keys = keys[:limit]
	}

	res := make([]Model, len(keys))
	for i, k := range keys {
		res[i] = Model{
			Key:   []byte(k),
			Value: m.m[k],
		}
	}
	return res
}

// First iterates through all keys to find the one that matches
func (m *MemKVStore) First(start, end []byte) Model {
	key := ""
	for _, k := range m.keysInRange(start, end) {
		if key == "" || k < key {
			key = k
		}
	}
	if key == "" {
		return Model{}
	}
	return Model{
		Key:   []byte(key),
		Value: m.m[key],
	}
}

func (m *MemKVStore) Last(start, end []byte) Model {
	key := ""
	for _, k := range m.keysInRange(start, end) {
		if key == "" || k > key {
			key = k
		}
	}
	if key == "" {
		return Model{}
	}
	return Model{
		Key:   []byte(key),
		Value: m.m[key],
	}
}

func (m *MemKVStore) Discard() {
	m.m = make(map[string][]byte, 0)
}

func (m *MemKVStore) Checkpoint() SimpleDB {
	return NewMemKVCache(m)
}

func (m *MemKVStore) Commit(sub SimpleDB) error {
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
