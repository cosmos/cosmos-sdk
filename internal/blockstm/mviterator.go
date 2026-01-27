package blockstm

import (
	"bytes"

	storetypes "cosmossdk.io/store/types"

	tree2 "github.com/cosmos/cosmos-sdk/internal/blockstm/tree"
)

// MVIterator is an iterator for a multi-versioned store.
type MVIterator[V any] struct {
	opts IteratorOptions
	txn  TxnIndex

	mvData  *GMVData[V]
	keys    keyCursor[V]
	newKeys func() keyCursor[V]

	curKey  Key
	curTree *tree2.SmallBTree[secondaryDataItem[V]]

	// cache current found value and version
	value   V
	version TxnVersion

	// record the observed reads during iteration during execution
	reads []ReadDescriptor
	// blocking call to wait for dependent transaction to finish, `nil` in validation mode
	waitFn func(TxnIndex)
	// signal the validation to fail
	readEstimateValue bool

	err error
}

var _ storetypes.Iterator = (*MVIterator[[]byte])(nil)

func NewMVIterator[V any](
	opts IteratorOptions,
	txn TxnIndex,
	mvData *GMVData[V],
	keys keyCursor[V],
	newKeys func() keyCursor[V],
	waitFn func(TxnIndex),
) *MVIterator[V] {
	it := &MVIterator[V]{
		opts:    opts,
		txn:     txn,
		mvData:  mvData,
		keys:    keys,
		newKeys: newKeys,
		waitFn:  waitFn,
	}
	it.resolveValue()
	return it
}

// Executing returns if the iterator is running in execution mode.
func (it *MVIterator[V]) Executing() bool {
	return it.waitFn != nil
}

func (it *MVIterator[V]) Domain() (start, end []byte) {
	return it.opts.Start, it.opts.End
}

func (it *MVIterator[V]) Valid() bool {
	return !it.readEstimateValue && it.keys != nil && it.keys.Valid()
}

func (it *MVIterator[V]) Next() {
	if !it.Valid() {
		panic("iterator is invalid")
	}
	it.keys.Next()
	it.resolveValue()
}

func (it *MVIterator[V]) Key() (key []byte) {
	if !it.Valid() {
		panic("iterator is invalid")
	}
	return it.keys.Key()
}

func (it *MVIterator[V]) Value() V {
	if !it.Valid() {
		panic("iterator is invalid")
	}
	return it.value
}

func (it *MVIterator[V]) Error() error {
	return it.err
}

func (it *MVIterator[V]) Close() error {
	if it.keys != nil {
		it.keys.Close()
		it.keys = nil
	}
	it.curKey = nil
	it.curTree = nil
	it.reads = nil
	return nil
}

func (it *MVIterator[V]) treeForCurrentKey() *tree2.SmallBTree[secondaryDataItem[V]] {
	if it.keys == nil || !it.keys.Valid() {
		it.curKey = nil
		it.curTree = nil
		return nil
	}
	if tree := it.keys.Tree(); tree != nil {
		// Cursor already knows the per-key tree; avoid an extra mvIndex lookup.
		return tree
	}
	key := it.keys.Key()
	if it.curTree != nil && bytes.Equal(it.curKey, key) {
		return it.curTree
	}
	it.curKey = key
	it.curTree = it.mvData.getTree(key)
	return it.curTree
}

func (it *MVIterator[V]) Version() TxnVersion {
	return it.version
}

func (it *MVIterator[V]) Reads() []ReadDescriptor {
	return it.reads
}

func (it *MVIterator[V]) ReadEstimateValue() bool {
	return it.readEstimateValue
}

// resolveValue skips the non-exist values in the iterator based on the txn index, and caches the first existing one.
func (it *MVIterator[V]) resolveValue() {
	for it.keys != nil && it.keys.Valid() {
		tree := it.treeForCurrentKey()
		v, ok, needWait, waitOn := it.resolveValueInner(tree)
		if needWait {
			it.waitFn(waitOn)
			continue
		}
		if !ok {
			// signal the validation to fail
			it.readEstimateValue = true
			return
		}
		if v == nil {
			it.keys.Next()
			continue
		}

		it.value = v.Value
		it.version = v.Version()
		if it.Executing() {
			key := it.keys.Key()
			// Keys must be treated as immutable by callers, so we can record them
			// without allocating.
			it.reads = append(it.reads, ReadDescriptor{Key: key, Version: it.version})
		}
		return
	}
}

// resolveValueInner loop until we find a value that is not an estimate,
// wait for dependency if gets an ESTIMATE.
// returns:
// - (nil, true) if the value is not found
// - (nil, false) if the value is an estimate and we should fail the validation
// - (v, true) if the value is found
func (it *MVIterator[V]) resolveValueInner(tree *tree2.SmallBTree[secondaryDataItem[V]]) (*secondaryDataItem[V], bool, bool, TxnIndex) {
	if tree == nil {
		return nil, true, false, 0
	}
	v, ok := seekClosestTxn(tree, shiftedIndex(it.txn))
	if !ok {
		return nil, true, false, 0
	}

	// Index 0 is cached pre-state and is not exposed by iterators.
	// Base state comes from the parent storage iterator.
	if v.Index == 0 {
		return nil, true, false, 0
	}

	if v.Estimate {
		if it.Executing() {
			return nil, true, true, v.Version().Index
		}
		// in validation mode, it should fail validation immediately
		return nil, false, false, 0
	}

	return &v, true, false, 0
}
