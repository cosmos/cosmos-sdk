package memstore

import (
	"sync/atomic"

	"cosmossdk.io/store/memstore/internal"
	"cosmossdk.io/store/types"
)

var (
	_ types.MemStoreManager = &memStoreManager{}
	_ types.MemStore        = &memStore{}
)

type (
	btree = internal.BTree

	memStoreManager struct {
		// root is an atomic pointer to the current root of the memStoreManager.
		// When a branch is committed, it creates a new root node and atomically
		// swaps it with the existing one.
		root *atomic.Pointer[btree]
		// The current B-tree is stored in the root atomic.Pointer when memStoreManager.Commit() occurs.
		// The reason for creating it temporarily in this way is that it should not be read
		// in the FinalizeBlock state before BaseApp.Commit happens.
		//
		// `current` is implemented with the assumption that it is accessed only by a single writer.
		current *btree
		// base ensures that only one branch (L1) can be committed.
		//
		// It is set to nil for Trees retrieved from the snapshotPool.
		base *atomic.Pointer[btree]

		snapshotPool SnapshotPool
	}

	// `memStore` implements a copy-on-write memStore operation pattern.
	// Nested branches follow this structure as well, updating their parent's current on `Commit()`.
	memStore struct {
		// parent is nil for top-level branches, non-nil for nested branches
		parent *memStore

		// current holds the current working copy of the btree for this top-level branch.
		current *btree

		// base points to memStoreManager.base when the branch is created.
		//
		// If memstoreManager.base differs from base at commit time, the commit fails and a panic occurs.
		base *btree

		// manager is a reference to the parent memStoreManager, used by top-level branches to update the manager during commit.
		manager *memStoreManager
	}

	UncommittableMemStore struct {
		types.MemStore
	}
)

func (u *UncommittableMemStore) Commit() {
	panic("uncommittable MemStore cannot be committed")
}

// NewMemStoreManager creates a new empty memStoreManager.
func NewMemStoreManager() *memStoreManager {
	tree := internal.NewBTree()

	root := &atomic.Pointer[btree]{}
	root.Store(tree)

	base := &atomic.Pointer[btree]{}
	base.Store(tree)

	return &memStoreManager{
		root:    root,
		current: tree.Copy(),
		base:    base,

		snapshotPool: newSnapshotPool(),
	}
}

func (t *memStoreManager) SetSnapshotPoolLimit(limit int64) {
	t.snapshotPool.Limit(limit)
}

// Committing a branch created here is unsafe.
func (t *memStoreManager) GetSnapshotBranch(height int64) (types.MemStore, bool) {
	reader, ok := t.snapshotPool.Get(height)
	if !ok {
		return nil, false
	}

	branch := reader.Branch()
	// The snapshot branch is a branch type used in CacheMultiStoreWithVersion(..).
	// Since it handles data for queries at past heights, it is immutable.
	// Therefore, Commit() cannot be performed.
	return &UncommittableMemStore{branch}, true
}

// Branch creates a top-level branch.
// It creates a copy-on-write snapshot of the tree's root btree as its working copy.
func (t *memStoreManager) Branch() types.MemStore {
	root := t.root.Load()
	// Create a copy-on-write snapshot for the current
	current := root.Copy()

	var base *btree
	if t.base != nil {
		base = t.base.Load()
	}

	return &memStore{
		// This is a top-level branch, so parent is nil
		parent: nil,

		current: current,
		base:    base,
		manager: t,
	}
}

// Commit finalizes the current state at the specified height by:
// 1. Ensuring the base hasn't changed (preventing concurrent commits)
// 2. Atomically updating the root btree with current changes
// 3. Creating a snapshot at the given height for future queries
// The height must be non-negative.
func (t *memStoreManager) Commit(height int64) {
	if height < 0 {
		// NOTE: When height is 0, it occurs when calling LoadLatestVersion() on an empty app.
		// Since InitChain, which registers the genesis, is called afterward,
		// MemStore.Commit(0) can happen.
		//
		// For this reason, it should be allowed.
		panic("height cannot be a negative value.")
	}

	// Since `current` is used only in a single thread, direct access is safe.
	current := t.current

	if current == nil {
		panic("`MemStore.Commit(..)` should not be called on an memstore retrieved from the snapshot pool.")
	}

	copiedTree := current.Copy()
	if !t.root.CompareAndSwap(t.base.Load(), current) {
		panic("commit failed: concurrent modification detected")
	}
	t.current = copiedTree
	t.base.Store(current)

	snapshotTree := copiedTree.Copy()

	root := &atomic.Pointer[btree]{}
	root.Store(snapshotTree)

	t.snapshotPool.Set(height, &memStoreManager{
		root:    root,
		current: nil, // Trees stored in the snapshot pool cannot be committed
		base:    nil,

		snapshotPool: nil,
	})
}

// Get retrieves a value for the given key from the current branch.
func (b *memStore) Get(key []byte) any {
	return b.current.Get(key)
}

// Iterator returns an iterator over the key-value pairs in the branch
// within the specified range.
//
// The iterator will include items with key >= start and key < end.
// If start is nil, it returns all items from the beginning.
// If end is nil, it returns all items until the end.
//
// If an error occurs during initialization, this method panics.
func (b *memStore) Iterator(start, end []byte) types.MemStoreIterator {
	// NOTE: If a snapshot is not created, the current BTree cannot be modified until the Iterator is closed.
	snapshot := b.current.Copy()

	iter, err := snapshot.Iterator(start, end)
	if err != nil {
		panic(err)
	}

	return iter
}

// ReverseIterator returns an iterator over the key-value pairs in the branch
// within the specified range, in reverse order (from end to start).
//
// The iterator will include items with key >= start and key < end.
// If start is nil, it returns all items from the beginning.
// If end is nil, it returns all items until the end.
//
// If an error occurs during initialization, this method panics.
func (b *memStore) ReverseIterator(start, end []byte) types.MemStoreIterator {
	// NOTE: If a snapshot is not created, the current BTree cannot be modified until the Iterator is closed.
	snapshot := b.current.Copy()

	iter, err := snapshot.ReverseIterator(start, end)
	if err != nil {
		panic(err)
	}

	return iter
}

// Set adds or updates a key-value pair in the current branch.
func (b *memStore) Set(key []byte, value any) {
	b.current.Set(key, value)
}

// Delete removes a key from the current branch.
func (b *memStore) Delete(key []byte) {
	b.current.Delete(key)
}

// Branch creates a nested branch on top of the current branch.
// It copies the current branch's btree to create an independent workspace.
func (b *memStore) Branch() types.MemStore {
	// Here, current refers to the current level's -1 level branch:
	//   -> If current is L3: points to L2's current
	//   -> If current is L2: points to L1's current
	//   -> If current is L1: points to memstoreManager.current
	//
	// newCurrent is a Copy()'d memstoreManager.
	newCurrent := b.current.Copy()

	return &memStore{
		parent: b,

		current: newCurrent,
	}
}

// Commit applies the changes in the branch:
// - For nested branches, it updates the parent branch's current pointer.
// - For top-level branches, it updates memStoreManager.current with the branch's current btree.
func (b *memStore) Commit() {
	if b.parent != nil {
		// nested branch: update parent's current pointer
		b.parent.current = b.current
		return
	}

	if b.current != nil {
		// top-level branch: swap *memstoreManager.current
		if b.manager.base.Load() != b.base {
			panic("commit failed: concurrent modification detected")
		}

		b.manager.current = b.current
		return
	}

	panic("unreachable code, parent is nil & current is nil")
}
