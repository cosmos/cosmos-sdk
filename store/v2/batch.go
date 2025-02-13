package store

// Batch is a write-only database that commits changes to the underlying database
// when Write is called. A batch cannot be used concurrently.
type Batch interface {
	// Set inserts the given value into the key-value data store.
	//
	// Note: <key, value> are safe to modify and read after calling Set.
	Set(storeKey, key, value []byte) error

	// Delete removes the key from the backing key-value data store.
	//
	// Note: <key> is safe to modify and read after calling Delete.
	Delete(storeKey, key []byte) error

	// Size retrieves the amount of data queued up for writing, this includes
	// the keys, values, and deleted keys.
	Size() int

	// Write flushes any accumulated data to disk.
	Write() error

	// Reset resets the batch.
	Reset() error
}
