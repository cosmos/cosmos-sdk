package blockstm

import (
	"bytes"
	"fmt"
	"sync/atomic"
)

type ErrReadError struct {
	BlockingTxn TxnIndex
}

func (e ErrReadError) Error() string {
	return fmt.Sprintf("read error: blocked by txn %d", e.BlockingTxn)
}

// StoreMin implements a compare-and-swap operation that stores the minimum of the current value and the given value.
func StoreMin(a *atomic.Uint64, b uint64) {
	for {
		old := a.Load()
		if old <= b {
			return
		}
		if a.CompareAndSwap(old, b) {
			return
		}
	}
}

// DecrAtomic decreases the atomic value by 1
func DecrAtomic(a *atomic.Uint64) {
	a.Add(^uint64(0))
}

// IncrAtomic increases the atomic value by 1
func IncrAtomic(a *atomic.Uint64) {
	a.Add(1)
}

// FetchIncr increases the atomic value by 1 and returns the old value
func FetchIncr(a *atomic.Uint64) uint64 {
	return a.Add(1) - 1
}

// DiffOrderedList compares two ordered lists
// callback arguments: (value, is_new)
func DiffOrderedList(old, new []Key, callback func(Key, bool) bool) {
	i, j := 0, 0
	for i < len(old) && j < len(new) {
		switch bytes.Compare(old[i], new[j]) {
		case -1:
			if !callback(old[i], false) {
				return
			}
			i++
		case 1:
			if !callback(new[j], true) {
				return
			}
			j++
		default:
			i++
			j++
		}
	}
	for ; i < len(old); i++ {
		if !callback(old[i], false) {
			return
		}
	}
	for ; j < len(new); j++ {
		if !callback(new[j], true) {
			return
		}
	}
}

// DiffMemDB compares two tree
// callback arguments: (value, is_new)
func DiffMemDB[V any](old, new *GMemDB[V], callback func(Key, bool) bool) {
	oldIter := old.Iterator(nil, nil)
	defer oldIter.Close()

	newIter := new.Iterator(nil, nil)
	defer newIter.Close()

	for oldIter.Valid() && newIter.Valid() {
		switch bytes.Compare(oldIter.Key(), newIter.Key()) {
		case -1:
			if !callback(oldIter.Key(), false) {
				return
			}
			oldIter.Next()
		case 1:
			if !callback(newIter.Key(), true) {
				return
			}
			newIter.Next()
		default:
			oldIter.Next()
			newIter.Next()
		}
	}

	for ; oldIter.Valid(); oldIter.Next() {
		if !callback(oldIter.Key(), false) {
			return
		}
	}

	for ; newIter.Valid(); newIter.Next() {
		if !callback(newIter.Key(), true) {
			return
		}
	}
}

// BytesBeyond returns if a is beyond b in specified iteration order
func BytesBeyond(a, b []byte, ascending bool) bool {
	if ascending {
		return bytes.Compare(a, b) > 0
	}
	return bytes.Compare(a, b) < 0
}
