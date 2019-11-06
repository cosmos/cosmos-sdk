package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgTransfer{}, "ibc/transfer/MsgTransfer", nil)
	cdc.RegisterConcrete(MsgRecvPacket{}, "ibc/transfer/MsgRecvPacket", nil)
	cdc.RegisterConcrete(PacketData{}, "ibc/transfer/PacketData", nil)
}

var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
	channel.RegisterCodec(ModuleCdc)
	commitment.RegisterCodec(ModuleCdc)
}
