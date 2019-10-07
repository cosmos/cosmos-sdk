package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(PacketSequence{}, "ibcmock/PacketSequence", nil)
	cdc.RegisterConcrete(MsgSequence{}, "ibcmock/MsgSequence", nil)
}

func RegisterSend(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSequence{}, "ibcmock/MsgSequence", nil)
}

func RegisterRecv(cdc *codec.Codec) {
	cdc.RegisterConcrete(PacketSequence{}, "ibcmock/PacketSequence", nil)
}