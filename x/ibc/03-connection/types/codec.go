package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInterfaces register the ibc interfaces submodule implementations to protobuf
// Any.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgConnectionOpenInit{},
		&MsgConnectionOpenTry{},
		&MsgConnectionOpenAck{},
		&MsgConnectionOpenConfirm{},
	)
}

var (
	// SubModuleCdc references the global x/ibc/03-connectionl module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding.
	//
	// The actual codec used for serialization should be provided to x/ibc/03-connectionl and
	// defined at the application level.
	SubModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
