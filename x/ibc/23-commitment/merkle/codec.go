package merkle

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(Root{}, "ibc/commitment/merkle/Root", nil)
	cdc.RegisterConcrete(Proof{}, "ibc/commitment/merkle/Proof", nil)
}
