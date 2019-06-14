package commitment

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Root)(nil), nil)
	cdc.RegisterInterface((*Proof)(nil), nil)
}
