package state

import (
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
	eyes "github.com/tendermint/merkleeyes/client"
)

type State struct {
	chainID string
	eyesCli *eyes.Client

	LastBlockHeight uint64
	LastBlockHash   []byte
	GasLimit        int64
}

func NewState(chainID string, eyesCli *eyes.Client) *State {
	s := &State{
		chainID: chainID,
		eyesCli: eyesCli,
	}
	return s
}

func (s *State) GetAccount(addr []byte) *types.Account {
	accBytes, err := s.eyesCli.GetSync(addr)
	if err != nil {
		panic("Error loading account: " + err.Error())
	}
	if len(accBytes) == 0 {
		return nil
	}
	var acc types.Account
	err = wire.ReadBinaryBytes(accBytes, &acc)
	if err != nil {
		panic("Error reading account: " + err.Error())
	}
	return &acc
}

func (s *State) SetAccount(acc *types.Account) {
	accBytes := wire.BinaryBytes(acc)
	err := s.eyesCli.SetSync(acc.PubKey.Address(), accBytes)
	if err != nil {
		panic("Error storing account: " + err.Error())
	}
}
