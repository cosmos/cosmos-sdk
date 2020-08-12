package keeper

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/simulation"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

var _ simulation.ClientUnmarshaler = (*Keeper)(nil)

// MustUnmarshalClientState attempts to decode and return an ClientState object from
// raw encoded bytes. It panics on error.
func (k Keeper) MustUnmarshalClientState(bz []byte) exported.ClientState {
	return types.MustUnmarshalClientState(k.cdc, bz)
}

// MustUnmarshalConsensusState attempts to decode and return an ConsensusState object from
// raw encoded bytes. It panics on error.
func (k Keeper) MustUnmarshalConsensusState(bz []byte) exported.ConsensusState {
	return types.MustUnmarshalConsensusState(k.cdc, bz)
}

// MustMarshalClientState attempts to encode an ClientState object and returns the
// raw encoded bytes. It panics on error.
func (k Keeper) MustMarshalClientState(clientState exported.ClientState) []byte {
	return types.MustMarshalClientState(k.cdc, clientState)
}

// MustMarshalConsensusState attempts to encode an ConsensusState object and returns the
// raw encoded bytes. It panics on error.
func (k Keeper) MustMarshalConsensusState(consensusState exported.ConsensusState) []byte {
	return types.MustMarshalConsensusState(k.cdc, consensusState)
}
