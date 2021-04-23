package smt

import (
	"crypto/sha256"
	"encoding/binary"
	"io"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/lazyledger/smt"
)

var (
	_ types.KVStore                 = (*Store)(nil)
	_ types.CommitStore             = (*Store)(nil)
	_ types.CommitKVStore           = (*Store)(nil)
	_ types.Queryable               = (*Store)(nil)
	_ types.StoreWithInitialVersion = (*Store)(nil)
)

var (
	prefixLen      = 1
	versionsPrefix = []byte{0}
	indexPrefix    = []byte{1}
	afterIndex     = []byte{2}
)

// Store Implements types.KVStore and CommitKVStore.
type Store struct {
	tree *smt.SparseMerkleTree
	db   dbm.DB

	version int64

	opts struct {
		initialVersion int64
		pruningOptions types.PruningOptions
	}

	mtx sync.RWMutex
}

func NewStore(underlyingDB dbm.DB) *Store {
	return &Store{
		tree: smt.NewSparseMerkleTree(underlyingDB, sha256.New()),
		db:   underlyingDB,
	}
}

// KVStore interface below:

func (s *Store) GetStoreType() types.StoreType {
	return types.StoreTypeSMT
}

// CacheWrap branches a store.
func (s *Store) CacheWrap() types.CacheWrap {
	return cachekv.NewStore(s)
}

// CacheWrapWithTrace branches a store with tracing enabled.
func (s *Store) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return cachekv.NewStore(tracekv.NewStore(s, w, tc))
}

// Get returns nil iff key doesn't exist. Panics on nil key.
func (s *Store) Get(key []byte) []byte {
	defer telemetry.MeasureSince(time.Now(), "store", "smt", "get")
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	val, err := s.tree.Get(key)
	if err != nil {
		panic(err)
	}
	return val
}

// Has checks if a key exists. Panics on nil key.
func (s *Store) Has(key []byte) bool {
	defer telemetry.MeasureSince(time.Now(), "store", "smt", "has")
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	has, err := s.db.Has(indexKey(key))
	return err == nil && has
}

// Set sets the key. Panics on nil key or value.
func (s *Store) Set(key []byte, value []byte) {
	s.mtx.Lock()
	_, err := s.tree.Update(key, value)
	if err != nil {
		panic(err.Error())
	}
	err = s.db.Set(indexKey(key), []byte{})
	if err != nil {
		panic(err.Error())
	}
	s.mtx.Unlock()
}

// Delete deletes the key. Panics on nil key.
func (s *Store) Delete(key []byte) {
	defer telemetry.MeasureSince(time.Now(), "store", "smt", "delete")
	s.mtx.Lock()
	_, _ = s.tree.Delete(key)
	_ = s.db.Delete(indexKey(key))
	s.mtx.Unlock()
}

// Iterator over a domain of keys in ascending order. End is exclusive.
// Start must be less than end, or the Iterator is invalid.
// Iterator must be closed by caller.
// To iterate over entire domain, use store.Iterator(nil, nil)
// CONTRACT: No writes may happen within a domain while an iterator exists over it.
// Exceptionally allowed for cachekv.Store, safe to write in the modules.
func (s *Store) Iterator(start []byte, end []byte) types.Iterator {
	iter, err := newIterator(s, start, end, false)
	if err != nil {
		panic(err.Error())
	}
	return iter
}

// Iterator over a domain of keys in descending order. End is exclusive.
// Start must be less than end, or the Iterator is invalid.
// Iterator must be closed by caller.
// CONTRACT: No writes may happen within a domain while an iterator exists over it.
// Exceptionally allowed for cachekv.Store, safe to write in the modules.
func (s *Store) ReverseIterator(start []byte, end []byte) types.Iterator {
	iter, err := newIterator(s, start, end, true)
	if err != nil {
		panic(err.Error())
	}
	return iter
}

// CommitStore interface below:

func (s *Store) Commit() types.CommitID {
	defer telemetry.MeasureSince(time.Now(), "store", "smt", "commit")
	version := s.version + 1

	if version == 1 && s.opts.initialVersion != 0 {
		version = s.opts.initialVersion
	}

	s.version = version

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(version))
	s.db.Set(append(versionsPrefix, b...), s.tree.Root())

	return s.LastCommitID()
}

func (s *Store) LastCommitID() types.CommitID {
	return types.CommitID{
		Version: s.version,
		Hash:    s.tree.Root(),
	}
}

func (s *Store) SetPruning(p types.PruningOptions) {
	s.opts.pruningOptions = p
}

func (s *Store) GetPruning() types.PruningOptions {
	return s.opts.pruningOptions
}

// Queryable interface below:

func (s *Store) Query(_ abci.RequestQuery) abci.ResponseQuery {
	panic("not implemented")
}

// StoreWithInitialVersion interface below:

// SetInitialVersion sets the initial version of the SMT tree. It is used when
// starting a new chain at an arbitrary height.
func (s *Store) SetInitialVersion(version int64) {
	s.opts.initialVersion = version
}
