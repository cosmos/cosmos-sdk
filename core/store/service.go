package store

import "context"

// Service represents a unique, non-forgeable handle to a KVStore.
type Service interface {
	// Open retrieves the KVStore from the context.
	Open(context.Context) KVStore
}

// KVStoreService represents a unique, non-forgeable handle to a regular merkle-tree
// backed KVStore. It should be provided as a module-scoped dependency by the runtime
// module being used to build the app.
type KVStoreService interface {
	Service

	// IsKVStoreService marks this interface as a regular kv-store service.
	IsKVStoreService()
}

// MemoryStoreService represents a unique, non-forgeable handle to a memory-backed
// KVStore. It should be provided as a module-scoped dependency by the runtime
// module being used to build the app.
type MemoryStoreService interface {
	Service

	// IsMemoryStoreService marks this interface as a memory store service.
	IsMemoryStoreService()
}

// TransientStoreService represents a unique, non-forgeable handle to a memory-backed
// KVStore which is reset at the start of every block. It should be provided as
// a module-scoped dependency by the runtime module being used to build the app.
type TransientStoreService interface {
	Service

	// IsTransientStoreService marks this interface as a transient store service.
	IsTransientStoreService()
}
