package tree

import (
	"sync/atomic"

	"github.com/tidwall/btree"
)

// COWBTree is a copy-on-write wrapper around tidwall/btree.
//
// Readers iterate over an immutable snapshot (no locks). Writers copy, mutate,
// Freeze(), then atomically publish the new snapshot.
//
// This avoids holding locks during iteration (useful for key-set iteration).
//
// Note: The underlying tree is created with NoLocks=true and ReadOnly=true;
// safety relies on only publishing frozen snapshots.
type COWBTree[T any] struct {
	atomic.Pointer[btree.BTreeG[T]]
}

// NewCOWBTree returns a new copy-on-write BTree.
func NewCOWBTree[T any](less func(a, b T) bool, degree int) *COWBTree[T] {
	tree := btree.NewBTreeGOptions(less, btree.Options{
		NoLocks:  true,
		ReadOnly: true,
		Degree:   degree,
	})
	bt := &COWBTree[T]{}
	bt.Store(tree)
	return bt
}

func (bt *COWBTree[T]) Get(item T) (result T, ok bool) {
	return bt.Load().Get(item)
}

func (bt *COWBTree[T]) GetOrDefault(item T, fillDefaults func(*T)) T {
	for {
		t := bt.Load()
		result, ok := t.Get(item)
		if ok {
			return result
		}
		fillDefaults(&item)
		c := t.Copy()
		c.Set(item)
		c.Freeze()
		if bt.CompareAndSwap(t, c) {
			return item
		}
	}
}

func (bt *COWBTree[T]) Set(item T) (prev T, ok bool) {
	for {
		t := bt.Load()
		c := t.Copy()
		prev, ok = c.Set(item)
		c.Freeze()
		if bt.CompareAndSwap(t, c) {
			return prev, ok
		}
	}
}

func (bt *COWBTree[T]) Delete(item T) (prev T, ok bool) {
	for {
		t := bt.Load()
		c := t.Copy()
		prev, ok = c.Delete(item)
		c.Freeze()
		if bt.CompareAndSwap(t, c) {
			return prev, ok
		}
	}
}

func (bt *COWBTree[T]) Scan(iter func(item T) bool) {
	bt.Load().Scan(iter)
}

func (bt *COWBTree[T]) Max() (T, bool) {
	return bt.Load().Max()
}

func (bt *COWBTree[T]) Iter() btree.IterG[T] {
	return bt.Load().Iter()
}

// ReverseSeek returns the first item that is less than or equal to the pivot.
func (bt *COWBTree[T]) ReverseSeek(pivot T) (result T, ok bool) {
	bt.Load().Descend(pivot, func(item T) bool {
		result = item
		ok = true
		return false
	})
	return result, ok
}

func (bt *COWBTree[T]) Ascend(pivot T, iter func(item T) bool) {
	bt.Load().Ascend(pivot, iter)
}

func (bt *COWBTree[T]) Descend(pivot T, iter func(item T) bool) {
	bt.Load().Descend(pivot, iter)
}
