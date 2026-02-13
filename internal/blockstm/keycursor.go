package blockstm

import (
	"bytes"

	"github.com/tidwall/btree"

	"github.com/cosmos/cosmos-sdk/internal/blockstm/tree"
)

type keyCursor[V any] interface {
	Valid() bool
	Key() Key
	// Tree returns the per-key version tree, when available.
	Tree() *tree.SmallBTree[secondaryDataItem[V]]
	Next()
	// Seek positions the cursor on the given key (if present).
	Seek(Key) bool
	Close()
}

type noopKeyCursor[V any] struct{}

func (noopKeyCursor[V]) Valid() bool { return false }
func (noopKeyCursor[V]) Key() Key    { return nil }
func (noopKeyCursor[V]) Tree() *tree.SmallBTree[secondaryDataItem[V]] {
	return nil
}
func (noopKeyCursor[V]) Next()         {}
func (noopKeyCursor[V]) Seek(Key) bool { return false }
func (noopKeyCursor[V]) Close()        {}

// btreeKeyCursor iterates the ordered key set using a tidwall/btree iterator.
//
// Depending on the underlying tree options, the iterator may hold a read lock
// until Close(). Callers should not block while the cursor is open.
type btreeKeyCursor[V any] struct {
	iter btree.IterG[mvIndexKeyEntry[V]]

	start     []byte
	end       []byte
	ascending bool
	valid     bool
}

func newBTreeKeyCursor[V any](keys interface {
	Iter() btree.IterG[mvIndexKeyEntry[V]]
}, start, end []byte, ascending bool,
) *btreeKeyCursor[V] {
	if keys == nil {
		return &btreeKeyCursor[V]{start: start, end: end, ascending: ascending, valid: false}
	}

	it := keys.Iter()
	c := &btreeKeyCursor[V]{iter: it, start: start, end: end, ascending: ascending}
	c.valid = c.seekInitial()
	return c
}

func (c *btreeKeyCursor[V]) seekInitial() bool {
	var ok bool
	if c.ascending {
		if c.start != nil {
			search := mvIndexKeyEntry[V]{Key: Key(c.start)}
			ok = c.iter.Seek(search)
		} else {
			ok = c.iter.First()
		}
	} else {
		if c.end != nil {
			search := mvIndexKeyEntry[V]{Key: Key(c.end)}
			ok = c.iter.Seek(search)
			if !ok {
				ok = c.iter.Last()
			} else {
				// end is exclusive
				ok = c.iter.Prev()
			}
		} else {
			ok = c.iter.Last()
		}
	}
	if !ok {
		return false
	}
	return c.keyInRange(c.Key())
}

func (c *btreeKeyCursor[V]) Close() {
	c.iter.Release()
}

func (c *btreeKeyCursor[V]) Valid() bool {
	return c.valid
}

func (c *btreeKeyCursor[V]) Key() Key {
	return c.iter.Item().Key
}

func (c *btreeKeyCursor[V]) Tree() *tree.SmallBTree[secondaryDataItem[V]] {
	return c.iter.Item().Tree
}

func (c *btreeKeyCursor[V]) Next() {
	if !c.valid {
		return
	}
	if c.ascending {
		c.valid = c.iter.Next()
	} else {
		c.valid = c.iter.Prev()
	}
	if c.valid {
		c.valid = c.keyInRange(c.Key())
	}
}

func (c *btreeKeyCursor[V]) Seek(key Key) bool {
	search := mvIndexKeyEntry[V]{Key: key}
	c.valid = c.iter.Seek(search)
	if c.valid {
		c.valid = c.keyInRange(c.Key())
	}
	return c.valid
}

func (c *btreeKeyCursor[V]) keyInRange(key []byte) bool {
	if c.ascending {
		return c.end == nil || bytes.Compare(key, c.end) < 0
	}
	return c.start == nil || bytes.Compare(key, c.start) >= 0
}
