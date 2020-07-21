package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
)

// RegisterCodec registers the necessary x/ibc/04-channel interfaces and concrete types
// on the provided Amino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.PacketI)(nil), nil)
	cdc.RegisterConcrete(Channel{}, "ibc/channel/Channel", nil)
	cdc.RegisterConcrete(Packet{}, "ibc/channel/Packet", nil)

	cdc.RegisterConcrete(&MsgChannelOpenInit{}, "ibc/channel/MsgChannelOpenInit", nil)
	cdc.RegisterConcrete(&MsgChannelOpenTry{}, "ibc/channel/MsgChannelOpenTry", nil)
	cdc.RegisterConcrete(&MsgChannelOpenAck{}, "ibc/channel/MsgChannelOpenAck", nil)
	cdc.RegisterConcrete(&MsgChannelOpenConfirm{}, "ibc/channel/MsgChannelOpenConfirm", nil)
	cdc.RegisterConcrete(&MsgChannelCloseInit{}, "ibc/channel/MsgChannelCloseInit", nil)
	cdc.RegisterConcrete(&MsgChannelCloseConfirm{}, "ibc/channel/MsgChannelCloseConfirm", nil)
	cdc.RegisterConcrete(&MsgPacket{}, "ibc/channel/MsgPacket", nil)
	cdc.RegisterConcrete(&MsgAcknowledgement{}, "ibc/channel/MsgAcknowledgement", nil)
	cdc.RegisterConcrete(&MsgTimeout{}, "ibc/channel/MsgTimeout", nil)
}

// RegisterInterfaces register the ibc channel submodule interfaces to protobuf
// Any.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgChannelOpenInit{},
		&MsgChannelOpenTry{},
		&MsgChannelOpenAck{},
		&MsgChannelOpenConfirm{},
		&MsgChannelCloseInit{},
		&MsgChannelCloseConfirm{},
		&MsgPacket{},
		&MsgAcknowledgement{},
		&MsgTimeout{},
	)
}

var (
	amino = codec.New()

	// SubModuleCdc references the global x/ibc/04-channel module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/ibc/04-channel and
	// defined at the application level.
	SubModuleCdc = codec.NewHybridCodec(amino, cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(amino)
	amino.Seal()
}
