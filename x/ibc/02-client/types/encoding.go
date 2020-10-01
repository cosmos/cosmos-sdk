package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

// MustUnmarshalClientState attempts to decode and return an ClientState object from
// raw encoded bytes. It panics on error.
func MustUnmarshalClientState(cdc codec.BinaryMarshaler, bz []byte) exported.ClientState {
	clientState, err := UnmarshalClientState(cdc, bz)
	if err != nil {
		panic(fmt.Errorf("failed to decode client state: %w", err))
	}

	return clientState
}

// MustMarshalClientState attempts to encode an ClientState object and returns the
// raw encoded bytes. It panics on error.
func MustMarshalClientState(cdc codec.BinaryMarshaler, clientState exported.ClientState) []byte {
	bz, err := MarshalClientState(cdc, clientState)
	if err != nil {
		panic(fmt.Errorf("failed to encode client state: %w", err))
	}

	return bz
}

// MarshalClientState marshals an ClientState interface. If the given type implements
// the Marshaler interface, it is treated as a Proto-defined message and
// serialized that way.
func MarshalClientState(cdc codec.BinaryMarshaler, clientStateI exported.ClientState) ([]byte, error) {
	return codec.MarshalAny(cdc, clientStateI)
}

// UnmarshalClientState returns an ClientState interface from raw encoded clientState
// bytes of a Proto-based ClientState type. An error is returned upon decoding
// failure.
func UnmarshalClientState(cdc codec.BinaryMarshaler, bz []byte) (exported.ClientState, error) {
	var clientState exported.ClientState
	if err := codec.UnmarshalAny(cdc, &clientState, bz); err != nil {
		return nil, err
	}

	return clientState, nil
}

// MustUnmarshalConsensusState attempts to decode and return an ConsensusState object from
// raw encoded bytes. It panics on error.
func MustUnmarshalConsensusState(cdc codec.BinaryMarshaler, bz []byte) exported.ConsensusState {
	consensusState, err := UnmarshalConsensusState(cdc, bz)
	if err != nil {
		panic(fmt.Errorf("failed to decode consensus state: %w", err))
	}

	return consensusState
}

// MustMarshalConsensusState attempts to encode an ConsensusState object and returns the
// raw encoded bytes. It panics on error.
func MustMarshalConsensusState(cdc codec.BinaryMarshaler, consensusState exported.ConsensusState) []byte {
	bz, err := MarshalConsensusState(cdc, consensusState)
	if err != nil {
		panic(fmt.Errorf("failed to encode consensus state: %w", err))
	}

	return bz
}

// MarshalConsensusState marshals an ConsensusState interface. If the given type implements
// the Marshaler interface, it is treated as a Proto-defined message and
// serialized that way.
func MarshalConsensusState(cdc codec.BinaryMarshaler, consensusStateI exported.ConsensusState) ([]byte, error) {
	return codec.MarshalAny(cdc, consensusStateI)
}

// UnmarshalConsensusState returns an ConsensusState interface from raw encoded clientState
// bytes of a Proto-based ConsensusState type. An error is returned upon decoding
// failure.
func UnmarshalConsensusState(cdc codec.BinaryMarshaler, bz []byte) (exported.ConsensusState, error) {
	var consensusState exported.ConsensusState
	if err := codec.UnmarshalAny(cdc, &consensusState, bz); err != nil {
		return nil, err
	}

	return consensusState, nil
}
