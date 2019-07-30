package commitment

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registeres types declared in this package
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Root)(nil), nil)
	cdc.RegisterInterface((*Path)(nil), nil)
	cdc.RegisterInterface((*Proof)(nil), nil)
}
