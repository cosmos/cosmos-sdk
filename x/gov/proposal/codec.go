package proposal

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var internalCdc = codec.New()

// Registers types on the codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Content)(nil), nil)
}
