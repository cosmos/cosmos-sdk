package store

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	writeOp     operation = "write"
	readOp      operation = "read"
	deleteOp    operation = "delete"
	iterKeyOp   operation = "iterKey"
	iterValueOp operation = "iterValue"
)

type (
	// TraceKVStore implements the KVStore interface with tracing enabled.
	// Operations are traced on each core KVStore call and written to the
	// underlying io.writer.
	//
	// TODO: Should we use a buffered writer and implement Commit on
	// TraceKVStore?
	TraceKVStore struct {
		parent  sdk.KVStore
		writer  io.Writer
		context TraceContext
	}

	// operation represents an IO operation
	operation string

	// traceOperation implements a traced KVStore operation
	traceOperation struct {
		Operation operation              `json:"operation"`
		Key       string                 `json:"key"`
		Value     string                 `json:"value"`
		Metadata  map[string]interface{} `json:"metadata"`
	}
)

// NewTraceKVStore returns a reference to a new traceKVStore given a parent
// KVStore implementation and a buffered writer.
func NewTraceKVStore(parent sdk.KVStore, writer io.Writer, tc TraceContext) *TraceKVStore {
	return &TraceKVStore{parent: parent, writer: writer, context: tc}
}

// Get implements the KVStore interface. It traces a read operation and
// delegates a Get call to the parent KVStore.
func (tkv *TraceKVStore) Get(key []byte) []byte {
	value := tkv.parent.Get(key)

	writeOperation(tkv.writer, readOp, tkv.context, key, value)
	return value
}

// Set implements the KVStore interface. It traces a write operation and
// delegates the Set call to the parent KVStore.
func (tkv *TraceKVStore) Set(key []byte, value []byte) {
	writeOperation(tkv.writer, writeOp, tkv.context, key, value)
	tkv.parent.Set(key, value)
}

// Delete implements the KVStore interface. It traces a write operation and
// delegates the Delete call to the parent KVStore.
func (tkv *TraceKVStore) Delete(key []byte) {
	writeOperation(tkv.writer, deleteOp, tkv.context, key, nil)
	tkv.parent.Delete(key)
}

// Has implements the KVStore interface. It delegates the Has call to the
// parent KVStore.
func (tkv *TraceKVStore) Has(key []byte) bool {
	return tkv.parent.Has(key)
}

// Prefix implements the KVStore interface.
func (tkv *TraceKVStore) Prefix(prefix []byte) KVStore {
	return prefixStore{tkv, prefix}
}

// Gas implements the KVStore interface.
func (tkv *TraceKVStore) Gas(meter GasMeter, config GasConfig) KVStore {
	return NewGasKVStore(meter, config, tkv.parent)
}

// Iterator implements the KVStore interface. It delegates the Iterator call
// the to the parent KVStore.
func (tkv *TraceKVStore) Iterator(start, end []byte) sdk.Iterator {
	return tkv.iterator(start, end, true)
}

// ReverseIterator implements the KVStore interface. It delegates the
// ReverseIterator call the to the parent KVStore.
func (tkv *TraceKVStore) ReverseIterator(start, end []byte) sdk.Iterator {
	return tkv.iterator(start, end, false)
}

// iterator facilitates iteration over a KVStore. It delegates the necessary
// calls to it's parent KVStore.
func (tkv *TraceKVStore) iterator(start, end []byte, ascending bool) sdk.Iterator {
	var parent sdk.Iterator

	if ascending {
		parent = tkv.parent.Iterator(start, end)
	} else {
		parent = tkv.parent.ReverseIterator(start, end)
	}

	return newTraceIterator(tkv.writer, parent, tkv.context)
}

type traceIterator struct {
	parent  sdk.Iterator
	writer  io.Writer
	context TraceContext
}

func newTraceIterator(w io.Writer, parent sdk.Iterator, tc TraceContext) sdk.Iterator {
	return &traceIterator{writer: w, parent: parent, context: tc}
}

// Domain implements the Iterator interface.
func (ti *traceIterator) Domain() (start []byte, end []byte) {
	return ti.parent.Domain()
}

// Valid implements the Iterator interface.
func (ti *traceIterator) Valid() bool {
	return ti.parent.Valid()
}

// Next implements the Iterator interface.
func (ti *traceIterator) Next() {
	ti.parent.Next()
}

// Key implements the Iterator interface.
func (ti *traceIterator) Key() []byte {
	key := ti.parent.Key()

	writeOperation(ti.writer, iterKeyOp, ti.context, key, nil)
	return key
}

// Value implements the Iterator interface.
func (ti *traceIterator) Value() []byte {
	value := ti.parent.Value()

	writeOperation(ti.writer, iterValueOp, ti.context, nil, value)
	return value
}

// Close implements the Iterator interface.
func (ti *traceIterator) Close() {
	ti.parent.Close()
}

// GetStoreType implements the KVStore interface. It returns the underlying
// KVStore type.
func (tkv *TraceKVStore) GetStoreType() sdk.StoreType {
	return tkv.parent.GetStoreType()
}

// CacheWrap implements the KVStore interface. It panics as a TraceKVStore
// cannot be cache wrapped.
func (tkv *TraceKVStore) CacheWrap() sdk.CacheWrap {
	panic("cannot CacheWrap a TraceKVStore")
}

// CacheWrapWithTrace implements the KVStore interface. It panics as a
// TraceKVStore cannot be cache wrapped.
func (tkv *TraceKVStore) CacheWrapWithTrace(_ io.Writer, _ TraceContext) CacheWrap {
	panic("cannot CacheWrapWithTrace a TraceKVStore")
}

// writeOperation writes a KVStore operation to the underlying io.Writer as
// JSON-encoded data where the key/value pair is base64 encoded.
// nolint: errcheck
func writeOperation(w io.Writer, op operation, tc TraceContext, key, value []byte) {
	traceOp := traceOperation{
		Operation: op,
		Key:       base64.StdEncoding.EncodeToString(key),
		Value:     base64.StdEncoding.EncodeToString(value),
	}

	if tc != nil {
		traceOp.Metadata = tc
	}

	raw, err := json.Marshal(traceOp)
	if err != nil {
		panic(fmt.Sprintf("failed to serialize trace operation: %v", err))
	}

	if _, err := w.Write(raw); err != nil {
		panic(fmt.Sprintf("failed to write trace operation: %v", err))
	}

	io.WriteString(w, "\n")
}
