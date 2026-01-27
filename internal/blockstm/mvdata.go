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
	index    *mvIndex[V]
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
		// Object stores don't necessarily have a deterministic equality relation.
		// Keep version-based validation by default.
		return NewGMVData(storetypes.AnyIsZero, storetypes.AnyValueLen, nil)
	default:
		return NewGMVData(storetypes.BytesIsZero, storetypes.BytesValueLen, bytes.Equal)
	}
}

func NewGMVData[V any](isZero func(V) bool, valueLen func(V) int, eq func(V, V) bool) *GMVData[V] {
	d := &GMVData[V]{
		index:    newMVIndex[V](),
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
func (d *GMVData[V]) getTree(key Key) *tree2.SmallBTree[secondaryDataItem[V]] {
	return d.index.get(key)
}

// getTreeOrDefault set a new tree atomically if not found.
func (d *GMVData[V]) getTreeOrDefault(key Key) *tree2.SmallBTree[secondaryDataItem[V]] {
	return d.index.getOrCreate(key)
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

	// find the closest internal version that's < shiftedIndex(txn)
	item, ok := seekClosestTxn(inner, shiftedIndex(txn))
	if !ok {
		return zero, InvalidTxnVersion, false, false
	}

	// Index 0 corresponds to the pre-state from storage (InvalidTxnVersion).
	if item.Index == 0 {
		return item.Value, InvalidTxnVersion, item.Estimate, true
	}

	return item.Value, item.Version(), item.Estimate, true
}

func (d *GMVData[V]) Iterator(
	opts IteratorOptions, txn TxnIndex,
	waitFn func(TxnIndex),
) *MVIterator[V] {
	newKeys := func() keyCursor[V] {
		return d.index.cursorKeys(opts.Start, opts.End, opts.Ascending)
	}
	cursor := newKeys()
	return NewMVIterator(opts, txn, d, cursor, newKeys, waitFn)
}

// ValidateReadSet validates that the values read during execution are consistent with the current state.
// It returns true if the read set is valid, false otherwise.
//
// Note: This function does not consult storage; callers must ensure any storage reads are cached in MVData
// at internal index 0 (pre-state) when the read happens.
func (d *GMVData[V]) ValidateReadSet(txn TxnIndex, rs *ReadSet) bool {
	for _, desc := range rs.Reads {
		if desc.Has {
			value, version, estimate, found := d.readFound(desc.Key, txn)
			if estimate {
				return false
			}
			var exists bool
			if version.Valid() {
				exists = !d.isZero(value)
			} else {
				// Storage-based Has() does not require caching; storage is immutable.
				// If the key wasn't cached into MVData, trust the recorded ExistsExpected.
				if !found {
					continue
				}
				exists = !d.isZero(value)
			}
			if exists != desc.ExistsExpected {
				return false
			}
			continue
		}

		value, version, estimate := d.Read(desc.Key, txn)
		if estimate {
			// Dependency is still estimated, so this read is invalid.
			return false
		}
		// Fast path: if versions match, the read is valid.
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
	newKeys := func() keyCursor[V] {
		return d.index.cursorKeys(desc.Start, desc.End, desc.Ascending)
	}
	cursor := newKeys()
	it := NewMVIterator(desc.IteratorOptions, txn, d, cursor, newKeys, nil)
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
	keys := d.index.snapshotAllKeys(true)
	for i := 0; i < len(keys); i++ {
		inner := d.getTree(keys[i])
		if inner == nil {
			continue
		}
		item, ok := inner.Max()
		if !ok {
			continue
		}
		if item.Estimate {
			continue
		}
		if !cb(keys[i], item.Value) {
			return
		}
	}
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
	Tree *tree2.SmallBTree[secondaryDataItem[V]]
}

func (d *dataItem[V]) Init() {
	if d.Tree == nil {
		d.Tree = tree2.NewSmallBTree(secondaryLesser[V], InnerBTreeDegree)
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
func seekClosestTxn[V any](tree *tree2.SmallBTree[secondaryDataItem[V]], txn TxnIndex) (secondaryDataItem[V], bool) {
	return tree.ReverseSeek(secondaryDataItem[V]{Index: txn - 1})
}
