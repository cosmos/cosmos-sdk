package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

func GetConsensusStateFromAny(any *codectypes.Any) exported.ConsensusState {
	consensusState, ok := any.GetCachedValue().(exported.ConsensusState)
	if !ok {
		panic("Any is invalid consensus state")
	}

	return consensusState
}
