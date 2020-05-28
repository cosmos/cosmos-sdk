package keeper

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
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
// serialized that way. Otherwise, it falls back on the internal Amino codec.
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

// MarshalClientStateJSON JSON encodes an clientState object implementing the ClientState
// interface.
func (k Keeper) MarshalClientStateJSON(clientState exported.ClientState) ([]byte, error) {
	msg, ok := clientState.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("cannot proto marshal %T", clientState)
	}

	any, err := cdctypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return k.cdc.MarshalJSON(any)
}

// UnmarshalClientStateJSON returns an ClientState from JSON encoded bytes
func (k Keeper) UnmarshalClientStateJSON(bz []byte) (exported.ClientState, error) {
	var any cdctypes.Any
	if err := k.cdc.UnmarshalJSON(bz, &any); err != nil {
		return nil, err
	}

	var clientState exported.ClientState
	if err := k.cdc.UnpackAny(&any, &clientState); err != nil {
		return nil, err
	}

	return clientState, nil
}

// MustUnmarshalConsensusState attempts to decode and return an ConsensusState object from
// raw encoded bytes. It panics on error.
func (k Keeper) MustUnmarshalConsensusState(bz []byte) exported.ConsensusState {
	consensusState, err := k.UnmarshalConsensusState(bz)
	if err != nil {
		panic(fmt.Errorf("failed to decode client state: %w", err))
	}

	return consensusState
}

// MustMarshalConsensusState attempts to encode an ConsensusState object and returns the
// raw encoded bytes. It panics on error.
func (k Keeper) MustMarshalConsensusState(clientState exported.ConsensusState) []byte {
	bz, err := k.MarshalConsensusState(clientState)
	if err != nil {
		panic(fmt.Errorf("failed to encode client state: %w", err))
	}

	return bz
}

// MarshalConsensusState marshals an ConsensusState interface. If the given type implements
// the Marshaler interface, it is treated as a Proto-defined message and
// serialized that way. Otherwise, it falls back on the internal Amino codec.
func (k Keeper) MarshalConsensusState(clientStateI exported.ConsensusState) ([]byte, error) {
	return codec.MarshalAny(k.cdc, clientStateI)
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

// MarshalConsensusStateJSON JSON encodes an clientState object implementing the ConsensusState
// interface.
func (k Keeper) MarshalConsensusStateJSON(consensusState exported.ConsensusState) ([]byte, error) {
	msg, ok := consensusState.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("cannot proto marshal %T", consensusState)
	}

	any, err := cdctypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return k.cdc.MarshalJSON(any)
}

// UnmarshalConsensusStateJSON returns an ConsensusState from JSON encoded bytes
func (k Keeper) UnmarshalConsensusStateJSON(bz []byte) (exported.ConsensusState, error) {
	var any cdctypes.Any
	if err := k.cdc.UnmarshalJSON(bz, &any); err != nil {
		return nil, err
	}

	var consensusState exported.ConsensusState
	if err := k.cdc.UnpackAny(&any, &consensusState); err != nil {
		return nil, err
	}

	return consensusState, nil
}
