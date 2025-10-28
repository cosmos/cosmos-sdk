package iavlx

import (
	io "io"

	"cosmossdk.io/store/types"
)

type Tracer interface {
	StartSpan(operation TraceOperation, kvs ...any) Span
	WithContextPtr(ctxPtr any) Tracer
}

type TraceOperation int

const (
	TraceOperationGet TraceOperation = iota
	TraceOperationHas
	TraceOperationSet
	TraceOperationDelete
	TraceOperationIterator
	TraceOperationReverseIterator
	TraceOperationIteratorNext
	TraceOperationIteratorClose
	TraceOperationWorkingHash
	TraceOperationCommit
)

type Span interface {
	End(kvs ...any)
}

type TraceStore struct {
	store  types.KVStore
	tracer Tracer
}

func (t *TraceStore) GetStoreType() types.StoreType {
	return t.store.GetStoreType()
}

func (t *TraceStore) CacheWrap() types.CacheWrap {
	panic("cannot CacheWrap a TraceStore")
}

func (t *TraceStore) CacheWrapWithTrace(io.Writer, types.TraceContext) types.CacheWrap {
	panic("cannot CacheWrap a TraceStore")
}

func (t *TraceStore) Get(key []byte) []byte {
	span := t.tracer.StartSpan(TraceOperationGet, "key", key)
	value := t.store.Get(key)
	span.End("value", value)
	return value
}

func (t *TraceStore) Has(key []byte) bool {
	span := t.tracer.StartSpan(TraceOperationHas, "key", key)
	value := t.store.Has(key)
	span.End("has", value)
	return value
}

func (t *TraceStore) Set(key, value []byte) {
	span := t.tracer.StartSpan(TraceOperationSet, "key", key, "value", value)
	defer span.End()
	t.store.Set(key, value)
}

func (t *TraceStore) Delete(key []byte) {
	span := t.tracer.StartSpan(TraceOperationDelete, "key", key)
	defer span.End()
	t.store.Delete(key)
}

func (t *TraceStore) Iterator(start, end []byte) types.Iterator {
	span := t.tracer.StartSpan(TraceOperationIterator, "start", start, "end", end)
	iter := t.store.Iterator(start, end)
	span.End()
	return &traceIterator{iter: iter, tracer: t.tracer.WithContextPtr(iter)}
}

func (t *TraceStore) ReverseIterator(start, end []byte) types.Iterator {
	span := t.tracer.StartSpan(TraceOperationReverseIterator, "start", start, "end", end)
	iter := t.store.ReverseIterator(start, end)
	span.End()
	return &traceIterator{iter: iter, tracer: t.tracer.WithContextPtr(iter)}
}

type traceIterator struct {
	iter   types.Iterator
	tracer Tracer
}

func (t *traceIterator) Domain() (start []byte, end []byte) {
	return t.iter.Domain()
}

func (t *traceIterator) Valid() bool {
	return t.iter.Valid()
}

func (t *traceIterator) Next() {
	span := t.tracer.StartSpan(TraceOperationIteratorNext)
	defer span.End()
	t.iter.Next()
}

func (t *traceIterator) Key() (key []byte) {
	return t.iter.Key()
}

func (t *traceIterator) Value() (value []byte) {
	return t.iter.Value()
}

func (t *traceIterator) Error() error {
	return t.iter.Error()
}

func (t *traceIterator) Close() error {
	span := t.tracer.StartSpan(TraceOperationIteratorClose)
	defer span.End()
	return t.iter.Close()
}

var _ types.KVStore = (*TraceStore)(nil)
var _ types.Iterator = (*traceIterator)(nil)
