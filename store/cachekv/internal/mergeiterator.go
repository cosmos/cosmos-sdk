package internal

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
//
// TODO: Optimize by memoizing.
type cacheMergeIterator[V any] struct {
	parent    types.GIterator[V]
	cache     types.GIterator[V]
	ascending bool

	valid bool

	isZero func(V) bool
}

var _ types.Iterator = (*cacheMergeIterator[[]byte])(nil)

func NewCacheMergeIterator[V any](parent, cache types.GIterator[V], ascending bool, isZero func(V) bool) types.GIterator[V] {
	iter := &cacheMergeIterator[V]{
		parent:    parent,
		cache:     cache,
		ascending: ascending,
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

	switch {
	case !iter.parent.Valid():
		// If parent is invalid, get the next cache item.
		iter.cache.Next()
	case !iter.cache.Valid():
		// If cache is invalid, get the next parent item.
		iter.parent.Next()
	default:
		// Both are valid.  Compare keys.
		keyP, keyC := iter.parent.Key(), iter.cache.Key()
		switch iter.compare(keyP, keyC) {
		case -1: // parent < cache
			iter.parent.Next()
		case 0: // parent == cache
			iter.parent.Next()
			iter.cache.Next()
		case 1: // parent > cache
			iter.cache.Next()
		}
	}
	iter.valid = iter.skipUntilExistsOrInvalid()
}

// Key implements Iterator
func (iter *cacheMergeIterator[V]) Key() []byte {
	iter.assertValid()

	// If parent is invalid, get the cache key.
	if !iter.parent.Valid() {
		return iter.cache.Key()
	}

	// If cache is invalid, get the parent key.
	if !iter.cache.Valid() {
		return iter.parent.Key()
	}

	// Both are valid.  Compare keys.
	keyP, keyC := iter.parent.Key(), iter.cache.Key()

	cmp := iter.compare(keyP, keyC)
	switch cmp {
	case -1: // parent < cache
		return keyP
	case 0: // parent == cache
		return keyP
	case 1: // parent > cache
		return keyC
	default:
		panic("invalid compare result")
	}
}

// Value implements Iterator
func (iter *cacheMergeIterator[V]) Value() V {
	iter.assertValid()

	// If parent is invalid, get the cache value.
	if !iter.parent.Valid() {
		return iter.cache.Value()
	}

	// If cache is invalid, get the parent value.
	if !iter.cache.Valid() {
		return iter.parent.Value()
	}

	// Both are valid.  Compare keys.
	keyP, keyC := iter.parent.Key(), iter.cache.Key()

	cmp := iter.compare(keyP, keyC)
	switch cmp {
	case -1: // parent < cache
		return iter.parent.Value()
	case 0: // parent == cache
		return iter.cache.Value()
	case 1: // parent > cache
		return iter.cache.Value()
	default:
		panic("invalid comparison result")
	}
}

// Close implements Iterator
func (iter *cacheMergeIterator[V]) Close() error {
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
			return iter.cache.Valid()
		}
		// Parent is valid.

		if !iter.cache.Valid() {
			return true
		}
		// Parent is valid, cache is valid.

		// Compare parent and cache.
		keyP := iter.parent.Key()
		keyC := iter.cache.Key()

		switch iter.compare(keyP, keyC) {
		case -1: // parent < cache.
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

			return true // cache exists.
		case 1: // cache < parent
			// Skip over if cache item is a delete.
			valueC := iter.cache.Value()
			if iter.isZero(valueC) {
				iter.skipCacheDeletes(keyP)
				continue
			}
			// Cache is not a delete.

			return true // cache exists.
		}
	}
}
