package memstore

import (
	"errors"

	"cosmossdk.io/store/types"
)

var (
	_ types.TypedMemStore[any]         = &typedMemStore[any]{}
	_ types.TypedMemStoreIterator[any] = &typedMemStoreIterator[any]{}
)

// typedMemStore provides a type-safe wrapper around a MemStore
type typedMemStore[T any] struct {
	types.MemStore
}

// NewTypedMemStore creates a new TypedMemStore with the specified type parameter
func NewTypedMemStore[T any](ms types.MemStore) types.TypedMemStore[T] {
	return &typedMemStore[T]{
		MemStore: ms,
	}
}

// Branch creates a new typed memory store that can be committed back to the parent
func (t *typedMemStore[T]) Branch() types.TypedMemStore[T] {
	return NewTypedMemStore[T](t.MemStore.Branch())
}

// Get retrieves a value for the given key and converts it to type T
func (t *typedMemStore[T]) Get(key []byte) T {
	val := t.MemStore.Get(key)
	if val == nil {
		var zero T
		return zero
	}

	return val.(T)
}

// Set adds or updates a key-value pair
func (t *typedMemStore[T]) Set(key []byte, value T) {
	t.MemStore.Set(key, value)
}

// Delete removes a key from the store
func (t *typedMemStore[T]) Delete(key []byte) {
	t.MemStore.Delete(key)
}

// Commit applies the changes in the current store to its parent
func (t *typedMemStore[T]) Commit() {
	t.MemStore.Commit()
}

// Iterator returns an iterator over the key-value pairs within the specified range
func (t *typedMemStore[T]) Iterator(start, end []byte) types.TypedMemStoreIterator[T] {
	return newTypedMemStoreIterator[T](t.MemStore.Iterator(start, end))
}

// ReverseIterator returns an iterator over the key-value pairs in reverse order
func (t *typedMemStore[T]) ReverseIterator(start, end []byte) types.TypedMemStoreIterator[T] {
	return newTypedMemStoreIterator[T](t.MemStore.ReverseIterator(start, end))
}

// typedMemStoreIterator provides a type-safe iterator for TypedMemStore
type typedMemStoreIterator[T any] struct {
	iter types.MemStoreIterator
}

// newTypedMemStoreIterator creates a new typed iterator from a MemStoreIterator
func newTypedMemStoreIterator[T any](iter types.MemStoreIterator) types.TypedMemStoreIterator[T] {
	return &typedMemStoreIterator[T]{
		iter: iter,
	}
}

// Domain returns the start and end keys defining the range of this iterator
func (ti *typedMemStoreIterator[T]) Domain() ([]byte, []byte) {
	return ti.iter.Domain()
}

// Valid returns whether the iterator is positioned at a valid item
func (ti *typedMemStoreIterator[T]) Valid() bool {
	return ti.iter.Valid()
}

// Next moves the iterator to the next item
func (ti *typedMemStoreIterator[T]) Next() {
	ti.iter.Next()
}

// Key returns the current key
func (ti *typedMemStoreIterator[T]) Key() []byte {
	return ti.iter.Key()
}

// Value returns the current value as type T
func (ti *typedMemStoreIterator[T]) Value() T {
	val := ti.iter.Value()
	if val == nil {
		var zero T
		return zero
	}

	return val.(T)
}

// Close releases any resources associated with the iterator
func (ti *typedMemStoreIterator[T]) Close() error {
	return ti.iter.Close()
}

// Error returns an error if the iterator is invalid
func (ti *typedMemStoreIterator[T]) Error() error {
	if !ti.Valid() {
		return errors.New("invalid typedMemStoreIterator")
	}
	return nil
}
