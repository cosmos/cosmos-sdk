package merkle

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(Root{}, "ibc/commitment/merkle/Root", nil)
	cdc.RegisterConcrete(Path{}, "ibc/commitment/merkle/Path", nil)
	cdc.RegisterConcrete(Proof{}, "ibc/commitment/merkle/Proof", nil)
}
