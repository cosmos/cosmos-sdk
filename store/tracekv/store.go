package tracekv

import (
	"io"

	"cosmossdk.io/store/v2"
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
		Operation string                 `json:"operation"`
		Key       string                 `json:"key"`
		Value     string                 `json:"value"`
		Metadata  map[string]interface{} `json:"metadata"`
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

func (s *Store) Get(key []byte) []byte {
	panic("not implemented!")
}

func (s *Store) Has(key []byte) bool {
	panic("not implemented!")
}

func (s *Store) Set(key, value []byte) {
	panic("not implemented!")
}

func (s *Store) Delete(key []byte) {
	panic("not implemented!")
}

func (s *Store) Reset() error {
	panic("not implemented!")
}

func (s *Store) Branch() store.BranchedKVStore {
	panic("not implemented!")
}

func (s *Store) BranchWithTrace(w io.Writer, tc store.TraceContext) store.BranchedKVStore {
	panic("not implemented!")
}

func (s *Store) GetChangeset() *store.Changeset {
	panic("not implemented!")
}

func (s *Store) Iterator(start, end []byte) store.Iterator {
	panic("not implemented!")
}

func (s *Store) ReverseIterator(start, end []byte) store.Iterator {
	panic("not implemented!")
}
