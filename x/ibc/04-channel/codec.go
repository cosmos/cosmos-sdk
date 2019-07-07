package channel

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Packet)(nil), nil)
}
