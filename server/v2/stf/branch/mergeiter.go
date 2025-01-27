package branch

import (
	"bytes"
	"errors"

	corestore "cosmossdk.io/core/store"
)

var errInvalidIterator = errors.New("invalid iterator")

// mergedIterator merges a parent Iterator and a child Iterator.
// The child iterator may contain items that shadow or override items in the parent iterator.
// If the child iterator has the same key as the parent, the child's value takes precedence.
// Deleted items in the child (indicated by nil values) are skipped.
type mergedIterator[Parent, Child corestore.Iterator] struct {
	parent    Parent // Iterator for the parent store
	child     Child  // Iterator for the child store
	ascending bool   // Direction of iteration
	valid     bool   // Indicates if the iterator is in a valid state
	currKey   []byte // Current key pointed by the iterator
	currValue []byte // Current value corresponding to currKey
	err       error  // Error encountered during iteration
}

// Ensure mergedIterator implements the corestore.Iterator interface.
var _ corestore.Iterator = (*mergedIterator[corestore.Iterator, corestore.Iterator])(nil)

// mergeIterators creates a new merged iterator from parent and child iterators.
// The 'ascending' parameter determines the direction of iteration.
func mergeIterators[Parent, Child corestore.Iterator](parent Parent, child Child, ascending bool) *mergedIterator[Parent, Child] {
	iter := &mergedIterator[Parent, Child]{
		parent:    parent,
		child:     child,
		ascending: ascending,
	}
	iter.advance() // Initialize the iterator by advancing to the first valid item
	return iter
}

// Domain returns the start and end range of the iterator.
// It delegates to the parent iterator as both iterators share the same domain.
func (i *mergedIterator[Parent, Child]) Domain() (start, end []byte) {
	return i.parent.Domain()
}

// Valid checks if the iterator is in a valid state.
// It returns true if the iterator has not reached the end.
func (i *mergedIterator[Parent, Child]) Valid() bool {
	return i.valid
}

// Next advances the iterator to the next valid item.
// It skips over deleted items (with nil values) and updates the current key and value.
func (i *mergedIterator[Parent, Child]) Next() {
	if !i.valid {
		i.err = errInvalidIterator
		return
	}
	i.advance()
}

// Key returns the current key pointed by the iterator.
// If the iterator is invalid, it returns nil.
func (i *mergedIterator[Parent, Child]) Key() []byte {
	if !i.valid {
		panic("called key on invalid iterator")
	}
	return i.currKey
}

// Value returns the current value corresponding to the current key.
// If the iterator is invalid, it returns nil.
func (i *mergedIterator[Parent, Child]) Value() []byte {
	if !i.valid {
		panic("called value on invalid iterator")
	}
	return i.currValue
}

// Close closes both the parent and child iterators.
// It returns any error encountered during the closing of the iterators.
func (i *mergedIterator[Parent, Child]) Close() (err error) {
	err = errors.Join(err, i.parent.Close())
	err = errors.Join(err, i.child.Close())
	i.valid = false
	return err
}

// Error returns any error that occurred during iteration.
// If the iterator is valid, it returns nil.
func (i *mergedIterator[Parent, Child]) Error() error {
	return i.err
}

// advance moves the iterator to the next valid (non-deleted) item.
// It handles merging logic between the parent and child iterators.
func (i *mergedIterator[Parent, Child]) advance() {
	for {
		// Check if both iterators have reached the end
		if !i.parent.Valid() && !i.child.Valid() {
			i.valid = false
			return
		}

		var key, value []byte

		// If parent iterator is exhausted, use the child iterator
		if !i.parent.Valid() {
			key = i.child.Key()
			value = i.child.Value()
			i.child.Next()
		} else if !i.child.Valid() {
			// If child iterator is exhausted, use the parent iterator
			key = i.parent.Key()
			value = i.parent.Value()
			i.parent.Next()
		} else {
			// Both iterators are valid; compare keys
			keyP, keyC := i.parent.Key(), i.child.Key()
			switch cmp := i.compare(keyP, keyC); {
			case cmp < 0:
				// Parent key is less than child key
				key = keyP
				value = i.parent.Value()
				i.parent.Next()
			case cmp == 0:
				// Keys are equal; child overrides parent
				key = keyC
				value = i.child.Value()
				i.parent.Next()
				i.child.Next()
			case cmp > 0:
				// Child key is less than parent key
				key = keyC
				value = i.child.Value()
				i.child.Next()
			}
		}

		// Skip deleted items (value is nil)
		if value == nil {
			continue
		}

		// Update the current key and value, and mark iterator as valid
		i.currKey = key
		i.currValue = value
		i.valid = true
		return
	}
}

// compare compares two byte slices a and b.
// It returns an integer comparing a and b:
//   - Negative if a < b
//   - Zero if a == b
//   - Positive if a > b
//
// The comparison respects the iterator's direction (ascending or descending).
func (i *mergedIterator[Parent, Child]) compare(a, b []byte) int {
	if i.ascending {
		return bytes.Compare(a, b)
	}
	return bytes.Compare(b, a)
}
