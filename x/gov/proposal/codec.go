package proposal

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// Registeres types on the codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Content)(nil), nil)
}
