package tracekv

import (
	"encoding/base64"
	"encoding/json"
	"io"

	"cosmossdk.io/store/v2"
	"github.com/cockroachdb/errors"
)

const (
	writeOp     = "write"
	readOp      = "read"
	deleteOp    = "delete"
	iterKeyOp   = "iterKey"
	iterValueOp = "iterValue"
)

var _ store.KVStore = (*Store)(nil)

type (
	// Store defines a KVStore used for tracing capabilities, which typically wraps
	// another KVStore implementation.
	Store struct {
		parent  store.KVStore
		context store.TraceContext
		writer  io.Writer
	}

	// traceOperation defines a traced KVStore operation, such as a read or write
	traceOperation struct {
		Operation string         `json:"operation"`
		Key       string         `json:"key"`
		Value     string         `json:"value"`
		Metadata  map[string]any `json:"metadata"`
	}
)

func New(p store.KVStore, w io.Writer, tc store.TraceContext) store.KVStore {
	return &Store{
		parent:  p,
		writer:  w,
		context: tc,
	}
}

func (s *Store) GetStoreKey() string {
	return s.parent.GetStoreKey()
}

func (s *Store) GetStoreType() store.StoreType {
	return store.StoreTypeTrace
}

func (s *Store) GetChangeset() *store.Changeset {
	return s.parent.GetChangeset()
}

func (s *Store) Get(key []byte) []byte {
	value := s.parent.Get(key)
	writeOperation(s.writer, readOp, s.context, key, value)
	return value
}

func (s *Store) Has(key []byte) bool {
	return s.parent.Has(key)
}

func (s *Store) Set(key, value []byte) {
	writeOperation(s.writer, writeOp, s.context, key, value)
	s.parent.Set(key, value)
}

func (s *Store) Delete(key []byte) {
	writeOperation(s.writer, deleteOp, s.context, key, nil)
	s.parent.Delete(key)
}

func (s *Store) Reset() error {
	return s.parent.Reset()
}

func (s *Store) Iterator(start, end []byte) store.Iterator {
	return newIterator(s.writer, s.parent.Iterator(start, end), s.context)
}

func (s *Store) ReverseIterator(start, end []byte) store.Iterator {
	return newIterator(s.writer, s.parent.ReverseIterator(start, end), s.context)
}

// writeOperation writes a KVStore operation to the underlying io.Writer as
// JSON-encoded data where the key/value pair is base64 encoded.
func writeOperation(w io.Writer, op string, tc store.TraceContext, key, value []byte) {
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
		panic(errors.Wrap(err, "failed to serialize trace operation"))
	}

	if _, err := w.Write(raw); err != nil {
		panic(errors.Wrap(err, "failed to write trace operation"))
	}

	_, err = io.WriteString(w, "\n")
	if err != nil {
		panic(err)
	}
}
