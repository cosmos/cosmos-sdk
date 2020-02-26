package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
)

// RegisterCodec registers the necessary x/ibc/04-channel interfaces and concrete types
// on the provided Amino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.PacketI)(nil), nil)
	cdc.RegisterInterface((*exported.PacketDataI)(nil), nil)
	cdc.RegisterInterface((*exported.PacketAcknowledgementI)(nil), nil)
	cdc.RegisterConcrete(Channel{}, "ibc/channel/Channel", nil)
	cdc.RegisterConcrete(Packet{}, "ibc/channel/Packet", nil)

	cdc.RegisterConcrete(MsgChannelOpenInit{}, "ibc/channel/MsgChannelOpenInit", nil)
	cdc.RegisterConcrete(MsgChannelOpenTry{}, "ibc/channel/MsgChannelOpenTry", nil)
	cdc.RegisterConcrete(MsgChannelOpenAck{}, "ibc/channel/MsgChannelOpenAck", nil)
	cdc.RegisterConcrete(MsgChannelOpenConfirm{}, "ibc/channel/MsgChannelOpenConfirm", nil)
	cdc.RegisterConcrete(MsgChannelCloseInit{}, "ibc/channel/MsgChannelCloseInit", nil)
	cdc.RegisterConcrete(MsgChannelCloseConfirm{}, "ibc/channel/MsgChannelCloseConfirm", nil)
}

var (
	amino = codec.New()

	// SubModuleCdc references the global x/ibc/04-channel module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/ibc/04-channel and
	// defined at the application level.
	SubModuleCdc = codec.NewHybridCodec(amino)
)

func init() {
	RegisterCodec(amino)
	amino.Seal()
}
