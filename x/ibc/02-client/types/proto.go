package types

import (
	"fmt"

	proto "github.com/gogo/protobuf/proto"

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

func GetAnyFromConsensusState(consensusState exported.ConsensusState) *codectypes.Any {
	msg, ok := consensusState.(proto.Message)
	if !ok {
		panic(fmt.Errorf("cannot proto marshal %T", consensusState))
	}

	anyConsensusState, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		panic(err)
	}

	return anyConsensusState
}
