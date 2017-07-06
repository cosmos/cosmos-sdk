package state

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/types"
	eyes "github.com/tendermint/merkleeyes/client"
	"github.com/tendermint/tmlibs/log"
)

// CONTRACT: State should be quick to copy.
// See CacheWrap().
type State struct {
	chainID    string
	store      types.KVStore
	readCache  map[string][]byte // optional, for caching writes to store
	writeCache *types.KVCache    // optional, for caching writes w/o writing to store
	logger     log.Logger
}

func NewState(store types.KVStore, l log.Logger) *State {
	return &State{
		chainID:    "",
		store:      store,
		readCache:  make(map[string][]byte),
		writeCache: nil,
		logger:     l,
	}
}

func (s *State) SetChainID(chainID string) {
	s.chainID = chainID
	s.store.Set([]byte("base/chain_id"), []byte(chainID))
}

func (s *State) GetChainID() string {
	if s.chainID != "" {
		return s.chainID
	}
	s.chainID = string(s.store.Get([]byte("base/chain_id")))
	return s.chainID
}

func (s *State) Get(key []byte) (value []byte) {
	if s.readCache != nil { //if not a cachewrap
		value, ok := s.readCache[string(key)]
		if ok {
			return value
		}
	}
	return s.store.Get(key)
}

func (s *State) Set(key []byte, value []byte) {
	if s.readCache != nil { //if not a cachewrap
		s.readCache[string(key)] = value
	}
	s.store.Set(key, value)
}

func (s *State) CacheWrap() *State {
	cache := types.NewKVCache(s)
	return &State{
		chainID:    s.chainID,
		store:      cache,
		readCache:  nil,
		writeCache: cache,
		logger:     s.logger,
	}
}

// NOTE: errors if s is not from CacheWrap()
func (s *State) CacheSync() {
	s.writeCache.Sync()
}

func (s *State) Commit() abci.Result {
	switch s.store.(type) {
	case *eyes.Client:
		s.readCache = make(map[string][]byte)
		return s.store.(*eyes.Client).CommitSync()
	default:
		return abci.NewError(abci.CodeType_InternalError, "can only use Commit if store is merkleeyes")
	}

}
