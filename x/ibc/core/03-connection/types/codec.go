package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/ibc connection module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/staking and
	// defined at the application level.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}

// RegisterLegacyAminoCodec registers the necessary x/ibc connection interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgConnectionOpenInit{}, "cosmos-sdk/MsgConnectionOpenInit", nil)
	cdc.RegisterConcrete(&MsgConnectionOpenTry{}, "cosmos-sdk/MsgConnectionOpenTry", nil)
	cdc.RegisterConcrete(&MsgConnectionOpenAck{}, "cosmos-sdk/MsgConnectionOpenAck", nil)
	cdc.RegisterConcrete(&MsgConnectionOpenConfirm{}, "cosmos-sdk/MsgConnectionOpenConfirm", nil)
}

// RegisterInterfaces register the ibc interfaces submodule implementations to protobuf
// Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterInterface(
		"ibc.core.connection.v1.ConnectionI",
		(*exported.ConnectionI)(nil),
		&ConnectionEnd{},
	)
	registry.RegisterInterface(
		"ibc.core.connection.v1.CounterpartyConnectionI",
		(*exported.CounterpartyConnectionI)(nil),
		&Counterparty{},
	)
	registry.RegisterInterface(
		"ibc.core.connection.v1.Version",
		(*exported.Version)(nil),
		&Version{},
	)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgConnectionOpenInit{},
		&MsgConnectionOpenTry{},
		&MsgConnectionOpenAck{},
		&MsgConnectionOpenConfirm{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	// SubModuleCdc references the global x/ibc/core/03-connection module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding.
	//
	// The actual codec used for serialization should be provided to x/ibc/core/03-connection and
	// defined at the application level.
	SubModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
)
