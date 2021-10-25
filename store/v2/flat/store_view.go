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
type storeView struct {
	stateView   dbm.DBReader
	dataBucket  dbm.DBReader
	indexBucket dbm.DBReader
	merkleView  dbm.DBReader
	merkleStore *smt.Store
}

func (s *Store) GetVersion(version int64) (ret *storeView, err error) {
	stateView, err := s.stateDB.ReaderAt(uint64(version))
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
		merkleView, err = s.opts.MerkleDB.ReaderAt(uint64(version))
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
	return &storeView{
		stateView:   stateView,
		dataBucket:  prefix.NewPrefixReader(stateView, dataPrefix),
		indexBucket: prefix.NewPrefixReader(stateView, indexPrefix),
		merkleView:  merkleView,
		merkleStore: loadSMT(dbm.ReaderAsReadWriter(merkleView), root),
	}, nil
}

func (s *storeView) GetMerkleStore() *smt.Store {
	return s.merkleStore
}

// Get implements KVStore.
func (s *storeView) Get(key []byte) []byte {
	val, err := s.dataBucket.Get(key)
	if err != nil {
		panic(err)
	}
	return val
}

// Has implements KVStore.
func (s *storeView) Has(key []byte) bool {
	has, err := s.dataBucket.Has(key)
	if err != nil {
		panic(err)
	}
	return has
}

// Set implements KVStore.
func (s *storeView) Set(key []byte, value []byte) {
	panic(ErrReadOnly)
}

// Delete implements KVStore.
func (s *storeView) Delete(key []byte) {
	panic(ErrReadOnly)
}

// Iterator implements KVStore.
func (s *storeView) Iterator(start, end []byte) types.Iterator {
	iter, err := s.dataBucket.Iterator(start, end)
	if err != nil {
		panic(err)
	}
	return newIterator(iter)
}

// ReverseIterator implements KVStore.
func (s *storeView) ReverseIterator(start, end []byte) types.Iterator {
	iter, err := s.dataBucket.ReverseIterator(start, end)
	if err != nil {
		panic(err)
	}
	return newIterator(iter)
}

// GetStoreType implements Store.
func (s *storeView) GetStoreType() types.StoreType {
	return types.StoreTypeDecoupled
}

func (st *storeView) CacheWrap() types.CacheWrap {
	return cachekv.NewStore(st)
}

func (st *storeView) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return cachekv.NewStore(tracekv.NewStore(st, w, tc))
}

func (st *storeView) CacheWrapWithListeners(storeKey types.StoreKey, listeners []types.WriteListener) types.CacheWrap {
	return cachekv.NewStore(listenkv.NewStore(st, storeKey, listeners))
}
