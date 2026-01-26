package tree

import (
	"github.com/tidwall/btree"
)

// BTree wraps a btree.BTreeG using tidwall/btree's internal locking.
//
// Avoids copy-on-write (Copy+Freeze+CAS) allocations on updates.
type BTree[T any] struct {
	t *btree.BTreeG[T]
}

// NewBTree returns a new BTree.
func NewBTree[T any](less func(a, b T) bool, degree int) *BTree[T] {
	return &BTree[T]{
		t: btree.NewBTreeGOptions(less, btree.Options{
			NoLocks: false,
			Degree:  degree,
		}),
	}
}

func (bt *BTree[T]) Get(item T) (result T, ok bool) {
	return bt.t.Get(item)
}

func (bt *BTree[T]) GetOrDefault(item T, fillDefaults func(*T)) T {
	result, ok := bt.t.Get(item)
	if ok {
		return result
	}
	fillDefaults(&item)
	bt.t.Set(item)
	return item
}

func (bt *BTree[T]) Set(item T) (prev T, ok bool) {
	return bt.t.Set(item)
}

func (bt *BTree[T]) Delete(item T) (prev T, ok bool) {
	return bt.t.Delete(item)
}

func (bt *BTree[T]) Scan(iter func(item T) bool) {
	bt.t.Scan(iter)
}

func (bt *BTree[T]) Max() (T, bool) {
	return bt.t.Max()
}

func (bt *BTree[T]) Iter() btree.IterG[T] {
	return bt.t.Iter()
}

// ReverseSeek returns the first item that is less than or equal to the pivot
func (bt *BTree[T]) ReverseSeek(pivot T) (result T, ok bool) {
	bt.t.Descend(pivot, func(item T) bool {
		result = item
		ok = true
		return false
	})
	return result, ok
}
