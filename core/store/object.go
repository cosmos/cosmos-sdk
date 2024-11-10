package store

import "context"

// CacheService is a service that provides a cache for storing objects.
type DecodedCacheService interface {
	OpenCache(ctx context.Context) ObjectStore
}

// ObjectStore describes the basic interface for interacting with key-value stores.
// The object store stores decoded values in a cache that must be populated and cleared on each block.
type ObjectStore interface {
	// Get returns nil if key doesn't exist. Panics on nil key.
	Get(key []byte) (any, bool)

	// Set sets the key. Panics on nil key or value.
	Set(key []byte, value any)

	// Delete deletes the key. Panics on nil key.
	Delete(key []byte)
}
