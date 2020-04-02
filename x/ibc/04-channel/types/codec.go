package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

// SubModuleCdc defines the IBC channel codec.
var SubModuleCdc *codec.Codec

func init() {
	SubModuleCdc = codec.New()
	commitmenttypes.RegisterCodec(SubModuleCdc)
	client.RegisterCodec(SubModuleCdc)
	RegisterCodec(SubModuleCdc)
}

// RegisterCodec registers all the necessary types and interfaces for the
// IBC channel.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.PacketI)(nil), nil)
	cdc.RegisterConcrete(Channel{}, "ibc/channel/Channel", nil)
	cdc.RegisterConcrete(Packet{}, "ibc/channel/Packet", nil)

	cdc.RegisterConcrete(MsgChannelOpenInit{}, "ibc/channel/MsgChannelOpenInit", nil)
	cdc.RegisterConcrete(MsgChannelOpenTry{}, "ibc/channel/MsgChannelOpenTry", nil)
	cdc.RegisterConcrete(MsgChannelOpenAck{}, "ibc/channel/MsgChannelOpenAck", nil)
	cdc.RegisterConcrete(MsgChannelOpenConfirm{}, "ibc/channel/MsgChannelOpenConfirm", nil)
	cdc.RegisterConcrete(MsgChannelCloseInit{}, "ibc/channel/MsgChannelCloseInit", nil)
	cdc.RegisterConcrete(MsgChannelCloseConfirm{}, "ibc/channel/MsgChannelCloseConfirm", nil)

	cdc.RegisterConcrete(MsgPacket{}, "ibc/channel/MsgPacket", nil)

	SetSubModuleCodec(cdc)
}

// SetSubModuleCodec sets the ibc channel codec
func SetSubModuleCodec(cdc *codec.Codec) {
	SubModuleCdc = cdc
}
