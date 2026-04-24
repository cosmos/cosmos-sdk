package tree

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/cosmos/btree"
	"go.opentelemetry.io/otel/metric"
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

func (bt *BTree[T]) Get(ctx context.Context, item T) (result T, ok bool) {
	defer measureSince(ctx, func() metric.Int64Histogram { return treeInst.Get }, time.Now())
	return bt.Load().Get(item)
}

func (bt *BTree[T]) GetOrDefault(ctx context.Context, item T, fillDefaults func(*T)) T {
	defer measureSince(ctx, func() metric.Int64Histogram { return treeInst.GetOrDefault }, time.Now())
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

func (bt *BTree[T]) Set(ctx context.Context, item T) (prev T, ok bool) {
	defer measureSince(ctx, func() metric.Int64Histogram { return treeInst.Set }, time.Now())
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

func (bt *BTree[T]) Delete(ctx context.Context, item T) (prev T, ok bool) {
	defer measureSince(ctx, func() metric.Int64Histogram { return treeInst.Delete }, time.Now())
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

func (bt *BTree[T]) Scan(ctx context.Context, iter func(item T) bool) {
	defer measureSince(ctx, func() metric.Int64Histogram { return treeInst.Scan }, time.Now())
	bt.Load().Scan(iter)
}

func (bt *BTree[T]) Max() (T, bool) {
	return bt.Load().Max()
}

func (bt *BTree[T]) Iter() btree.IterG[T] {
	return bt.Load().Iter()
}

// ReverseSeek returns the first item that is less than or equal to the pivot
func (bt *BTree[T]) ReverseSeek(ctx context.Context, pivot T) (result T, ok bool) {
	defer measureSince(ctx, func() metric.Int64Histogram { return treeInst.ReverseSeek }, time.Now())
	bt.Load().Descend(pivot, func(item T) bool {
		result = item
		ok = true
		return false
	})
	return result, ok
}
