package store

import (
	corestore "cosmossdk.io/server/v2/core/store"
)

// KVPair defines a key-value pair with additional metadata that is used to
// track writes. Deletion can be denoted by a nil value or explicitly by the
// Delete field.
type KVPair struct {
	Key      []byte
	Value    []byte
	StoreKey string // Optional for snapshot restore
}

type KVPairs []KVPair

func NewChangeset() *corestore.ChangeSet {
	return &corestore.ChangeSet{
		Pairs: make(map[string]KVPairs),
	}
}

func NewChangesetWithPairs(pairs map[string]KVPairs) *corestore.ChangeSet {
	return &corestore.ChangeSet{
		Pairs: pairs,
	}
}

// Size returns the number of key-value pairs in the batch.
func (cs *Changeset) Size() int {
	cnt := 0
	for _, pairs := range cs.Pairs {
		cnt += len(pairs)
	}

	return cnt
}

// Add adds a key-value pair to the ChangeSet.
func (cs *Changeset) Add(storeKey string, key, value []byte) {
	cs.Pairs[storeKey] = append(cs.Pairs[storeKey], KVPair{
		Key:      key,
		Value:    value,
		StoreKey: storeKey,
	})
}

// AddKVPair adds a KVPair to the ChangeSet.
func (cs *Changeset) AddKVPair(storeKey string, pair KVPair) {
	cs.Pairs[storeKey] = append(cs.Pairs[storeKey], pair)
}

// Merge merges the provided Changeset argument into the receiver. This may be
// useful when you have a Changeset that only pertains to a single store key,
// i.e. a map of size one, and you want to merge it into another.
func (cs *Changeset) Merge(other *Changeset) {
	for storeKey, pairs := range other.Pairs {
		cs.Pairs[storeKey] = append(cs.Pairs[storeKey], pairs...)
	}
}
