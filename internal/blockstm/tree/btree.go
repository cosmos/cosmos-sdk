package tree

import (
	"context"
	"sync/atomic"

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
	defer measureSince(ctx, func() metric.Int64Histogram { return treeInst.Get }, treeNow())
	return bt.Load().Get(item)
}

func (bt *BTree[T]) GetOrDefault(ctx context.Context, item T, fillDefaults func(*T)) T {
	defer measureSince(ctx, func() metric.Int64Histogram { return treeInst.GetOrDefault }, treeNow())
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

// BatchGetOrDefault atomically inserts defaults for missing items and returns the stored values.
func (bt *BTree[T]) BatchGetOrDefault(ctx context.Context, items []T, fillDefaults func(*T)) []T {
	defer measureSince(ctx, func() metric.Int64Histogram { return treeInst.BatchGetOrDefault }, treeNow())

	results := make([]T, len(items))
	defaults := make([]T, len(items))
	initialized := make([]bool, len(items))
	for {
		t := bt.Load()
		var updated *btree.BTreeG[T]
		for i, item := range items {
			lookup := t
			if updated != nil {
				lookup = updated
			}
			if result, ok := lookup.Get(item); ok {
				results[i] = result
				continue
			}

			if !initialized[i] {
				defaults[i] = item
				fillDefaults(&defaults[i])
				initialized[i] = true
			}
			if updated == nil {
				updated = t.Copy()
			}
			updated.Set(defaults[i])
			results[i] = defaults[i]
		}

		if updated == nil {
			return results
		}
		updated.Freeze()
		if bt.CompareAndSwap(t, updated) {
			return results
		}
	}
}

func (bt *BTree[T]) Set(ctx context.Context, item T) (prev T, ok bool) {
	defer measureSince(ctx, func() metric.Int64Histogram { return treeInst.Set }, treeNow())
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
	defer measureSince(ctx, func() metric.Int64Histogram { return treeInst.Delete }, treeNow())
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
	defer measureSince(ctx, func() metric.Int64Histogram { return treeInst.Scan }, treeNow())
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
	defer measureSince(ctx, func() metric.Int64Histogram { return treeInst.ReverseSeek }, treeNow())
	bt.Load().Descend(pivot, func(item T) bool {
		result = item
		ok = true
		return false
	})
	return result, ok
}
