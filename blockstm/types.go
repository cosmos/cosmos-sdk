package blockstm

import (
	storetypes "cosmossdk.io/store/types"
)

const (
	TelemetrySubsystem = "blockstm"
	KeyExecutedTxs     = "executed_txs"
	KeyTryExecuteTime  = "try_execute_time"
	KeyValidatedTxs    = "validated_txs"
	KeyDecreaseCount   = "decrease_count"

	// MVData Metrics
	KeyMVDataRead  = "mvdata_read"
	KeyMVDataWrite = "mvdata_write"

	// MVView Metrics
	KeyMVViewReadWriteSet    = "mvview_read_writeset"
	KeyMVViewReadMVData      = "mvview_read_mvdata"
	KeyMVViewReadStorage     = "mvview_read_storage"
	KeyMVViewWrite           = "mvview_write"
	KeyMVViewDelete          = "mvview_delete"
	KeyMVViewApplyWriteSet   = "mvview_apply_writeset"
	KeyMVViewIteratorKeys    = "mvview_iterator_keys_read"
	KeyMVViewIteratorKeysCnt = "mvview_iterator_keys_read_count"

	// Executor/Transaction Metrics
	KeyTxReadCount        = "tx_read_count"
	KeyTxWriteCount       = "tx_write_count"
	KeyTxNewLocationWrite = "tx_new_location_write"
)

type (
	TxnIndex    int
	Incarnation uint
)

type TxnVersion struct {
	Index       TxnIndex
	Incarnation Incarnation
}

var InvalidTxnVersion = TxnVersion{-1, 0}

func (v TxnVersion) Valid() bool {
	return v.Index >= 0
}

type Key []byte

type ReadDescriptor struct {
	Key Key
	// invalid Version means the key is read from storage
	Version TxnVersion
}

type IteratorOptions struct {
	// [Start, End) is the range of the iterator
	Start     Key
	End       Key
	Ascending bool
}

type IteratorDescriptor struct {
	IteratorOptions
	// Stop is not `nil` if the iteration is not exhausted and stops at a key before reaching the end of the range,
	// the effective range is `[start, stop]`.
	// when replaying, it should also stops at the stop key.
	Stop Key
	// Reads is the list of keys that is observed by the iterator.
	Reads []ReadDescriptor
}

type ReadSet struct {
	Reads     []ReadDescriptor
	Iterators []IteratorDescriptor
}

type MultiReadSet = map[int]*ReadSet

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
	ValidateReadSet(TxnIndex, *ReadSet) bool
	SnapshotToStore(storetypes.Store)
}

// MVView is a value type agnostic interface for `MVMemoryView`, to keep `MultiMVMemoryView` value type agnostic.
type MVView interface {
	storetypes.Store

	ApplyWriteSet(TxnVersion) Locations
	ReadSet() *ReadSet
	WriteCount() int
}
