package testkv

import (
	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/core/store"
)

type TestStore struct {
	db dbm.DB
}

func (ts TestStore) Get(bz []byte) ([]byte, error) {
	return ts.db.Get(bz)
}

// // Has checks if a key exists.
func (ts TestStore) Has(key []byte) (bool, error) {
	return ts.db.Has(key)
}

func (ts TestStore) Set(k []byte, v []byte) error {
	return ts.db.Set(k, v)
}

// // SetSync sets the value for the given key, and flushes it to storage before returning.
func (ts TestStore) SetSync(k []byte, v []byte) error {
	return ts.db.SetSync(k, v)
}

// // Delete deletes the key, or does nothing if the key does not exist.
// // CONTRACT: key readonly []byte
func (ts TestStore) Delete(bz []byte) error {
	return ts.db.Delete(bz)
}

// // DeleteSync deletes the key, and flushes the delete to storage before returning.
func (ts TestStore) DeleteSync(bz []byte) error {
	return ts.db.DeleteSync(bz)
}

func (ts TestStore) Iterator(start, end []byte) (store.Iterator, error) {
	return IteratorWrapper{itr: ts.db.Iterator}, nil
}

func (ts TestStore) ReverseIterator(start, end []byte) (store.Iterator, error) {
	return nil, nil
}

// Close closes the database connection.
func (ts TestStore) Close() error {
	return ts.db.Close()
}

// NewBatch creates a batch for atomic updates. The caller must call Batch.Close.
func (ts TestStore) NewBatch() dbm.Batch {
	return ts.db.NewBatch()
}

// NewBatchWithSize create a new batch for atomic updates, but with pre-allocated size.
// This will does the same thing as NewBatch if the batch implementation doesn't support pre-allocation.
func (ts TestStore) NewBatchWithSize(i int) dbm.Batch {
	return ts.db.NewBatchWithSize(i)
}

// Print is used for debugging.
func (ts TestStore) Print() error {
	return ts.db.Print()
}

// Stats returns a map of property values for all keys and the size of the cache.
func (ts TestStore) Stats() map[string]string {
	return ts.db.Stats()
}

var _ store.Iterator = IteratorWrapper{}

type IteratorWrapper struct {
	itr dbm.Iterator
}

// Domain returns the start (inclusive) and end (exclusive) limits of the iterator.
// CONTRACT: start, end readonly []byte
func (iw IteratorWrapper) Domain() (start []byte, end []byte) {
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

// Close closes the iterator, relasing any allocated resources.
func (iw IteratorWrapper) Close() error {
	return iw.itr.Close()
}
