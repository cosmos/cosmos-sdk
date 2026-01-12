package blockstm

import (
	"bytes"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/blockstm/tree"
)

const (
	OuterBTreeDegree = 4 // Since we do copy-on-write a lot, smaller degree means smaller allocations
	InnerBTreeDegree = 4
)

type MVData = GMVData[[]byte]

func NewMVData() *MVData {
	return NewGMVData(storetypes.BytesIsZero, storetypes.BytesValueLen)
}

type GMVData[V any] struct {
	tree.BTree[dataItem[V]]
	isZero   func(V) bool
	valueLen func(V) int
}

func NewMVStore(key storetypes.StoreKey) MVStore {
	switch key.(type) {
	case *storetypes.ObjectStoreKey:
		return NewGMVData(storetypes.AnyIsZero, storetypes.AnyValueLen)
	default:
		return NewGMVData(storetypes.BytesIsZero, storetypes.BytesValueLen)
	}
}

func NewGMVData[V any](isZero func(V) bool, valueLen func(V) int) *GMVData[V] {
	return &GMVData[V]{
		BTree:    *tree.NewBTree(tree.KeyItemLess[dataItem[V]], OuterBTreeDegree),
		isZero:   isZero,
		valueLen: valueLen,
	}
}

// getStore returns `nil` if not found
func (d *GMVData[V]) getStore(key Key) *SecondaryStore[V] {
	outer, _ := d.Get(dataItem[V]{Key: key})
	return outer.Store
}

// getStoreOrDefault set a new tree atomically if not found.
func (d *GMVData[V]) getStoreOrDefault(key Key) *SecondaryStore[V] {
	return d.GetOrDefault(dataItem[V]{Key: key}, func(item *dataItem[V]) {
		if item.Store == nil {
			item.Store = NewSecondaryStore[V]()
		}
	}).Store
}

func (d *GMVData[V]) Write(key Key, value V, version TxnVersion) {
	tree := d.getStoreOrDefault(key)
	tree.Set(version.Index, secondaryDataItem[V]{Incarnation: version.Incarnation, Value: value})
}

func (d *GMVData[V]) WriteEstimate(key Key, txn TxnIndex) {
	tree := d.getStoreOrDefault(key)
	tree.Set(txn, secondaryDataItem[V]{Estimate: true})
}

func (d *GMVData[V]) Delete(key Key, txn TxnIndex) {
	tree := d.getStore(key)
	if tree != nil {
		tree.Delete(txn)
	}
}

// Read returns the value and the version of the value that's less than the given txn.
// If the key is not found, returns `(zero, InvalidTxnVersion, false)`.
// If the key is found but value is an estimate, returns `(value, version, true)`.
// If the key is found, returns `(value, version, false)`, `value` can be zero value which means deleted.
func (d *GMVData[V]) Read(key Key, txn TxnIndex) (V, TxnVersion, bool) {
	var zero V
	if txn == 0 {
		return zero, InvalidTxnVersion, false
	}

	store := d.getStore(key)
	if store == nil {
		return zero, InvalidTxnVersion, false
	}

	// find the closest txn that's less than the given txn
	idx, item, ok := store.PreviousValue(txn)
	if !ok {
		return zero, InvalidTxnVersion, false
	}

	return item.Value, TxnVersion{idx, item.Incarnation}, item.Estimate
}

func (d *GMVData[V]) Iterator(
	opts IteratorOptions, txn TxnIndex,
	waitFn func(TxnIndex),
) *MVIterator[V] {
	return NewMVIterator(opts, txn, d.Iter(), waitFn)
}

// ValidateReadSet validates the read descriptors,
// returns true if valid.
func (d *GMVData[V]) ValidateReadSet(txn TxnIndex, rs *ReadSet) bool {
	for _, desc := range rs.Reads {
		_, version, estimate := d.Read(desc.Key, txn)
		if estimate {
			// previously read entry from data, now ESTIMATE
			return false
		}
		if version != desc.Version {
			// previously read entry from data, now NOT_FOUND,
			// or read some entry, but not the same version as before
			return false
		}
	}

	for _, desc := range rs.Iterators {
		if !d.validateIterator(desc, txn) {
			return false
		}
	}

	return true
}

// validateIterator validates the iteration descriptor by replaying and compare the recorded reads.
// returns true if valid.
func (d *GMVData[V]) validateIterator(desc IteratorDescriptor, txn TxnIndex) bool {
	it := NewMVIterator(desc.IteratorOptions, txn, d.Iter(), nil)
	defer it.Close()

	var i int
	for ; it.Valid(); it.Next() {
		if desc.Stop != nil {
			if BytesBeyond(it.Key(), desc.Stop, desc.Ascending) {
				break
			}
		}

		if i >= len(desc.Reads) {
			return false
		}

		read := desc.Reads[i]
		if read.Version != it.Version() || !bytes.Equal(read.Key, it.Key()) {
			return false
		}

		i++
	}

	// we read an estimate value, fail the validation.
	if it.ReadEstimateValue() {
		return false
	}

	return i == len(desc.Reads)
}

func (d *GMVData[V]) Snapshot() (snapshot []GKVPair[V]) {
	d.SnapshotTo(func(key Key, value V) bool {
		snapshot = append(snapshot, GKVPair[V]{key, value})
		return true
	})
	return snapshot
}

func (d *GMVData[V]) SnapshotTo(cb func(Key, V) bool) {
	d.Scan(func(outer dataItem[V]) bool {
		item, ok := outer.Store.Max()
		if !ok {
			return true
		}

		if item.Estimate {
			// should not happen, just to keep it complete
			return true
		}

		return cb(outer.Key, item.Value)
	})
}

func (d *GMVData[V]) SnapshotToStore(store storetypes.Store) {
	kv := store.(storetypes.GKVStore[V])
	d.SnapshotTo(func(key Key, value V) bool {
		if d.isZero(value) {
			kv.Delete(key)
		} else {
			kv.Set(key, value)
		}
		return true
	})
}

type GKVPair[V any] struct {
	Key   Key
	Value V
}
type KVPair = GKVPair[[]byte]

type dataItem[V any] struct {
	Key   Key
	Store *SecondaryStore[V]
}

var _ tree.KeyItem = dataItem[[]byte]{}

func (item dataItem[V]) GetKey() []byte {
	return item.Key
}
