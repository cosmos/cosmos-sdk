package state

import (
	"github.com/tendermint/basecoin/types"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-wire"
	eyes "github.com/tendermint/merkleeyes/client"
)

type State struct {
	chainID    string
	eyesCli    *eyes.Client
	checkCache *types.AccountCache

	LastBlockHeight uint64
	LastBlockHash   []byte
	GasLimit        int64
}

func NewState(eyesCli *eyes.Client) *State {
	s := &State{
		chainID: "",
		eyesCli: eyesCli,
	}
	s.checkCache = types.NewAccountCache(s)
	return s
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

func (s *State) GetAccount(addr []byte) *types.Account {
	res := s.eyesCli.GetSync(AccountKey(addr))
	if res.IsErr() {
		panic(Fmt("Error loading account addr %X error: %v", addr, res.Error()))
	}
	if len(res.Data) == 0 {
		return nil
	}
	var acc *types.Account
	err := wire.ReadBinaryBytes(res.Data, &acc)
	if err != nil {
		panic(Fmt("Error reading account %X error: %v", res.Data, err.Error()))
	}
	return acc
}

func (s *State) SetAccount(addr []byte, acc *types.Account) {
	accBytes := wire.BinaryBytes(acc)
	res := s.eyesCli.SetSync(AccountKey(addr), accBytes)
	if res.IsErr() {
		panic(Fmt("Error storing account addr %X error: %v", addr, res.Error()))
	}
}

func (s *State) GetCheckCache() *types.AccountCache {
	return s.checkCache
}

func (s *State) ResetCacheState() {
	s.checkCache = types.NewAccountCache(s)
}

//----------------------------------------

func AccountKey(addr []byte) []byte {
	return append([]byte("base/a/"), addr...)
}
