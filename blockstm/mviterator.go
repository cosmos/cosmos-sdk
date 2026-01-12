package blockstm

import (
	"github.com/tidwall/btree"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/blockstm/tree"
)

// MVIterator is an iterator for a multi-versioned store.
type MVIterator[V any] struct {
	tree.BTreeIteratorG[dataItem]
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

	resolveInnerValue func(Key, TxnIndex, *BitmapIndex) (V, TxnVersion, bool)
}

var _ storetypes.Iterator = (*MVIterator[[]byte])(nil)

func NewMVIterator[V any](
	opts IteratorOptions, txn TxnIndex, iter btree.IterG[dataItem],
	waitFn func(TxnIndex),
	resolveInnerValue func(Key, TxnIndex, *BitmapIndex) (V, TxnVersion, bool),
) *MVIterator[V] {
	it := &MVIterator[V]{
		BTreeIteratorG: *tree.NewBTreeIteratorG(
			dataItem{Key: opts.Start},
			dataItem{Key: opts.End},
			iter,
			opts.Ascending,
		),
		txn:               txn,
		waitFn:            waitFn,
		resolveInnerValue: resolveInnerValue,
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
		v, ver, ok := it.resolveValueInner(inner.Item())
		if !ok {
			// abort the iterator
			it.Invalidate()
			// signal the validation to fail
			it.readEstimateValue = true
			return
		}

		if !ver.Valid() {
			// not found, skip
			continue
		}

		it.value = v
		it.version = ver
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
// - (nil, Invalid, true) if the value is not found
// - (nil, Invalid, false) if the value is an estimate and we should fail the validation
// - (v, ver, true) if the value is found at idx
func (it *MVIterator[V]) resolveValueInner(item dataItem) (V, TxnVersion, bool) {
	for {
		v, ver, estimate := it.resolveInnerValue(item.Key, it.txn, item.Index)
		if estimate {
			if it.Executing() {
				it.waitFn(ver.Index)
				continue
			}
			// in validation mode, it should fail validation immediately
			return v, ver, false
		}

		return v, ver, true
	}
}
