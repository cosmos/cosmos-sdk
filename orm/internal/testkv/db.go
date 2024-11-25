package testkv

import (
	"cosmossdk.io/core/store"
)

type TestStore struct {
	Db store.KVStoreWithBatch
}

func (ts TestStore) Get(bz []byte) ([]byte, error) {
	return ts.Db.Get(bz)
}

// Has checks if a key exists.
func (ts TestStore) Has(key []byte) (bool, error) {
	return ts.Db.Has(key)
}

func (ts TestStore) Set(k, v []byte) error {
	return ts.Db.Set(k, v)
}

// Delete deletes the key, or does nothing if the key does not exist.
// CONTRACT: key readonly []byte
func (ts TestStore) Delete(bz []byte) error {
	return ts.Db.Delete(bz)
}

func (ts TestStore) Iterator(start, end []byte) (store.Iterator, error) {
	itr, err := ts.Db.Iterator(start, end)
	return IteratorWrapper{itr: itr}, err
}

func (ts TestStore) ReverseIterator(start, end []byte) (store.Iterator, error) {
	itr, err := ts.Db.ReverseIterator(start, end)
	return itr, err
}

// Close closes the database connection.
func (ts TestStore) Close() error {
	return ts.Db.Close()
}

// NewBatch creates a batch for atomic updates. The caller must call Batch.Close.
func (ts TestStore) NewBatch() store.Batch {
	return ts.Db.NewBatch()
}

// NewBatchWithSize create a new batch for atomic updates, but with pre-allocated size.
// This will does the same thing as NewBatch if the batch implementation doesn't support pre-allocation.
func (ts TestStore) NewBatchWithSize(i int) store.Batch {
	return ts.Db.NewBatchWithSize(i)
}

var _ store.Iterator = IteratorWrapper{}

type IteratorWrapper struct {
	itr store.Iterator
}

// Domain returns the start (inclusive) and end (exclusive) limits of the iterator.
// CONTRACT: start, end readonly []byte
func (iw IteratorWrapper) Domain() (start, end []byte) {
	return iw.itr.Domain()
}

// Valid returns whether the current iterator is valid. Once invalid, the Iterator remains
// invalid forever.
func (iw IteratorWrapper) Valid() bool {
	return iw.itr.Valid()
}

// Next moves the iterator to the next key in the database, as defined by order of iteration.
// If Valid returns false, this method will panic.
func (iw IteratorWrapper) Next() {
	iw.itr.Next()
}

// Key returns the key at the current position. Panics if the iterator is invalid.
// CONTRACT: key readonly []byte
func (iw IteratorWrapper) Key() (key []byte) {
	return iw.itr.Key()
}

// Value returns the value at the current position. Panics if the iterator is invalid.
// CONTRACT: value readonly []byte
func (iw IteratorWrapper) Value() (value []byte) {
	return iw.itr.Value()
}

// Error returns the last error encountered by the iterator, if any.
func (iw IteratorWrapper) Error() error {
	return iw.itr.Error()
}

// Close closes the iterator, releasing any allocated resources.
func (iw IteratorWrapper) Close() error {
	return iw.itr.Close()
}
