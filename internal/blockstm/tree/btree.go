package tree

import (
	"sync/atomic"

	"github.com/tidwall/btree"
)

// BTree wraps an atomic pointer to an unsafe btree.BTreeG
type BTree[T any] struct {
	atomic.Pointer[btree.BTreeG[T]]
}

// NewBTree returns a new BTree.
func NewBTree[T any](less func(a, b T) bool, degree int) *BTree[T] {
	tree := btree.NewBTreeGOptions(less, btree.Options{
		NoLocks:  true,
		ReadOnly: true,
		Degree:   degree,
	})
	t := &BTree[T]{}
	t.Store(tree)
	return t
}

func (bt *BTree[T]) Get(item T) (result T, ok bool) {
	return bt.Load().Get(item)
}

func (bt *BTree[T]) GetOrDefault(item T, fillDefaults func(*T)) T {
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

func (bt *BTree[T]) Set(item T) (prev T, ok bool) {
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

func (bt *BTree[T]) Delete(item T) (prev T, ok bool) {
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

func (bt *BTree[T]) Scan(iter func(item T) bool) {
	bt.Load().Scan(iter)
}

func (bt *BTree[T]) Max() (T, bool) {
	return bt.Load().Max()
}

func (bt *BTree[T]) Iter() btree.IterG[T] {
	return bt.Load().Iter()
}

// ReverseSeek returns the first item that is less than or equal to the pivot
func (bt *BTree[T]) ReverseSeek(pivot T) (result T, ok bool) {
	bt.Load().Descend(pivot, func(item T) bool {
		result = item
		ok = true
		return false
	})
	return result, ok
}
