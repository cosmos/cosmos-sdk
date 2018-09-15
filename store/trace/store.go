package trace

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
	// Store implements the KVStore interface with tracing enabled.
	// Operations are traced on each core KVStore call and written to the
	// underlying io.writer.
	//
	// TODO: Should we use a buffered writer and implement Commit on
	// Store?
	Store struct {
		parent sdk.KVStore
		tracer *sdk.Tracer
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

// NewStore returns a reference to a new traceKVStore given a parent
// KVStore implementation and a buffered writer.
func NewStore(parent sdk.KVStore, tracer *sdk.Tracer) *Store {
	return &Store{parent: parent, tracer: tracer}
}

// Get implements the KVStore interface. It traces a read operation and
// delegates a Get call to the parent KVStore.
func (tkv *Store) Get(key []byte) []byte {
	value := tkv.parent.Get(key)

	writeOperation(tkv.tracer, readOp, key, value)
	return value
}

// Set implements the KVStore interface. It traces a write operation and
// delegates the Set call to the parent KVStore.
func (tkv *Store) Set(key []byte, value []byte) {
	writeOperation(tkv.tracer, writeOp, key, value)
	tkv.parent.Set(key, value)
}

// Delete implements the KVStore interface. It traces a write operation and
// delegates the Delete call to the parent KVStore.
func (tkv *Store) Delete(key []byte) {
	writeOperation(tkv.tracer, deleteOp, key, nil)
	tkv.parent.Delete(key)
}

// Has implements the KVStore interface. It delegates the Has call to the
// parent KVStore.
func (tkv *Store) Has(key []byte) bool {
	return tkv.parent.Has(key)
}

// Iterator implements the KVStore interface. It delegates the Iterator call
// the to the parent KVStore.
func (tkv *Store) Iterator(start, end []byte) sdk.Iterator {
	return tkv.iterator(start, end, true)
}

// ReverseIterator implements the KVStore interface. It delegates the
// ReverseIterator call the to the parent KVStore.
func (tkv *Store) ReverseIterator(start, end []byte) sdk.Iterator {
	return tkv.iterator(start, end, false)
}

// iterator facilitates iteration over a KVStore. It delegates the necessary
// calls to it's parent KVStore.
func (tkv *Store) iterator(start, end []byte, ascending bool) sdk.Iterator {
	var parent sdk.Iterator

	if ascending {
		parent = tkv.parent.Iterator(start, end)
	} else {
		parent = tkv.parent.ReverseIterator(start, end)
	}

	return newTraceIterator(tkv.tracer, parent)
}

type traceIterator struct {
	parent sdk.Iterator
	tracer *sdk.Tracer
}

func newTraceIterator(tracer *sdk.Tracer, parent sdk.Iterator) sdk.Iterator {
	return &traceIterator{parent: parent, tracer: tracer}
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

	writeOperation(ti.tracer, iterKeyOp, key, nil)
	return key
}

// Value implements the Iterator interface.
func (ti *traceIterator) Value() []byte {
	value := ti.parent.Value()

	writeOperation(ti.tracer, iterValueOp, nil, value)
	return value
}

// Close implements the Iterator interface.
func (ti *traceIterator) Close() {
	ti.parent.Close()
}

// CacheWrap implements the KVStore interface. It panics as a Store
// cannot be cache wrapped.
func (tkv *Store) CacheWrap() sdk.CacheKVStore {
	panic("cannot CacheWrap a Store")
}

// writeOperation writes a KVStore operation to the underlying io.Writer as
// JSON-encoded data where the key/value pair is base64 encoded.
// nolint: errcheck
func writeOperation(tracer *sdk.Tracer, op operation, key, value []byte) {
	traceOp := traceOperation{
		Operation: op,
		Key:       base64.StdEncoding.EncodeToString(key),
		Value:     base64.StdEncoding.EncodeToString(value),
	}

	tc := tracer.Context
	if tc != nil {
		traceOp.Metadata = tc
	}

	raw, err := json.Marshal(traceOp)
	if err != nil {
		panic(fmt.Sprintf("failed to serialize trace operation: %v", err))
	}

	w := tracer.Writer
	if _, err := w.Write(raw); err != nil {
		panic(fmt.Sprintf("failed to write trace operation: %v", err))
	}

	io.WriteString(w, "\n")
}
