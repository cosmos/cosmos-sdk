package types

import (
	"fmt"

	proto "github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

// RegisterInterfaces registers the client interfaces to protobuf Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterInterface(
		"cosmos_sdk.ibc.v1.client.ClientState",
		(*exported.ClientState)(nil),
	)
	registry.RegisterInterface(
		"cosmos_sdk.ibc.v1.client.ConsensusState",
		(*exported.ConsensusState)(nil),
	)
	registry.RegisterInterface(
		"cosmos_sdk.ibc.v1.client.Header",
		(*exported.Header)(nil),
	)
	registry.RegisterInterface(
		"cosmos_sdk.ibc.v1.client.Height",
		(*exported.Height)(nil),
	)
	registry.RegisterImplementations(
		(*exported.Height)(nil),
		&Height{},
	)
}

var (
	// SubModuleCdc references the global x/ibc/02-client module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding.
	//
	// The actual codec used for serialization should be provided to x/ibc/02-client and
	// defined at the application level.
	SubModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
)

// PackClientState constructs a new Any packed with the given client state value. It returns
// an error if the client state can't be casted to a protobuf message or if the concrete
// implemention is not registered to the protobuf codec.
func PackClientState(clientState exported.ClientState) (*codectypes.Any, error) {
	msg, ok := clientState.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("cannot proto marshal %T", clientState)
	}

	anyClientState, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return anyClientState, nil
}

// MustPackClientState calls PackClientState and panics on error.
func MustPackClientState(clientState exported.ClientState) *codectypes.Any {
	anyClientState, err := PackClientState(clientState)
	if err != nil {
		panic(err)
	}

	return anyClientState
}

// UnpackClientState unpacks an Any into a ClientState. It returns an error if the
// client state can't be unpacked into a ClientState.
func UnpackClientState(any *codectypes.Any) (exported.ClientState, error) {
	clientState, ok := any.GetCachedValue().(exported.ClientState)
	if !ok {
		return nil, fmt.Errorf("cannot unpack Any into ClientState %T", any)
	}

	return clientState, nil
}

// PackConsensusState constructs a new Any packed with the given consensus state value. It returns
// an error if the consensus state can't be casted to a protobuf message or if the concrete
// implemention is not registered to the protobuf codec.
func PackConsensusState(consensusState exported.ConsensusState) (*codectypes.Any, error) {
	msg, ok := consensusState.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("cannot proto marshal %T", consensusState)
	}

	anyConsensusState, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return anyConsensusState, nil
}

// MustPackConsensusState calls PackConsensusState and panics on error.
func MustPackConsensusState(consensusState exported.ConsensusState) *codectypes.Any {
	anyConsensusState, err := PackConsensusState(consensusState)
	if err != nil {
		panic(err)
	}

	return anyConsensusState
}

// UnpackConsensusState unpacks an Any into a ConsensusState. It returns an error if the
// consensus state can't be unpacked into a ConsensusState.
func UnpackConsensusState(any *codectypes.Any) (exported.ConsensusState, error) {
	consensusState, ok := any.GetCachedValue().(exported.ConsensusState)
	if !ok {
		return nil, fmt.Errorf("cannot unpack Any into ConsensusState %T", any)
	}

	return consensusState, nil
}
