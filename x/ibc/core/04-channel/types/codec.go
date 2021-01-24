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

	// ModuleCdc references the global x/ibc channel codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/ibc channel and
	// defined at the application level.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}

// RegisterLegacyAminoCodec registers the necessary x/ibc channel interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgChannelOpenInit{}, "cosmos-sdk/MsgChannelOpenInit", nil)
	cdc.RegisterConcrete(&MsgChannelOpenTry{}, "cosmos-sdk/MsgChannelOpenTry", nil)
	cdc.RegisterConcrete(&MsgChannelOpenAck{}, "cosmos-sdk/MsgChannelOpenAck", nil)
	cdc.RegisterConcrete(&MsgChannelOpenConfirm{}, "cosmos-sdk/MsgChannelOpenConfirm", nil)
	cdc.RegisterConcrete(&MsgChannelCloseInit{}, "cosmos-sdk/MsgChannelCloseInit", nil)
	cdc.RegisterConcrete(&MsgChannelCloseConfirm{}, "cosmos-sdk/MsgChannelCloseConfirm", nil)
	cdc.RegisterConcrete(&MsgRecvPacket{}, "cosmos-sdk/MsgRecvPacket", nil)
	cdc.RegisterConcrete(&MsgAcknowledgement{}, "cosmos-sdk/MsgAcknowledgement", nil)
	cdc.RegisterConcrete(&MsgTimeout{}, "cosmos-sdk/MsgTimeout", nil)
	cdc.RegisterConcrete(&MsgTimeoutOnClose{}, "cosmos-sdk/MsgTimeoutOnClose", nil)
}

// RegisterInterfaces register the ibc channel submodule interfaces to protobuf
// Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterInterface(
		"ibc.core.channel.v1.ChannelI",
		(*exported.ChannelI)(nil),
	)
	registry.RegisterInterface(
		"ibc.core.channel.v1.CounterpartyChannelI",
		(*exported.CounterpartyChannelI)(nil),
	)
	registry.RegisterInterface(
		"ibc.core.channel.v1.PacketI",
		(*exported.PacketI)(nil),
	)
	registry.RegisterImplementations(
		(*exported.ChannelI)(nil),
		&Channel{},
	)
	registry.RegisterImplementations(
		(*exported.CounterpartyChannelI)(nil),
		&Counterparty{},
	)
	registry.RegisterImplementations(
		(*exported.PacketI)(nil),
		&Packet{},
	)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgChannelOpenInit{},
		&MsgChannelOpenTry{},
		&MsgChannelOpenAck{},
		&MsgChannelOpenConfirm{},
		&MsgChannelCloseInit{},
		&MsgChannelCloseConfirm{},
		&MsgRecvPacket{},
		&MsgAcknowledgement{},
		&MsgTimeout{},
		&MsgTimeoutOnClose{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// SubModuleCdc references the global x/ibc/core/04-channel module codec. Note, the codec should
// ONLY be used in certain instances of tests and for JSON encoding.
//
// The actual codec used for serialization should be provided to x/ibc/core/04-channel and
// defined at the application level.
var SubModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
