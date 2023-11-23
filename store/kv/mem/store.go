package mem

import (
	"bytes"

	"github.com/tidwall/btree"

	"cosmossdk.io/store/v2"
)

// degree defines the approximate number of items and children per B-tree node.
const degree = 32

var _ store.KVStore = (*Store)(nil)

// Store defines an in-memory KVStore backed by a BTree for storage, indexing,
// and iteration. Note, the store is ephemeral and does not support commitment.
// If using the store between blocks or commitments, the caller must ensure to
// either create a new store or call Reset() on the existing store.
type Store struct {
	storeKey string
	tree     *btree.BTreeG[store.KVPair]
}

func New(storeKey string) store.KVStore {
	return &Store{
		storeKey: storeKey,
		tree: btree.NewBTreeGOptions(
			func(a, b store.KVPair) bool { return bytes.Compare(a.Key, b.Key) <= -1 },
			btree.Options{
				Degree:  degree,
				NoLocks: false,
			}),
	}
}

func (s *Store) GetStoreKey() string {
	return s.storeKey
}

func (s *Store) GetStoreType() store.StoreType {
	return store.StoreTypeMem
}

func (s *Store) Get(key []byte) []byte {
	store.AssertValidKey(key)

	kvPair, ok := s.tree.Get(store.KVPair{Key: key})
	if !ok || kvPair.Value == nil {
		return nil
	}

	return kvPair.Value
}

func (s *Store) Has(key []byte) bool {
	store.AssertValidKey(key)

	return s.Get(key) != nil
}

func (s *Store) Set(key, value []byte) {
	store.AssertValidKey(key)
	store.AssertValidValue(value)

	s.tree.Set(store.KVPair{Key: key, Value: value})
}

func (s *Store) Delete(key []byte) {
	store.AssertValidKey(key)

	s.tree.Set(store.KVPair{Key: key, Value: nil})
}

func (s *Store) GetChangeset() *store.Changeset {
	itr := s.Iterator(nil, nil)
	defer itr.Close()

	var kvPairs store.KVPairs
	for ; itr.Valid(); itr.Next() {
		kvPairs = append(kvPairs, store.KVPair{
			Key:   itr.Key(),
			Value: itr.Value(),
		})
	}

	return store.NewChangeset(map[string]store.KVPairs{s.storeKey: kvPairs})
}

func (s *Store) Reset(_ uint64) error {
	s.tree.Clear()
	return nil
}

func (s *Store) Iterator(start, end []byte) store.Iterator {
	return newIterator(s.tree, start, end, false)
}

func (s *Store) ReverseIterator(start, end []byte) store.Iterator {
	return newIterator(s.tree, start, end, true)
}
