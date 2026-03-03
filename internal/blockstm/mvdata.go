package blockstm

import (
	"bytes"
	"sync/atomic"
	"time"

	storetypes "cosmossdk.io/store/types"

	tree2 "github.com/cosmos/cosmos-sdk/internal/blockstm/tree"
	"github.com/cosmos/cosmos-sdk/telemetry"
)

const (
	OuterBTreeDegree = 4 // Since we do copy-on-write a lot, smaller degree means smaller allocations
	InnerBTreeDegree = 4
)

type dataEntry[V any] struct {
	Incarnation Incarnation
	WriteSet    *GMemDB[V]

	// mark all the writes in this txn as ESTIMATE
	Estimate bool
}

type MVData = GMVData[[]byte]

func NewMVData(blockSize int) *MVData {
	return NewGMVData(blockSize, storetypes.BytesIsZero, storetypes.BytesValueLen)
}

type GMVData[V any] struct {
	isZero   func(V) bool
	valueLen func(V) int

	// key -> bitmap(txn)
	index tree2.BTree[indexEntry]
	// txn -> (incarnation, estimate, key -> value)
	data []atomic.Pointer[dataEntry[V]]
}

func NewMVStore(key storetypes.StoreKey, blockSize int) MVStore {
	switch key.(type) {
	case *storetypes.ObjectStoreKey:
		return NewGMVData(blockSize, storetypes.AnyIsZero, storetypes.AnyValueLen)
	default:
		return NewGMVData(blockSize, storetypes.BytesIsZero, storetypes.BytesValueLen)
	}
}

func NewGMVData[V any](blockSize int, isZero func(V) bool, valueLen func(V) int) *GMVData[V] {
	return &GMVData[V]{
		isZero:   isZero,
		valueLen: valueLen,

		index: *tree2.NewBTree(tree2.KeyItemLess[indexEntry], OuterBTreeDegree),
		data:  make([]atomic.Pointer[dataEntry[V]], blockSize),
	}
}

// getStore returns `nil` if not found
func (d *GMVData[V]) getStore(key Key) *BitmapIndex {
	outer, _ := d.index.Get(indexEntry{Key: key})
	return outer.Index
}

// getTreeOrDefault set a new tree atomically if not found.
func (d *GMVData[V]) getStoreOrDefault(key Key) *BitmapIndex {
	return d.index.GetOrDefault(indexEntry{Key: key}, (*indexEntry).Init).Index
}

// Consolidate returns wroteNewLocation
func (d *GMVData[V]) Consolidate(version TxnVersion, writeSet *GMemDB[V]) bool {
	defer telemetry.MeasureSince(time.Now(), TelemetrySubsystem, KeyMVDataConsolidate) //nolint:staticcheck // TODO: switch to OpenTelemetry

	if writeSet == nil || writeSet.Len() == 0 {
		// delete old indexes
		d.ConsolidateEmpty(version.Index)
		return false
	}

	prevData := d.data[version.Index].Swap(&dataEntry[V]{
		Incarnation: version.Incarnation,
		WriteSet:    writeSet,
	})

	var wroteNewLocation bool
	if prevData == nil {
		writeSet.Scan(func(key Key, _ V) bool {
			d.Set(key, version.Index)
			return true
		})
		wroteNewLocation = true
	} else {
		// diff writeSet to update indexes
		DiffMemDB(prevData.WriteSet, writeSet, func(key Key, is_new bool) bool {
			if is_new {
				// new key, add to index
				d.Set(key, version.Index)
				wroteNewLocation = true
			} else {
				// deleted key, delete from index
				d.Delete(key, version.Index)
			}
			return true
		})
	}

	return wroteNewLocation
}

func (d *GMVData[V]) ConsolidateEmpty(txn TxnIndex) {
	old := d.data[txn].Swap(nil)
	if old != nil {
		old.WriteSet.Scan(func(key Key, _ V) bool {
			d.Delete(key, txn)
			return true
		})
	}
}

func (d *GMVData[V]) ConvertWritesToEstimates(txn TxnIndex) {
	for {
		old := d.data[txn].Load()
		if old == nil {
			// nothing to mark
			return
		}
		if old.Estimate {
			return
		}

		new := *old
		new.Estimate = true
		if d.data[txn].CompareAndSwap(old, &new) {
			break
		}
	}
}

func (d *GMVData[V]) ClearEstimates(txn TxnIndex) {
	for {
		old := d.data[txn].Load()
		if old == nil {
			// nothing to remove
			return
		}
		if !old.Estimate {
			return
		}

		new := *old
		new.Estimate = false
		if d.data[txn].CompareAndSwap(old, &new) {
			break
		}
	}
}

func (d *GMVData[V]) InitWithEstimates(txn TxnIndex, estimates Locations) {
	for _, key := range estimates {
		d.Set(key, txn)
	}
	d.data[txn].Store(&dataEntry[V]{
		Estimate: true,
		WriteSet: NewWriteSet(d.isZero, d.valueLen),
	})
}

// Set add txn to the key's bitmap index.
func (d *GMVData[V]) Set(key Key, txn TxnIndex) {
	tree := d.getStoreOrDefault(key)
	tree.Set(txn)
}

// Delete removes txn from the key's bitmap index.
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
	defer telemetry.MeasureSince(time.Now(), TelemetrySubsystem, KeyMVDataRead) //nolint:staticcheck // TODO: switch to OpenTelemetry
	var zero V
	if txn == 0 {
		return zero, InvalidTxnVersion, false
	}

	store := d.getStore(key)
	if store == nil {
		return zero, InvalidTxnVersion, false
	}

	return d.resolveValue(key, txn, store)
}

func (d *GMVData[V]) resolveValue(key Key, txn TxnIndex, store *BitmapIndex) (V, TxnVersion, bool) {
	var zero V
	for {
		if txn == 0 {
			return zero, InvalidTxnVersion, false
		}

		// find the closest txn that's less than the given txn
		idx, ok := store.PreviousValue(txn)
		if !ok {
			return zero, InvalidTxnVersion, false
		}

		entry := d.data[idx].Load()
		if entry == nil {
			// could happen because we don't synchronize bitmap and data, just try again to find the next closest txn
			txn = idx
			continue
		}

		if entry.Estimate {
			// ESTIMATE mark
			return zero, TxnVersion{Index: idx, Incarnation: 0}, true
		}

		v, ok := entry.WriteSet.OverlayGet(key)
		if !ok {
			// could happen because we don't synchronize bitmap and data, just try again to find the next closest txn
			txn = idx
			continue
		}

		return v, TxnVersion{Index: idx, Incarnation: entry.Incarnation}, false
	}
}

func (d *GMVData[V]) Iterator(
	opts IteratorOptions, txn TxnIndex,
	waitFn func(TxnIndex),
) *MVIterator[V] {
	return NewMVIterator(opts, txn, d.index.Iter(), waitFn, d.resolveValue)
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
	it := NewMVIterator(desc.IteratorOptions, txn, d.index.Iter(), nil, d.resolveValue)
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
	d.index.Scan(func(outer indexEntry) bool {
		txn, ok := outer.Index.Max()
		if !ok {
			return true
		}

		v, ok := d.data[txn].Load().WriteSet.OverlayGet(outer.Key)
		if !ok {
			return true
		}

		return cb(outer.Key, v)
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

type indexEntry struct {
	Key   Key
	Index *BitmapIndex
}

func (d *indexEntry) Init() {
	if d.Index == nil {
		d.Index = NewBitmapIndex()
	}
}

var _ tree2.KeyItem = indexEntry{}

func (item indexEntry) GetKey() []byte {
	return item.Key
}
