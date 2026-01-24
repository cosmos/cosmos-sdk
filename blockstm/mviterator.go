package blockstm

import (
	"github.com/tidwall/btree"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/blockstm/tree"
)

// MVIterator is an iterator for a multi-versioned store.
type MVIterator[V any] struct {
	tree.BTreeIteratorG[dataItem[V]]
	txn TxnIndex

	// cache current found value and version
	value   V
	version TxnVersion

	// record the observed reads during iteration during execution
	reads []ReadDescriptor
	// blocking call to wait for dependent transaction to finish, `nil` in validation mode
	waitFn func(TxnIndex)
	// signal the validation to fail
	readEstimateValue bool
}

var _ storetypes.Iterator = (*MVIterator[[]byte])(nil)

func NewMVIterator[V any](
	opts IteratorOptions, txn TxnIndex, iter btree.IterG[dataItem[V]],
	waitFn func(TxnIndex),
) *MVIterator[V] {
	it := &MVIterator[V]{
		BTreeIteratorG: *tree.NewBTreeIteratorG(
			dataItem[V]{Key: opts.Start},
			dataItem[V]{Key: opts.End},
			iter,
			opts.Ascending,
		),
		txn:    txn,
		waitFn: waitFn,
	}
	it.resolveValue()
	return it
}

// Executing returns if the iterator is running in execution mode.
func (it *MVIterator[V]) Executing() bool {
	return it.waitFn != nil
}

func (it *MVIterator[V]) Next() {
	it.BTreeIteratorG.Next()
	it.resolveValue()
}

func (it *MVIterator[V]) Value() V {
	return it.value
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
	inner := &it.BTreeIteratorG
	for ; inner.Valid(); inner.Next() {
		v, ok := it.resolveValueInner(inner.Item().Tree)
		if !ok {
			// abort the iterator
			it.Invalidate()
			// signal the validation to fail
			it.readEstimateValue = true
			return
		}
		if v == nil {
			continue
		}

		it.value = v.Value
		it.version = v.Version()
		if it.Executing() {
			it.reads = append(it.reads, ReadDescriptor{
				Key:     inner.Item().Key,
				Version: it.version,
			})
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
func (it *MVIterator[V]) resolveValueInner(tree *tree.BTree[secondaryDataItem[V]]) (*secondaryDataItem[V], bool) {
	for {
		v, ok := seekClosestTxn(tree, it.txn)
		if !ok {
			return nil, true
		}

		if v.Estimate {
			if it.Executing() {
				// estimate won't be set by storage index.
				it.waitFn(FromShiftedIndex(v.Index))
				continue
			}
			// in validation mode, it should fail validation immediately
			return nil, false
		}

		return &v, true
	}
}
