package prefix

import (
	"bytes"
	"errors"

	"cosmossdk.io/store/memstore"
	"cosmossdk.io/store/types"
)

// prefixMemStore integrates prefix functionality
type prefixMemStore struct {
	parent types.MemStore
	prefix []byte
}

// NewTypedMemStore creates a prefix memory store that supports generic type T
func NewTypedMemStore[T any](ms types.MemStore, prefix []byte) types.TypedMemStore[T] {
	prefixed := NewMemStore(ms, prefix)
	return memstore.NewTypedMemStore[T](prefixed)
}

// NewMemStore creates a prefix memory store
func NewMemStore(parent types.MemStore, prefix []byte) types.MemStore {
	return &prefixMemStore{
		parent: parent,
		prefix: prefix,
	}
}

// key prefixes the given key with the store's prefix
func (b *prefixMemStore) key(key []byte) (res []byte) {
	if key == nil {
		panic("nil key on PrefixMemStore")
	}
	res = cloneAppend(b.prefix, key)
	return
}

// Get retrieves a value for the given key
func (b *prefixMemStore) Get(key []byte) any {
	return b.parent.Get(b.key(key))
}

// Set adds or updates a key-value pair
func (b *prefixMemStore) Set(key []byte, value any) {
	b.parent.Set(b.key(key), value)
}

// Delete removes a key
func (b *prefixMemStore) Delete(key []byte) {
	b.parent.Delete(b.key(key))
}

// Commit applies the changes in the current batch
func (b *prefixMemStore) Commit() {
	b.parent.Commit()
}

// Branch creates a nested branch
func (b *prefixMemStore) Branch() types.MemStore {
	return &prefixMemStore{
		parent: b.parent.Branch(),
		prefix: b.prefix,
	}
}

// Iterator returns an iterator over the key-value pairs within the specified range
func (b *prefixMemStore) Iterator(start, end []byte) types.MemStoreIterator {
	var newStart, newEnd []byte

	if start == nil {
		newStart = b.prefix
	} else {
		newStart = cloneAppend(b.prefix, start)
	}

	if end == nil {
		newEnd = cpIncr(b.prefix)
	} else {
		newEnd = cloneAppend(b.prefix, end)
	}

	iter := b.parent.Iterator(newStart, newEnd)

	return newPrefixMemStoreIterator(b.prefix, start, end, iter)
}

// ReverseIterator returns an iterator over the key-value pairs in reverse order
func (b *prefixMemStore) ReverseIterator(start, end []byte) types.MemStoreIterator {
	var newStart, newEnd []byte

	if start == nil {
		newStart = b.prefix
	} else {
		newStart = cloneAppend(b.prefix, start)
	}

	if end == nil {
		newEnd = cpIncr(b.prefix)
	} else {
		newEnd = cloneAppend(b.prefix, end)
	}

	iter := b.parent.ReverseIterator(newStart, newEnd)

	return newPrefixMemStoreIterator(b.prefix, start, end, iter)
}

// prefixMemStoreIterator is an iterator with prefix support
type prefixMemStoreIterator struct {
	prefix []byte
	start  []byte
	end    []byte
	iter   types.MemStoreIterator
	valid  bool
}

// newPrefixMemStoreIterator creates a new prefix iterator
func newPrefixMemStoreIterator(prefix, start, end []byte, parent types.MemStoreIterator) *prefixMemStoreIterator {
	valid := parent.Valid() && bytes.HasPrefix(parent.Key(), prefix)
	return &prefixMemStoreIterator{
		prefix: prefix,
		start:  start,
		end:    end,
		iter:   parent,
		valid:  valid,
	}
}

// Domain returns the start and end keys
func (pi *prefixMemStoreIterator) Domain() ([]byte, []byte) {
	return pi.start, pi.end
}

// Valid returns whether the iterator is positioned at a valid item
func (pi *prefixMemStoreIterator) Valid() bool {
	return pi.valid && pi.iter.Valid()
}

// Next moves to the next item
func (pi *prefixMemStoreIterator) Next() {
	if !pi.valid {
		panic("prefixIterator invalid, cannot call Next()")
	}

	pi.iter.Next()
	if !pi.iter.Valid() || !bytes.HasPrefix(pi.iter.Key(), pi.prefix) {
		pi.valid = false
	}
}

// Key returns the current key with the prefix stripped
func (pi *prefixMemStoreIterator) Key() []byte {
	if !pi.valid {
		panic("prefixIterator invalid, cannot call Key()")
	}

	key := pi.iter.Key()
	return stripPrefix(key, pi.prefix)
}

// Value returns the current value
func (pi *prefixMemStoreIterator) Value() any {
	if !pi.valid {
		return nil
	}

	return pi.iter.Value()
}

// Close releases resources
func (pi *prefixMemStoreIterator) Close() error {
	return pi.iter.Close()
}

// Error returns an error if the iterator is invalid
func (pi *prefixMemStoreIterator) Error() error {
	if !pi.Valid() {
		return errors.New("invalid prefixIterator")
	}

	return nil
}
