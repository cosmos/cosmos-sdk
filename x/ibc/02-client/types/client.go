package types

import (
	"fmt"

	proto "github.com/gogo/protobuf/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

var (
	_ codectypes.UnpackInterfacesMessage = IdentifiedClientState{}
	_ codectypes.UnpackInterfacesMessage = ConsensusStateWithHeight{}
)

// NewIdentifiedClientState creates a new IdentifiedClientState instance
func NewIdentifiedClientState(clientID string, clientState exported.ClientState) IdentifiedClientState {
	msg, ok := clientState.(proto.Message)
	if !ok {
		panic(fmt.Errorf("cannot proto marshal %T", clientState))
	}

	anyClientState, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		panic(err)
	}

	return IdentifiedClientState{
		ClientId:    clientID,
		ClientState: anyClientState,
	}
}

// UnpackInterfaces implements UnpackInterfacesMesssage.UnpackInterfaces
func (ics IdentifiedClientState) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var clientState exported.ClientState
	err := unpacker.UnpackAny(ics.ClientState, &clientState)
	if err != nil {
		return err
	}
	return nil
}

// NewConsensusStateWithHeight creates a new ConsensusStateWithHeight instance
func NewConsensusStateWithHeight(height Height, consensusState exported.ConsensusState) ConsensusStateWithHeight {
	msg, ok := consensusState.(proto.Message)
	if !ok {
		panic(fmt.Errorf("cannot proto marshal %T", consensusState))
	}

	anyConsensusState, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		panic(err)
	}

	return ConsensusStateWithHeight{
		Height:         height,
		ConsensusState: anyConsensusState,
	}
}

// UnpackInterfaces implements UnpackInterfacesMesssage.UnpackInterfaces
func (ics ConsensusStateWithHeight) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var consensusState exported.ConsensusState
	err := unpacker.UnpackAny(ics.ConsensusState, &consensusState)
	if err != nil {
		return err
	}
	return nil
}
