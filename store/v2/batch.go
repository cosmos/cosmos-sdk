package store

// Batch is a write-only database that commits changes to the underlying database
// when Write is called. A batch cannot be used concurrently.
type Batch interface {
	Writer

	// Size retrieves the amount of data queued up for writing, this includes
	// the keys, values, and deleted keys.
	Size() int

	// Write flushes any accumulated data to disk.
	Write() error

	// Reset resets the batch.
	Reset() error
}

// RawBatch represents a group of writes. They may or may not be written atomically depending on the
// backend. Callers must call Close on the batch when done.
//
// As with RawDB, given keys and values should be considered read-only, and must not be modified after
// passing them to the batch.
type RawBatch interface {
	// Set sets a key/value pair.
	// CONTRACT: key, value readonly []byte
	Set(key, value []byte) error

	// Delete deletes a key/value pair.
	// CONTRACT: key readonly []byte
	Delete(key []byte) error

	// Write writes the batch, possibly without flushing to disk. Only Close() can be called after,
	// other methods will error.
	Write() error

	// WriteSync writes the batch and flushes it to disk. Only Close() can be called after, other
	// methods will error.
	WriteSync() error

	// Close closes the batch. It is idempotent, but calls to other methods afterwards will error.
	Close() error

	// GetByteSize that returns the current size of the batch in bytes. Depending on the implementation,
	// this may return the size of the underlying LSM batch, including the size of additional metadata
	// on top of the expected key and value total byte count.
	GetByteSize() (int, error)
}
