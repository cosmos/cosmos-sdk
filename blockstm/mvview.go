package blockstm

import (
	"io"
	"time"

	"cosmossdk.io/store/cachekv"
	"cosmossdk.io/store/tracekv"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
)

var (
	_ storetypes.KVStore    = (*GMVMemoryView[[]byte])(nil)
	_ storetypes.ObjKVStore = (*GMVMemoryView[any])(nil)
	_ MVView                = (*GMVMemoryView[[]byte])(nil)
	_ MVView                = (*GMVMemoryView[any])(nil)
)

// GMVMemoryView wraps `MVMemory` for execution of a single transaction.
type GMVMemoryView[V any] struct {
	storage   storetypes.GKVStore[V]
	mvData    *GMVData[V]
	scheduler *Scheduler
	store     int

	txn      TxnIndex
	readSet  *ReadSet
	writeSet *GMemDB[V]
}

func NewMVView(store int, storage storetypes.Store, mvData MVStore, scheduler *Scheduler, txn TxnIndex) MVView {
	switch data := mvData.(type) {
	case *GMVData[any]:
		return NewGMVMemoryView(store, storage.(storetypes.ObjKVStore), data, scheduler, txn)
	case *GMVData[[]byte]:
		return NewGMVMemoryView(store, storage.(storetypes.KVStore), data, scheduler, txn)
	default:
		panic("unsupported value type")
	}
}

func NewGMVMemoryView[V any](store int, storage storetypes.GKVStore[V], mvData *GMVData[V], scheduler *Scheduler, txn TxnIndex) *GMVMemoryView[V] {
	return &GMVMemoryView[V]{
		store:     store,
		storage:   storage,
		mvData:    mvData,
		scheduler: scheduler,
		txn:       txn,
		readSet:   new(ReadSet),
	}
}

func (s *GMVMemoryView[V]) init() {
	if s.writeSet == nil {
		s.writeSet = NewGMemDBNonConcurrent(s.mvData.isZero, s.mvData.valueLen)
	}
}

func (s *GMVMemoryView[V]) waitFor(txn TxnIndex) {
	cond := s.scheduler.WaitForDependency(s.txn, txn)
	if cond != nil {
		cond.Wait()
	}
}

func (s *GMVMemoryView[V]) ApplyWriteSet(version TxnVersion) Locations {
	defer telemetry.MeasureSince(time.Now(), TelemetrySubsystem, KeyMVViewApplyWriteSet)
	if s.writeSet == nil || s.writeSet.Len() == 0 {
		return nil
	}

	newLocations := make([]Key, 0, s.writeSet.Len())
	s.writeSet.Scan(func(key Key, value V) bool {
		s.mvData.Write(key, value, version)
		newLocations = append(newLocations, key)
		return true
	})

	return newLocations
}

func (s *GMVMemoryView[V]) ReadSet() *ReadSet {
	return s.readSet
}

func (s *GMVMemoryView[V]) WriteCount() int {
	if s.writeSet == nil {
		return 0
	}
	return s.writeSet.Len()
}

func (s *GMVMemoryView[V]) Get(key []byte) V {
	start := time.Now()
	if s.writeSet != nil {
		if value, found := s.writeSet.OverlayGet(key); found {
			// value written by this txn
			// nil value means deleted
			telemetry.MeasureSince(start, TelemetrySubsystem, KeyMVViewReadWriteSet)
			return value
		}
	}

	for {
		value, version, estimate := s.mvData.Read(key, s.txn)
		if estimate {
			estimateStart := time.Now()
			// read ESTIMATE mark, wait for the blocking txn to finish
			s.waitFor(version.Index)
			telemetry.MeasureSince(estimateStart, TelemetrySubsystem, KeyMVViewEstimateWait)
			continue
		}

		// record the read version, invalid version is ⊥.
		// if not found, record version ⊥ when reading from storage.
		s.readSet.Reads = append(s.readSet.Reads, ReadDescriptor{key, version})
		if !version.Valid() {
			result := s.storage.Get(key)
			telemetry.MeasureSince(start, TelemetrySubsystem, KeyMVViewReadStorage)
			return result
		}
		telemetry.MeasureSince(start, TelemetrySubsystem, KeyMVViewReadMVData)
		return value
	}
}

func (s *GMVMemoryView[V]) Has(key []byte) bool {
	return !s.mvData.isZero(s.Get(key))
}

func (s *GMVMemoryView[V]) Set(key []byte, value V) {
	defer telemetry.MeasureSince(time.Now(), TelemetrySubsystem, KeyMVViewWrite)
	if s.mvData.isZero(value) {
		panic("nil value is not allowed")
	}
	s.init()
	s.writeSet.OverlaySet(key, value)
}

func (s *GMVMemoryView[V]) Delete(key []byte) {
	defer telemetry.MeasureSince(time.Now(), TelemetrySubsystem, KeyMVViewDelete)
	var empty V
	s.init()
	s.writeSet.OverlaySet(key, empty)
}

func (s *GMVMemoryView[V]) Iterator(start, end []byte) storetypes.GIterator[V] {
	return s.iterator(IteratorOptions{Start: start, End: end, Ascending: true})
}

func (s *GMVMemoryView[V]) ReverseIterator(start, end []byte) storetypes.GIterator[V] {
	return s.iterator(IteratorOptions{Start: start, End: end, Ascending: false})
}

func (s *GMVMemoryView[V]) iterator(opts IteratorOptions) storetypes.GIterator[V] {
	iterStart := time.Now()
	mvIter := s.mvData.Iterator(opts, s.txn, s.waitFor)

	var parentIter, wsIter storetypes.GIterator[V]

	if s.writeSet == nil {
		wsIter = NewNoopIterator[V](opts.Start, opts.End, opts.Ascending)
	} else {
		wsIter = s.writeSet.iterator(opts.Start, opts.End, opts.Ascending)
	}

	if opts.Ascending {
		parentIter = s.storage.Iterator(opts.Start, opts.End)
	} else {
		parentIter = s.storage.ReverseIterator(opts.Start, opts.End)
	}

	onClose := func(iter storetypes.GIterator[V]) {
		reads := mvIter.Reads()

		var stopKey Key
		if iter.Valid() {
			stopKey = iter.Key()

			// if the iterator is not exhausted, the merge iterator may have read one more key which is not observed by
			// the caller, in that case we remove that read descriptor.
			if len(reads) > 0 {
				lastRead := reads[len(reads)-1].Key
				if BytesBeyond(lastRead, stopKey, opts.Ascending) {
					reads = reads[:len(reads)-1]
				}
			}
		}

		s.readSet.Iterators = append(s.readSet.Iterators, IteratorDescriptor{
			IteratorOptions: opts,
			Stop:            stopKey,
			Reads:           reads,
		})

		// Measure iterator duration and track keys read
		telemetry.MeasureSince(iterStart, TelemetrySubsystem, KeyMVViewIteratorKeys)
		telemetry.IncrCounter(float32(len(reads)), TelemetrySubsystem, KeyMVViewIteratorKeysCnt) //nolint:staticcheck // TODO: switch to OpenTelemetry
	}

	// three-way merge iterator
	return NewCacheMergeIterator(
		NewCacheMergeIterator(parentIter, mvIter, opts.Ascending, nil, s.mvData.isZero),
		wsIter,
		opts.Ascending,
		onClose,
		s.mvData.isZero,
	)
}

// CacheWrap implements types.Store.
func (s *GMVMemoryView[V]) CacheWrap() storetypes.CacheWrap {
	return cachekv.NewGStore(s, s.mvData.isZero, s.mvData.valueLen)
}

// CacheWrapWithTrace implements types.Store.
func (s *GMVMemoryView[V]) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	if store, ok := any(s).(*GMVMemoryView[[]byte]); ok {
		return cachekv.NewGStore(tracekv.NewStore(store, w, tc), store.mvData.isZero, store.mvData.valueLen)
	}
	return s.CacheWrap()
}

// GetStoreType implements types.Store.
func (s *GMVMemoryView[V]) GetStoreType() storetypes.StoreType {
	return s.storage.GetStoreType()
}
