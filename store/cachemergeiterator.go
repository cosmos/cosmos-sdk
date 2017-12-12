package store

import "bytes"

// cacheMergeIterator merges a parent Iterator and a cache Iterator.
// The cache iterator may return nil keys to signal that an item
// had been deleted (but not deleted in the parent).
// If the cache iterator has the same key as the parent, the
// cache shadows (overrides) the parent.
//
// TODO: Optimize by memoizing.
type cacheMergeIterator struct {
	parent    Iterator
	cache     Iterator
	ascending bool
}

var _ Iterator = (*cacheMergeIterator)(nil)

func newCacheMergeIterator(parent, cache Iterator, ascending bool) *cacheMergeIterator {
	iter := &cacheMergeIterator{
		parent:    parent,
		cache:     cache,
		ascending: ascending,
	}
	return iter
}

// Domain implements Iterator.
// If the domains are different, returns the union.
func (iter *cacheMergeIterator) Domain() (start, end []byte) {
	startP, endP := iter.parent.Domain()
	startC, endC := iter.cache.Domain()
	if iter.compare(startP, startC) < 0 {
		start = startP
	} else {
		start = startC
	}
	if iter.compare(endP, endC) < 0 {
		end = endC
	} else {
		end = endP
	}
	return start, end
}

// Valid implements Iterator.
func (iter *cacheMergeIterator) Valid() bool {

	// If parent is valid, this is valid.
	if iter.parent.Valid() {
		return true
	}

	// Otherwise depends on child.
	iter.skipCacheDeletes(nil)
	return iter.cache.Valid()
}

// Next implements Iterator
func (iter *cacheMergeIterator) Next() {
	iter.skipUntilExistsOrInvalid()
	iter.assertValid()

	// If parent is invalid, get the next cache item.
	if !iter.parent.Valid() {
		iter.cache.Next()
		return
	}

	// If cache is invalid, get the next parent item.
	if !iter.cache.Valid() {
		iter.parent.Next()
		return
	}

	// Both are valid.  Compare keys.
	keyP, keyC := iter.parent.Key(), iter.cache.Key()
	cmp := iter.compare(keyP, keyC)
	switch cmp {
	case -1: // parent < cache
		iter.parent.Next()
	case 0: // parent == cache
		iter.parent.Next()
		iter.cache.Next()
	case 1: // parent > cache
		iter.cache.Next()
	}
}

// Key implements Iterator
func (iter *cacheMergeIterator) Key() []byte {
	iter.skipUntilExistsOrInvalid()
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
func (iter *cacheMergeIterator) Value() []byte {
	iter.skipUntilExistsOrInvalid()
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

// Release implements Iterator
func (iter *cacheMergeIterator) Release() {
	iter.parent.Release()
	iter.cache.Release()
}

// Like bytes.Compare but opposite if not ascending.
func (iter *cacheMergeIterator) compare(a, b []byte) int {
	if iter.ascending {
		return bytes.Compare(a, b)
	} else {
		return bytes.Compare(a, b) * -1
	}
}

// Skip all delete-items from the cache w/ `key < until`.  After this function,
// current item is a non-delete-item, or `until <= key`.
// If the current item is not a delete item, does noting.
// If `until` is nil, there is no limit.
// CONTRACT: cache is valid.
func (iter *cacheMergeIterator) skipCacheDeletes(until []byte) {
	for (until == nil || iter.compare(iter.cache.Key(), until) < 0) &&
		iter.cache.Value() == nil {

		iter.cache.Next()
		if !iter.cache.Valid() {
			return
		}
	}
}

// Fast forwards cache (or parent+cache in case of deleted items) until current
// item exists, or until iterator becomes invalid.
func (iter *cacheMergeIterator) skipUntilExistsOrInvalid() {
	for {

		// Invalid.
		if !iter.Valid() {
			return
		}

		// Parent and Cache items exist.
		keyP, keyC := iter.parent.Key(), iter.cache.Key()
		cmp := iter.compare(keyP, keyC)
		switch cmp {

		// parent < cache
		case -1:

			// Parent exists.
			return

		// parent == cache
		case 0:

			// Skip over if cache item is a delete.
			valueC := iter.cache.Value()
			if valueC == nil {
				iter.parent.Next()
				iter.cache.Next()
				continue
			}
			// Child shadows parent.
			return

		// parent > cache
		case 1:

			// Skip over if cache item is a delete.
			valueC := iter.cache.Value()
			if valueC == nil {
				iter.skipCacheDeletes(keyP)
				continue
			}
			// Child exists.
			return
		}
	}
}

// If not valid, panics.
// NOTE: May have side-effect of iterating over cache.
func (iter *cacheMergeIterator) assertValid() {
	if !iter.Valid() {
		panic("iterator is invalid")
	}
}
