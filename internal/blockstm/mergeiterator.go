package blockstm

import (
	"bytes"
	"errors"

	"cosmossdk.io/store/types"
)

// cacheMergeIterator merges a parent Iterator and a cache Iterator.
// The cache iterator may return nil keys to signal that an item
// had been deleted (but not deleted in the parent).
// If the cache iterator has the same key as the parent, the
// cache shadows (overrides) the parent.
type cacheMergeIterator[V any] struct {
	parent  types.GIterator[V]
	cache   types.GIterator[V]
	onClose func(types.GIterator[V])
	isZero  func(V) bool

	ascending bool
	// Memoize whether the current item comes from parent or cache.
	useCache      bool
	advanceParent bool
	advanceCache  bool
	valid         bool
}

var _ types.Iterator = (*cacheMergeIterator[[]byte])(nil)

func NewCacheMergeIterator[V any](
	parent, cache types.GIterator[V],
	ascending bool, onClose func(types.GIterator[V]),
	isZero func(V) bool,
) types.GIterator[V] {
	iter := &cacheMergeIterator[V]{
		parent:    parent,
		cache:     cache,
		ascending: ascending,
		onClose:   onClose,
		isZero:    isZero,
	}

	iter.valid = iter.skipUntilExistsOrInvalid()
	return iter
}

// Domain implements Iterator.
// Returns parent domain because cache and parent domains are the same.
func (iter *cacheMergeIterator[V]) Domain() (start, end []byte) {
	return iter.parent.Domain()
}

// Valid implements Iterator.
func (iter *cacheMergeIterator[V]) Valid() bool {
	return iter.valid
}

// Next implements Iterator
func (iter *cacheMergeIterator[V]) Next() {
	iter.assertValid()

	// Advance based on the memoized selection from skipUntilExistsOrInvalid().
	if iter.advanceParent {
		iter.parent.Next()
	}
	if iter.advanceCache {
		iter.cache.Next()
	}
	iter.valid = iter.skipUntilExistsOrInvalid()
}

// Key implements Iterator
func (iter *cacheMergeIterator[V]) Key() []byte {
	iter.assertValid()
	if iter.useCache {
		return iter.cache.Key()
	}
	return iter.parent.Key()
}

// Value implements Iterator
func (iter *cacheMergeIterator[V]) Value() V {
	iter.assertValid()
	if iter.useCache {
		return iter.cache.Value()
	}
	return iter.parent.Value()
}

func (iter *cacheMergeIterator[V]) selectParent() {
	iter.useCache = false
	iter.advanceParent = true
	iter.advanceCache = false
}

func (iter *cacheMergeIterator[V]) selectCache() {
	iter.useCache = true
	iter.advanceParent = false
	iter.advanceCache = true
}

// selectEqual: same key in parent+cache, cache is not delete; cache wins and Next advances both.
func (iter *cacheMergeIterator[V]) selectEqual() {
	iter.useCache = true
	iter.advanceParent = true
	iter.advanceCache = true
}

// Close implements Iterator
func (iter *cacheMergeIterator[V]) Close() error {
	if iter.onClose != nil {
		iter.onClose(iter)
	}

	err1 := iter.cache.Close()
	if err := iter.parent.Close(); err != nil {
		return err
	}

	return err1
}

// Error returns an error if the cacheMergeIterator is invalid defined by the
// Valid method.
func (iter *cacheMergeIterator[V]) Error() error {
	if !iter.Valid() {
		return errors.New("invalid cacheMergeIterator")
	}

	return nil
}

// If not valid, panics.
// NOTE: May have side-effect of iterating over cache.
func (iter *cacheMergeIterator[V]) assertValid() {
	if err := iter.Error(); err != nil {
		panic(err)
	}
}

// Like bytes.Compare but opposite if not ascending.
func (iter *cacheMergeIterator[V]) compare(a, b []byte) int {
	if iter.ascending {
		return bytes.Compare(a, b)
	}

	return bytes.Compare(a, b) * -1
}

// Skip all delete-items from the cache w/ `key < until`.  After this function,
// current cache item is a non-delete-item, or `until <= key`.
// If the current cache item is not a delete item, does nothing.
// If `until` is nil, there is no limit, and cache may end up invalid.
// CONTRACT: cache is valid.
func (iter *cacheMergeIterator[V]) skipCacheDeletes(until []byte) {
	for iter.cache.Valid() &&
		iter.isZero(iter.cache.Value()) &&
		(until == nil || iter.compare(iter.cache.Key(), until) < 0) {
		iter.cache.Next()
	}
}

// Fast forwards cache (or parent+cache in case of deleted items) until current
// item exists, or until iterator becomes invalid.
// Returns whether the iterator is valid.
func (iter *cacheMergeIterator[V]) skipUntilExistsOrInvalid() bool {
	for {
		// If parent is invalid, fast-forward cache.
		if !iter.parent.Valid() {
			iter.skipCacheDeletes(nil)
			if !iter.cache.Valid() {
				return false
			}
			iter.selectCache()
			return true
		}
		// Parent is valid.

		if !iter.cache.Valid() {
			iter.selectParent()
			return true
		}
		// Parent is valid, cache is valid.

		// Compare parent and cache.
		keyP := iter.parent.Key()
		keyC := iter.cache.Key()

		switch iter.compare(keyP, keyC) {
		case -1: // parent < cache.
			iter.selectParent()
			return true

		case 0: // parent == cache.
			// Skip over if cache item is a delete.
			valueC := iter.cache.Value()
			if iter.isZero(valueC) {
				iter.parent.Next()
				iter.cache.Next()

				continue
			}
			// Cache is not a delete.
			iter.selectEqual()
			return true // cache exists.
		case 1: // cache < parent
			// Skip over if cache item is a delete.
			valueC := iter.cache.Value()
			if iter.isZero(valueC) {
				iter.skipCacheDeletes(keyP)
				continue
			}
			// Cache is not a delete.
			iter.selectCache()
			return true // cache exists.
		}
	}
}
