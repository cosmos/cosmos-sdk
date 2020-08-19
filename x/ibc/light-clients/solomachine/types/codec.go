package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

// RegisterInterfaces register the ibc channel submodule interfaces to protobuf
// Any.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateClient{},
		&MsgUpdateClient{},
		&MsgSubmitClientMisbehaviour{},
	)
	registry.RegisterImplementations(
		(*clientexported.ClientState)(nil),
		&ClientState{},
	)
	registry.RegisterImplementations(
		(*clientexported.ConsensusState)(nil),
		&ConsensusState{},
	)
}

var (
	// SubModuleCdc references the global x/ibc/light-clients/solomachine module codec. Note, the codec
	// should ONLY be used in certain instances of tests and for JSON encoding..
	//
	// The actual codec used for serialization should be provided to x/ibc/light-clients/solomachine and
	// defined at the application level.
	SubModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
