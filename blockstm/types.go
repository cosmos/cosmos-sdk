package blockstm

import (
	storetypes "cosmossdk.io/store/types"
)

const (
	TelemetrySubsystem = "blockstm"
	KeyExecutedTxs     = "executed_txs"
	KeyValidatedTxs    = "validated_txs"
)

type (
	TxnIndex    int
	Incarnation uint
	Wave        uint32

	// ShiftedTxnIndex shift txn index by 1 to reserve 0 as storage state.
	// invariant: txn index + 1 < = max uint32
	// proof: block size is verified in entry api.
	ShiftedTxnIndex uint32
)

type TxnVersion struct {
	Index       TxnIndex
	Incarnation Incarnation
}

var InvalidTxnVersion = TxnVersion{-1, 0}

func (v TxnVersion) Valid() bool {
	return v.Index >= 0
}

func ToShiftedIndex(t TxnIndex) ShiftedTxnIndex {
	return ShiftedTxnIndex(t + 1)
}

// FromShiftedIndex converts ShiftedTxnIndex back to TxnIndex,
// it returns -1 if the shifted index is for storage.
func FromShiftedIndex(s ShiftedTxnIndex) TxnIndex {
	return TxnIndex(s) - 1
}

type Predicate[V any] func(V) bool

type Key []byte

type ReadDescriptor[V any] struct {
	Key Key
	// invalid Version means the key is read from storage
	Version TxnVersion

	// if Predicate is not nil, we only observed partial information about the value,
	// so we only need to ensure that the value satisfies the predicate in validation.
	Predicate Predicate[V]
	Observed  bool
}

func NewReadDescriptor[V any](key Key, version TxnVersion) ReadDescriptor[V] {
	return ReadDescriptor[V]{
		Key:     key,
		Version: version,
	}
}

func NewReadDescriptorWithPredicate[V any](
	key Key,
	version TxnVersion,
	predicate Predicate[V],
	observed bool,
) ReadDescriptor[V] {
	return ReadDescriptor[V]{
		Key:       key,
		Version:   version,
		Predicate: predicate,
		Observed:  observed,
	}
}

func (r ReadDescriptor[V]) Validate(value V, version TxnVersion) bool {
	// same version implies same value
	if r.Version == version {
		return true
	}

	if r.Predicate != nil {
		return r.Predicate(value) == r.Observed
	}

	return false
}

type IteratorOptions struct {
	// [Start, End) is the range of the iterator
	Start     Key
	End       Key
	Ascending bool
}

type IteratorDescriptor[V any] struct {
	IteratorOptions
	// Stop is not `nil` if the iteration is not exhausted and stops at a key before reaching the end of the range,
	// the effective range is `[start, stop]`.
	// when replaying, it should also stops at the stop key.
	Stop Key
	// Reads is the list of keys that is observed by the iterator.
	Reads []ReadDescriptor[V]
}

type ReadSet[V any] struct {
	Reads     []ReadDescriptor[V]
	Iterators []IteratorDescriptor[V]
}

type MultiReadSet = map[int]any

// TxExecutor executes transactions on top of a multi-version memory view.
type TxExecutor func(TxnIndex, MultiStore)

type MultiStore interface {
	GetStore(storetypes.StoreKey) storetypes.Store
	GetKVStore(storetypes.StoreKey) storetypes.KVStore
	GetObjKVStore(storetypes.StoreKey) storetypes.ObjKVStore
}

// MVStore is a value type agnostic interface for `MVData`, to keep `MVMemory` value type agnostic.
type MVStore interface {
	Delete(Key, TxnIndex)
	WriteEstimate(Key, TxnIndex)
	ValidateReadSet(TxnIndex, any) bool
	SnapshotToStore(storetypes.Store)
}

// MVView is a value type agnostic interface for `MVMemoryView`, to keep `MultiMVMemoryView` value type agnostic.
type MVView interface {
	storetypes.Store

	ApplyWriteSet(TxnVersion) Locations
	ReadSet() any
}
