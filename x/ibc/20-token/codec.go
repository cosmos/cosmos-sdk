package token

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var moduleCdc *codec.Codec

func init() {
	moduleCdc = codec.New()
	RegisterCodec(moduleCdc)
}

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSend{}, "ibc/token/MsgSend", nil)
	cdc.RegisterConcrete(PacketSend{}, "ibc/token/PacketSend", nil)
}
