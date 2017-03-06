package state

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/types"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-wire"
	eyes "github.com/tendermint/merkleeyes/client"
)

// CONTRACT: State should be quick to copy.
// See CacheWrap().
type State struct {
	chainID    string
	store      types.KVStore
	readCache  map[string][]byte // optional, for caching writes to store
	writeCache *types.KVCache    // optional, for caching writes w/o writing to store
}

func NewState(store types.KVStore) *State {
	return &State{
		chainID:    "",
		store:      store,
		readCache:  make(map[string][]byte),
		writeCache: nil,
	}
}

func (s *State) SetChainID(chainID string) {
	s.chainID = chainID
}

func (s *State) GetChainID() string {
	if s.chainID == "" {
		PanicSanity("Expected to have set SetChainID")
	}
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

func (s *State) GetAccount(addr []byte) *types.Account {
	return GetAccount(s, addr)
}

func (s *State) SetAccount(addr []byte, acc *types.Account) {
	SetAccount(s, addr, acc)
}

func (s *State) CacheWrap() *State {
	cache := types.NewKVCache(s)
	return &State{
		chainID:    s.chainID,
		store:      cache,
		readCache:  nil,
		writeCache: cache,
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

//----------------------------------------

func AccountKey(addr []byte) []byte {
	return append([]byte("base/a/"), addr...)
}

func GetAccount(store types.KVStore, addr []byte) *types.Account {
	data := store.Get(AccountKey(addr))
	if len(data) == 0 {
		return nil
	}
	var acc *types.Account
	err := wire.ReadBinaryBytes(data, &acc)
	if err != nil {
		panic(Fmt("Error reading account %X error: %v",
			data, err.Error()))
	}
	return acc
}

func SetAccount(store types.KVStore, addr []byte, acc *types.Account) {
	accBytes := wire.BinaryBytes(acc)
	store.Set(AccountKey(addr), accBytes)
}
