package types

import (
	"io"

	"cosmossdk.io/store/v2"
)

// StoreType defines a type of KVStore.
type StoreType int

// KVStore defines the core storage primitive for modules to read and write state.
type KVStore interface {
	// GetStoreType returns the concrete store type.
	GetStoreType() StoreType

	// Get returns a value for a given key from the store.
	Get(key []byte) []byte

	// Has checks if a key exists.
	Has(key []byte) bool

	// Set sets a key/value entry to the store.
	Set(key, value []byte)

	// Delete deletes the key from the store.
	Delete(key []byte)

	CacheWrapper

	// Iterator creates a new Iterator over the domain [start, end). Note:
	//
	// - Start must be less than end
	// - The iterator must be closed by caller
	// - To iterate over entire domain, use store.Iterator(nil, nil)
	//
	// CONTRACT: No writes may happen within a domain while an iterator exists over
	// it, with the exception of a branched/cached KVStore.
	Iterator(start, end []byte) store.Iterator

	// ReverseIterator creates a new reverse Iterator over the domain [start, end).
	// It has the some properties and contracts as Iterator.
	ReverseIterator(start, end []byte) store.Iterator
}

// CacheWrapper defines an interface for creating a CacheWrap from a KVStore.
type CacheWrapper interface {
	CacheWrap() CacheWrap

	// CacheWrapWithTrace branches a store with tracing enabled.
	CacheWrapWithTrace(w io.Writer, tc TraceContext) CacheWrap
}

// CacheWrap defines an interface for branching a KVStore's state, allowing writes
// to be cached and flushed to the underlying store or discarded completed. Reads
// should be checked against a cache before querying the underlying store upon a
// cache miss. A CacheWrap store allows for nested branching.
type CacheWrap interface {
	// Write flushes writes to the underlying store.
	Write()

	// CacheWrap recursively wraps.
	CacheWrap() CacheWrap

	// CacheWrapWithTrace recursively wraps with tracing enabled.
	CacheWrapWithTrace(w io.Writer, tc TraceContext) CacheWrap
}
