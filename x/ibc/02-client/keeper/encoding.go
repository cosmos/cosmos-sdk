package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

// MustUnmarshalClientState attempts to decode and return an ClientState object from
// raw encoded bytes. It panics on error.
func (k Keeper) MustUnmarshalClientState(bz []byte) exported.ClientState {
	clientState, err := k.UnmarshalClientState(bz)
	if err != nil {
		panic(fmt.Errorf("failed to decode client state: %w", err))
	}

	return clientState
}

// MustMarshalClientState attempts to encode an ClientState object and returns the
// raw encoded bytes. It panics on error.
func (k Keeper) MustMarshalClientState(clientState exported.ClientState) []byte {
	bz, err := k.MarshalClientState(clientState)
	if err != nil {
		panic(fmt.Errorf("failed to encode client state: %w", err))
	}

	return bz
}

// MarshalClientState marshals an ClientState interface. If the given type implements
// the Marshaler interface, it is treated as a Proto-defined message and
// serialized that way.
func (k Keeper) MarshalClientState(clientStateI exported.ClientState) ([]byte, error) {
	return codec.MarshalAny(k.cdc, clientStateI)
}

// UnmarshalClientState returns an ClientState interface from raw encoded clientState
// bytes of a Proto-based ClientState type. An error is returned upon decoding
// failure.
func (k Keeper) UnmarshalClientState(bz []byte) (exported.ClientState, error) {
	var clientState exported.ClientState
	if err := codec.UnmarshalAny(k.cdc, &clientState, bz); err != nil {
		return nil, err
	}

	return clientState, nil
}

// MustUnmarshalConsensusState attempts to decode and return an ConsensusState object from
// raw encoded bytes. It panics on error.
func (k Keeper) MustUnmarshalConsensusState(bz []byte) exported.ConsensusState {
	consensusState, err := k.UnmarshalConsensusState(bz)
	if err != nil {
		panic(fmt.Errorf("failed to decode consensus state: %w", err))
	}

	return consensusState
}

// MustMarshalConsensusState attempts to encode an ConsensusState object and returns the
// raw encoded bytes. It panics on error.
func (k Keeper) MustMarshalConsensusState(consensusState exported.ConsensusState) []byte {
	bz, err := k.MarshalConsensusState(consensusState)
	if err != nil {
		panic(fmt.Errorf("failed to encode consensus state: %w", err))
	}

	return bz
}

// MarshalConsensusState marshals an ConsensusState interface. If the given type implements
// the Marshaler interface, it is treated as a Proto-defined message and
// serialized that way.
func (k Keeper) MarshalConsensusState(consensusStateI exported.ConsensusState) ([]byte, error) {
	return codec.MarshalAny(k.cdc, consensusStateI)
}

// UnmarshalConsensusState returns an ConsensusState interface from raw encoded clientState
// bytes of a Proto-based ConsensusState type. An error is returned upon decoding
// failure.
func (k Keeper) UnmarshalConsensusState(bz []byte) (exported.ConsensusState, error) {
	var consensusState exported.ConsensusState
	if err := codec.UnmarshalAny(k.cdc, &consensusState, bz); err != nil {
		return nil, err
	}

	return consensusState, nil
}
