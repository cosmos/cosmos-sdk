package blockstm

import (
	"bytes"

	storetypes "cosmossdk.io/store/types"

	tree2 "github.com/cosmos/cosmos-sdk/internal/blockstm/tree"
)

const (
	OuterBTreeDegree = 4 // Since we do copy-on-write a lot, smaller degree means smaller allocations
	InnerBTreeDegree = 4

	// maxValueSnapshotBytes limits captured value snapshots, larger values validate by version only.
	maxValueSnapshotBytes = 16 << 10 // 16KiB
)

type MVData = GMVData[[]byte]

func NewMVData() *MVData {
	return NewGMVData(storetypes.BytesIsZero, storetypes.BytesValueLen, bytes.Equal)
}

type GMVData[V any] struct {
	tree2.BTree[dataItem[V]]
	isZero   func(V) bool
	valueLen func(V) int
	// eq enables value-based validation, nil means version-only validation.
	eq      func(V, V) bool
	isBytes bool
}

func (d *GMVData[V]) shouldSnapshotValue(value V) bool {
	if d.valueLen == nil {
		return false
	}
	sz := d.valueLen(value)
	return sz >= 0 && sz <= maxValueSnapshotBytes
}

func (d *GMVData[V]) captureBytesIfSmall(value V) []byte {
	if !d.isBytes || !d.shouldSnapshotValue(value) {
		return nil
	}
	b, ok := any(value).([]byte)
	if !ok {
		return nil
	}
	return append([]byte(nil), b...)
}

func NewMVStore(key storetypes.StoreKey) MVStore {
	switch key.(type) {
	case *storetypes.ObjectStoreKey:
		return NewGMVData(storetypes.AnyIsZero, storetypes.AnyValueLen, nil)
	default:
		return NewGMVData(storetypes.BytesIsZero, storetypes.BytesValueLen, bytes.Equal)
	}
}

func NewGMVData[V any](isZero func(V) bool, valueLen func(V) int, eq func(V, V) bool) *GMVData[V] {
	d := &GMVData[V]{
		BTree:    *tree2.NewBTree(tree2.KeyItemLess[dataItem[V]], OuterBTreeDegree),
		isZero:   isZero,
		valueLen: valueLen,
		eq:       eq,
	}
	var z V
	if _, ok := any(z).([]byte); ok {
		d.isBytes = true
	}
	return d
}

// getTree returns `nil` if not found
func (d *GMVData[V]) getTree(key Key) *tree2.BTree[secondaryDataItem[V]] {
	outer, _ := d.Get(dataItem[V]{Key: key})
	return outer.Tree
}

// getTreeOrDefault set a new tree atomically if not found.
func (d *GMVData[V]) getTreeOrDefault(key Key) *tree2.BTree[secondaryDataItem[V]] {
	return d.GetOrDefault(dataItem[V]{Key: key}, (*dataItem[V]).Init).Tree
}

func shiftedIndex(txn TxnIndex) TxnIndex {
	// Reserve internal index 0 (historical: used for cached pre-state).
	return txn + 1
}

func (d *GMVData[V]) Write(key Key, value V, version TxnVersion) {
	tree := d.getTreeOrDefault(key)
	tree.Set(secondaryDataItem[V]{Index: shiftedIndex(version.Index), Incarnation: version.Incarnation, Value: value})
}

func (d *GMVData[V]) WriteEstimate(key Key, txn TxnIndex) {
	tree := d.getTreeOrDefault(key)
	tree.Set(secondaryDataItem[V]{Index: shiftedIndex(txn), Estimate: true})
}

func (d *GMVData[V]) Delete(key Key, txn TxnIndex) {
	tree := d.getTreeOrDefault(key)
	tree.Delete(secondaryDataItem[V]{Index: shiftedIndex(txn)})
}

// Read returns the value and the version of the value that's less than the given txn.
// If the key is not found, returns `(zero, InvalidTxnVersion, false)`.
// If the key is found but value is an estimate, returns `(value, version, true)`.
// If the key is found, returns `(value, version, false)`, `value` can be zero value which means deleted.
func (d *GMVData[V]) Read(key Key, txn TxnIndex) (V, TxnVersion, bool) {
	v, ver, est, _ := d.readFound(key, txn)
	return v, ver, est
}

func (d *GMVData[V]) readFound(key Key, txn TxnIndex) (V, TxnVersion, bool, bool) {
	var zero V
	inner := d.getTree(key)
	if inner == nil {
		return zero, InvalidTxnVersion, false, false
	}

	// find the closest txn that's less than the given txn
	item, ok := seekClosestTxn(inner, shiftedIndex(txn))
	if !ok {
		return zero, InvalidTxnVersion, false, false
	}

	// Internal index 0 represents cached pre-state (storage). Externally, we keep
	// InvalidTxnVersion semantics for storage reads.
	if item.Index == 0 {
		return item.Value, InvalidTxnVersion, item.Estimate, true
	}

	return item.Value, item.Version(), item.Estimate, true
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
		value, version, estimate := d.Read(desc.Key, txn)
		if estimate {
			// previously read entry from data, now ESTIMATE
			return false
		}
		if version == desc.Version {
			continue
		}

		// Validation failed on version comparison.
		// If the original read was from storage (InvalidTxnVersion), we can re-verify against storage.
		if !desc.Version.Valid() {
			// If the current value is also from storage (version is invalid), it matches.
			if !version.Valid() {
				continue
			}

			// Storage vs Versioned:
			// The current value is a new version. Check if its value matches the cached pre-state (Index 0).
			if d.eq != nil {
				if inner := d.getTree(desc.Key); inner != nil {
					if item, ok := inner.Get(secondaryDataItem[V]{Index: 0}); ok {
						if d.eq(value, item.Value) {
							continue
						}
					}
				}
			}
			return false
		}

		// Versioned vs Versioned:
		// If we captured the value during execution, we can compare it with the current value.
		// This handles ABA scenarios where a value is updated but set to the same content.
		if d.isBytes && desc.Captured != nil {
			if cur, ok := any(value).([]byte); ok {
				if bytes.Equal(desc.Captured, cur) {
					continue
				}
			}
		}

		return false
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
		item, ok := outer.Tree.Max()
		if !ok {
			return true
		}

		if item.Estimate {
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
	Key  Key
	Tree *tree2.BTree[secondaryDataItem[V]]
}

func (d *dataItem[V]) Init() {
	if d.Tree == nil {
		d.Tree = tree2.NewBTree(secondaryLesser[V], InnerBTreeDegree)
	}
}

var _ tree2.KeyItem = dataItem[[]byte]{}

func (item dataItem[V]) GetKey() []byte {
	return item.Key
}

type secondaryDataItem[V any] struct {
	Index       TxnIndex
	Incarnation Incarnation
	Value       V
	Estimate    bool
}

func secondaryLesser[V any](a, b secondaryDataItem[V]) bool {
	return a.Index < b.Index
}

func (item secondaryDataItem[V]) Version() TxnVersion {
	if item.Index == 0 {
		return InvalidTxnVersion
	}
	return TxnVersion{Index: item.Index - 1, Incarnation: item.Incarnation}
}

// seekClosestTxn returns the closest txn that's less than the given txn.
func seekClosestTxn[V any](tree *tree2.BTree[secondaryDataItem[V]], txn TxnIndex) (secondaryDataItem[V], bool) {
	return tree.ReverseSeek(secondaryDataItem[V]{Index: txn - 1})
}
