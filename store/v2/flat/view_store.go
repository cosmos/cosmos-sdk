package flat

import (
	"errors"
	"io"

	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/prefix"

	util "github.com/cosmos/cosmos-sdk/internal"
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/listenkv"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	types "github.com/cosmos/cosmos-sdk/store/v2"
	"github.com/cosmos/cosmos-sdk/store/v2/smt"
)

var ErrReadOnly = errors.New("cannot modify read-only store")

// Represents a read-only view of a store's contents at a given version.
type viewStore struct {
	stateView   dbm.DBReader
	dataBucket  dbm.DBReader
	indexBucket dbm.DBReader
	merkleView  dbm.DBReader
	merkleStore *smt.Store
}

func (s *Store) GetVersion(ver int64) (ret *viewStore, err error) {
	stateView, err := s.stateDB.ReaderAt(uint64(ver))
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			err = util.CombineErrors(err, stateView.Discard(), "stateView.Discard also failed")
		}
	}()

	merkleView := stateView
	if s.opts.MerkleDB != nil {
		merkleView, err = s.opts.MerkleDB.ReaderAt(uint64(ver))
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				err = util.CombineErrors(err, merkleView.Discard(), "merkleView.Discard also failed")
			}
		}()
	}
	root, err := stateView.Get(merkleRootKey)
	if err != nil {
		return
	}
	return &viewStore{
		stateView:   stateView,
		dataBucket:  prefix.NewPrefixReader(stateView, dataPrefix),
		merkleView:  merkleView,
		indexBucket: prefix.NewPrefixReader(stateView, indexPrefix),
		merkleStore: loadSMT(dbm.ReaderAsReadWriter(merkleView), root),
	}, nil
}

// Get implements KVStore.
func (s *viewStore) Get(key []byte) []byte {
	val, err := s.dataBucket.Get(key)
	if err != nil {
		panic(err)
	}
	return val
}

// Has implements KVStore.
func (s *viewStore) Has(key []byte) bool {
	has, err := s.dataBucket.Has(key)
	if err != nil {
		panic(err)
	}
	return has
}

// Set implements KVStore.
func (s *viewStore) Set(key []byte, value []byte) {
	panic(ErrReadOnly)
}

// Delete implements KVStore.
func (s *viewStore) Delete(key []byte) {
	panic(ErrReadOnly)
}

// Iterator implements KVStore.
func (s *viewStore) Iterator(start, end []byte) types.Iterator {
	iter, err := s.dataBucket.Iterator(start, end)
	if err != nil {
		panic(err)
	}
	return newIterator(iter)
}

// ReverseIterator implements KVStore.
func (s *viewStore) ReverseIterator(start, end []byte) types.Iterator {
	iter, err := s.dataBucket.ReverseIterator(start, end)
	if err != nil {
		panic(err)
	}
	return newIterator(iter)
}

// GetStoreType implements Store.
func (s *viewStore) GetStoreType() types.StoreType {
	return types.StoreTypeDecoupled
}

func (st *viewStore) CacheWrap() types.CacheWrap {
	return cachekv.NewStore(st)
}

func (st *viewStore) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return cachekv.NewStore(tracekv.NewStore(st, w, tc))
}

func (st *viewStore) CacheWrapWithListeners(storeKey types.StoreKey, listeners []types.WriteListener) types.CacheWrap {
	return cachekv.NewStore(listenkv.NewStore(st, storeKey, listeners))
}
