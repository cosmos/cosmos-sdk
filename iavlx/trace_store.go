package iavlx

import (
	io "io"

	"cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
)

type TraceStore struct {
	store  types.KVStore
	tracer telemetry.Tracer
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
	span := t.tracer.StartSpan("get", "key", key)
	value := t.store.Get(key)
	span.End("value", value)
	return value
}

func (t *TraceStore) Has(key []byte) bool {
	span := t.tracer.StartSpan("has", "key", key)
	value := t.store.Has(key)
	span.End("has", value)
	return value
}

func (t *TraceStore) Set(key, value []byte) {
	span := t.tracer.StartSpan("set", "key", key, "value", value)
	defer span.End()
	t.store.Set(key, value)
}

func (t *TraceStore) Delete(key []byte) {
	span := t.tracer.StartSpan("delete", "key", key)
	defer span.End()
	t.store.Delete(key)
}

func (t *TraceStore) Iterator(start, end []byte) types.Iterator {
	span := t.tracer.StartSpan("iterate", "start", start, "end", end)
	iter := t.store.Iterator(start, end)
	return &traceIterator{iter: iter, parentSpan: span}
}

func (t *TraceStore) ReverseIterator(start, end []byte) types.Iterator {
	span := t.tracer.StartSpan("iterate", "start", start, "end", end, "reverse", true)
	iter := t.store.ReverseIterator(start, end)
	return &traceIterator{iter: iter, parentSpan: span}
}

type traceIterator struct {
	iter       types.Iterator
	parentSpan telemetry.Span
}

func (t *traceIterator) Domain() (start []byte, end []byte) {
	return t.iter.Domain()
}

func (t *traceIterator) Valid() bool {
	return t.iter.Valid()
}

func (t *traceIterator) Next() {
	span := t.parentSpan.StartSpan("next")
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
	span := t.parentSpan.StartSpan("close")
	defer span.End()
	err := t.iter.Close()
	// close the parent span
	t.parentSpan.End()
	return err
}

var _ types.KVStore = (*TraceStore)(nil)
var _ types.Iterator = (*traceIterator)(nil)
