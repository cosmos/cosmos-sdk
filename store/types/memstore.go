package types

type (
	// MemStoreManager defines the interface for a tree with batching capabilities.
	//
	// MemStoreManager defines the interface for managing the root (committed) state
	// of an in-memory B-Tree store and provides branching capabilities.
	// It ensures that only one top-level branch (L1 MemStore) can be successfully
	// committed at a time to maintain consistency.
	MemStoreManager interface {
		// Branch creates a new top-level MemStore (L1) based on the latest committed
		// state managed by the MemStoreManager.
		//
		// Each L1 MemStore operates on a copy-on-write (CoW) snapshot of the
		// manager's state at the time of branching, providing isolation. Multiple L1
		// branches can be created and used concurrently for reads and writes within
		// their respective instances.
		//
		// However, committing changes back requires synchronization. Only the *first*
		// L1 MemStore to call Commit() successfully updates the manager's pending state.
		// Subsequent attempts by other L1 branches (created from the *same* base state)
		// to Commit() will fail (panic) due to concurrent modification detection.
		Branch() MemStore

		// GetSnapshotBranch retrieves a MemStore representing the state at a specific
		// past height, if available in the snapshot pool.
		// The returned MemStore provides a read-only view of that historical state
		// and typically cannot be committed (often wrapped in an UncommittableMemStore).
		// Returns the MemStore and true if found, otherwise nil and false.
		GetSnapshotBranch(height int64) (MemStore, bool)

		SetSnapshotPoolLimit(limit int64)

		// Commit finalizes the pending changes (typically written by a single preceding
		// successful L1 MemStore.Commit() call) into the manager's main state (root)
		// at the specified height.
		//
		// This operation atomically updates the root pointer to reflect the new state
		// and potentially creates a snapshot of this state at the given height,
		// making it available via GetSnapshotBranch.
		// Panics if height is negative.
		Commit(height int64)
	}

	MemStoreReader interface {
		// Get retrieves a value for the given key from the current MemStore's btree.
		// Returns nil if the key does not exist.
		Get(key []byte) any

		// Iterator returns an iterator over the key-value pairs in the MemStore
		// within the specified range [start, end).
		//
		// IMPORTANT: The iterator operates on an immutable snapshot of the MemStore's
		// state taken at the moment Iterator() is called. Subsequent modifications
		// to the MemStore using Set() or Delete() will *not* be reflected in the
		// existing iterator.
		//
		// The iterator includes items with key >= start and key < end.
		// If start is nil, iteration starts from the first key.
		// If end is nil, iteration continues to the last key.
		// Panics if an error occurs during initialization (e.g., invalid range).
		Iterator(start, end []byte) MemStoreIterator
		// ReverseIterator returns an iterator over the key-value pairs in the MemStore
		// within the specified range [start, end), in reverse order.
		//
		// IMPORTANT: Like Iterator(), this operates on an immutable snapshot of the
		// MemStore's state at the time ReverseIterator() is called.
		//
		// The iterator includes items with key >= start and key < end.
		// If start is nil, iteration starts from the last key.
		// If end is nil, iteration continues to the first key.
		// Panics if an error occurs during initialization (e.g., invalid range).
		ReverseIterator(start, end []byte) MemStoreIterator
	}

	MemStoreWriter interface {
		// Set adds or updates a key-value pair in the current MemStore's btree.
		// If the key already exists, its value is overwritten.
		//
		// Changes are made to the Copy-on-Write btree of the current MemStore.
		Set(key []byte, value any)

		// Delete removes a key from the current MemStore's btree.
		//
		// Changes are made to the Copy-on-Write btree of the current MemStore.
		Delete(key []byte)

		// Commit applies the changes made within this MemStore.
		//
		// Behavior depends on whether it's a nested or top-level (L1) branch:
		// - Nested Branches: Updates the parent MemStore's internal state pointer
		//   to point to this branch's modified B-Tree. This operation itself is
		//   *not* guaranteed to be atomic or thread-safe with respect to other
		//   concurrent operations on the parent. Higher-level synchronization
		//   is needed if the parent is accessed concurrently.
		// - Top-Level (L1) Branches: Prepares the changes to be finalized by the
		//   MemStoreManager. It checks if the underlying base state managed by the
		//   MemStoreManager has changed since this L1 branch was created.
		//   If the base state *has* changed (indicating another L1 branch from the
		//   same base has already committed), this Commit() call will panic to
		//   prevent lost updates (concurrent modification detected). If successful,
		//   it updates the manager's *pending* state, which is then finalized by
		//   calling MemStoreManager.Commit().
		Commit()
	}

	// MemStore defines operations that can be performed on a memory store.
	//
	// The implementation of MemStore is "not thread-safe".
	// Therefore, when concurrent access is required, users should protect it
	// using rwlock at the application level.
	MemStore interface {
		// Branch creates a nested MemStore (e.g., L2, L3) based on the current state
		// of this MemStore (the parent).
		// It uses a copy-on-write (CoW) snapshot of the parent's current B-Tree,
		// allowing isolated modifications within the nested branch.
		// Changes in the nested branch are not visible to the parent until Commit()
		// is called on the nested branch.
		Branch() MemStore

		MemStoreReader
		MemStoreWriter
	}

	// MemStoreIterator defines an interface for traversing key-value pairs in order.
	// Callers must call Close when done to release any allocated resources.
	MemStoreIterator interface {
		// Domain returns the start and end keys defining the range of this iterator.
		// The returned values match what was passed to Iterator() or ReverseIterator().
		Domain() (start, end []byte)

		// Key returns the current key.
		// Panics if the iterator is not valid.
		Key() []byte

		// Value returns the current value.
		// Panics if the iterator is not valid.
		Value() any

		// Valid returns whether the iterator is positioned at a valid item.
		// Once false, Valid() will never return true again.
		Valid() bool

		// Next moves the iterator to the next item.
		// If Valid() returns false after this call, the iteration is complete.
		Next()

		// Close releases any resources associated with the iterator.
		// It must be called when done using the iterator.
		Close() error
	}

	// SnapshotPool defines an interface for storing and retrieving historical
	// versions (snapshots) of MemStoreManager states, typically keyed by block height.
	// Implementations may enforce limits on the number of snapshots stored.
	SnapshotPool interface {
		// Get retrieves a MemStoreManager representing the state at a specific height.
		// Returns the manager and true if found, otherwise nil and false.
		Get(height int64) (MemStoreManager, bool)

		// Set stores a MemStoreManager representing the state at a specific height.
		// If a manager for this height already exists, it might be overwritten.
		Set(height int64, store MemStoreManager)

		// Limit sets the maximum number of snapshots the pool should retain.
		// Implementations may use LRU or other strategies for pruning when the
		// limit is exceeded.
		Limit(length int64)
	}

	// TypedMemStore defines operations that can be performed on a memory store with generic type support.
	// It provides type-safe access to values of type T, automatically handling type conversion.
	// The interface is designed to allow for isolation of key spaces while maintaining type safety.
	TypedMemStore[T any] interface {
		// Branch creates a nested TypedMemStore with the same type parameter.
		// It creates an independent workspace that can be committed back to the parent.
		Branch() TypedMemStore[T]

		// Get retrieves a value for the given key and returns it as type T.
		// If the key does not exist, returns the zero value of T.
		Get(key []byte) T

		// Iterator returns an iterator over the key-value pairs within the specified range.
		//
		// The iterator will include items with key >= start and key < end.
		// If start is nil, it returns all items from the beginning.
		// If end is nil, it returns all items until the end.
		Iterator(start, end []byte) TypedMemStoreIterator[T]

		// ReverseIterator returns an iterator over the key-value pairs in reverse order.
		//
		// The iterator will include items with key >= start and key < end, in reverse order.
		// If start is nil, it returns all items from the beginning.
		// If end is nil, it returns all items until the end.
		ReverseIterator(start, end []byte) TypedMemStoreIterator[T]

		// Set adds or updates a key-value pair of type T.
		Set(key []byte, value T)

		// Delete removes a key from the store.
		Delete(key []byte)

		// Commit applies the changes in the current store to its parent.
		Commit()
	}

	// TypedMemStoreIterator defines an interface for traversing key-value pairs in a typed store.
	// It provides type-safe access to values.
	// Callers must call Close when done to release any allocated resources.
	TypedMemStoreIterator[T any] interface {
		// Domain returns the start and end keys defining the range of this iterator.
		// The returned values match what was passed to Iterator() or ReverseIterator().
		Domain() ([]byte, []byte)

		// Valid returns whether the iterator is positioned at a valid item.
		// Once false, Valid() will never return true again.
		Valid() bool

		// Next moves the iterator to the next item.
		// If Valid() returns false after this call, the iteration is complete.
		Next()

		// Key returns the current key.
		// Panics if the iterator is not valid.
		Key() []byte

		// Value returns the current value as type T.
		// If the iterator is not valid, returns the zero value of T.
		Value() T

		// Close releases any resources associated with the iterator.
		// It must be called when done using the iterator.
		Close() error

		// Error returns an error if the iterator is invalid.
		// Returns nil if the iterator is valid.
		Error() error
	}
)
