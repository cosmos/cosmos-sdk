package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
)

// RegisterInterfaces register the ibc interfaces submodule implementations to protobuf
// Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterInterface(
		"cosmos_sdk.ibc.v1.connection.ConnectionI",
		(*exported.ConnectionI)(nil),
	)
	registry.RegisterInterface(
		"cosmos_sdk.ibc.v1.connection.CounterpartyI",
		(*exported.CounterpartyI)(nil),
	)
	registry.RegisterImplementations(
		(*exported.ConnectionI)(nil),
		&ConnectionEnd{},
	)
	registry.RegisterImplementations(
		(*exported.CounterpartyI)(nil),
		&Counterparty{},
	)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgConnectionOpenInit{},
		&MsgConnectionOpenTry{},
		&MsgConnectionOpenAck{},
		&MsgConnectionOpenConfirm{},
	)
}

var (
	// SubModuleCdc references the global x/ibc/03-connection module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding.
	//
	// The actual codec used for serialization should be provided to x/ibc/03-connectionl and
	// defined at the application level.
	SubModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
)
