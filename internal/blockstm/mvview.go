package blockstm

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/metric"

	"github.com/cosmos/cosmos-sdk/store/v2/cachekv"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
)

var (
	_ storetypes.KVStore    = (*GMVMemoryView[[]byte])(nil)
	_ storetypes.ObjKVStore = (*GMVMemoryView[any])(nil)
	_ MVView                = (*GMVMemoryView[[]byte])(nil)
	_ MVView                = (*GMVMemoryView[any])(nil)
)

// GMVMemoryView wraps `MVMemory` for execution of a single transaction.
type GMVMemoryView[V any] struct {
	ctx       context.Context
	storage   storetypes.GKVStore[V]
	mvData    *GMVData[V]
	scheduler *Scheduler
	store     int

	txn      TxnIndex
	readSet  *ReadSet
	writeSet *GMemDB[V]
}

func NewMVView(ctx context.Context, store int, storage storetypes.Store, mvData MVStore, scheduler *Scheduler, txn TxnIndex) MVView {
	switch data := mvData.(type) {
	case *GMVData[any]:
		return NewGMVMemoryView(ctx, store, storage.(storetypes.ObjKVStore), data, scheduler, txn)
	case *GMVData[[]byte]:
		return NewGMVMemoryView(ctx, store, storage.(storetypes.KVStore), data, scheduler, txn)
	default:
		panic("unsupported value type")
	}
}

func NewGMVMemoryView[V any](ctx context.Context, store int, storage storetypes.GKVStore[V], mvData *GMVData[V], scheduler *Scheduler, txn TxnIndex) *GMVMemoryView[V] {
	return &GMVMemoryView[V]{
		ctx:       ctx,
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
		s.writeSet = NewWriteSet(s.mvData.isZero, s.mvData.valueLen)
	}
}

func (s *GMVMemoryView[V]) waitFor(txn TxnIndex) {
	cond := s.scheduler.WaitForDependency(s.txn, txn)
	if cond != nil {
		cond.Wait()
	}
}

func (s *GMVMemoryView[V]) ApplyWriteSet(version TxnVersion) bool {
	return s.mvData.Consolidate(s.ctx, version, s.writeSet)
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
			measureSince(s.ctx, func() metric.Int64Histogram { return inst.MVViewReadWriteSet }, start)
			// zero value means deleted
			return value
		}
	}

	for {
		value, version, estimate := s.mvData.Read(s.ctx, key, s.txn)
		if estimate {
			estimateStart := time.Now()
			// read ESTIMATE mark, wait for the blocking txn to finish
			s.waitFor(version.Index)
			measureSince(s.ctx, func() metric.Int64Histogram { return inst.MVViewEstimateWait }, estimateStart)
			continue
		}

		// record the read version, invalid version is ⊥.
		// if not found, record version ⊥ when reading from storage.
		s.readSet.Reads = append(s.readSet.Reads, ReadDescriptor{key, version})
		if !version.Valid() {
			result := s.storage.Get(key)
			measureSince(s.ctx, func() metric.Int64Histogram { return inst.MVViewReadStorage }, start)
			return result
		}
		measureSince(s.ctx, func() metric.Int64Histogram { return inst.MVViewReadMVData }, start)
		return value
	}
}

func (s *GMVMemoryView[V]) Has(key []byte) bool {
	return !s.mvData.isZero(s.Get(key))
}

func (s *GMVMemoryView[V]) Set(key []byte, value V) {
	start := time.Now()
	defer measureSince(s.ctx, func() metric.Int64Histogram { return inst.MVViewWrite }, start)
	if s.mvData.isZero(value) {
		panic("nil value is not allowed")
	}
	s.init()
	s.writeSet.OverlaySet(key, value)
}

func (s *GMVMemoryView[V]) Delete(key []byte) {
	start := time.Now()
	defer measureSince(s.ctx, func() metric.Int64Histogram { return inst.MVViewDelete }, start)
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

		measureSince(s.ctx, func() metric.Int64Histogram { return inst.MVViewIteratorKeys }, iterStart)
		if inst != nil {
			inst.MVViewIteratorKeysCnt.Add(s.ctx, int64(len(reads)))
		}
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

// GetStoreType implements types.Store.
func (s *GMVMemoryView[V]) GetStoreType() storetypes.StoreType {
	return s.storage.GetStoreType()
}
