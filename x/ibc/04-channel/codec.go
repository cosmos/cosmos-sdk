package channel

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var msgCdc = codec.New()

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Packet)(nil), nil)
}
