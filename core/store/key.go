package store

import "context"

// Key represents a unique, non-forgeable handle to a KVStore.
type Key interface {
	// Open retrieves the KVStore from the context.
	Open(context.Context) KVStore
}

// KVStoreKey represents a unique, non-forgeable handle to a regular merkle-tree
// backed KVStore. It should be provided as a module-scoped dependency by the runtime
// module being used to build the app.
type KVStoreKey interface {
	Key

	// IsKVStoreKey marks this interface as a regular kv-store key.
	IsKVStoreKey()
}

// MemoryStoreKey represents a unique, non-forgeable handle to a memory-backed
// KVStore. It should be provided as a module-scoped dependency by the runtime
// module being used to build the app.
type MemoryStoreKey interface {
	Key

	// IsMemoryStoreKey marks this interface as a memory store key.
	IsMemoryStoreKey()
}

// TransientStoreKey represents a unique, non-forgeable handle to a memory-backed
// KVStore which is reset at the start of every block. It should be provided as
// a module-scoped dependency by the runtime module being used to build the app.
type TransientStoreKey interface {
	Key

	// IsTransientStoreKey marks this interface as a transient store key.
	IsTransientStoreKey()
}
