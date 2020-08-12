package types

import (
	"fmt"

	proto "github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

// RegisterCodec registers the IBC client interfaces and types
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*exported.ClientState)(nil), nil) // remove after genesis migration
	cdc.RegisterInterface((*exported.MsgCreateClient)(nil), nil)
	cdc.RegisterInterface((*exported.MsgUpdateClient)(nil), nil)
	cdc.RegisterInterface((*exported.ConsensusState)(nil), nil)
	cdc.RegisterInterface((*exported.Header)(nil), nil)
	cdc.RegisterInterface((*exported.Misbehaviour)(nil), nil)
}

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
}

var (
	amino = codec.New()

	// SubModuleCdc references the global x/ibc/02-client module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/ibc/02-client and
	// defined at the application level.
	SubModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(amino)
	amino.Seal()
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
