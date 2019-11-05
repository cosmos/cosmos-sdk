package commitment

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers types declared in this package
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*RootI)(nil), nil)
	cdc.RegisterInterface((*PrefixI)(nil), nil)
	cdc.RegisterInterface((*PathI)(nil), nil)
	cdc.RegisterInterface((*ProofI)(nil), nil)

	cdc.RegisterConcrete(Root{}, "ibc/commitment/merkle/Root", nil)
	cdc.RegisterConcrete(Prefix{}, "ibc/commitment/merkle/Prefix", nil)
	cdc.RegisterConcrete(Path{}, "ibc/commitment/merkle/Path", nil)
	cdc.RegisterConcrete(Proof{}, "ibc/commitment/merkle/Proof", nil)
}
