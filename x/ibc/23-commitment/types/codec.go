package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

// RegisterCodec registers types declared in this package
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.RootI)(nil), nil)
	cdc.RegisterInterface((*exported.PrefixI)(nil), nil)
	cdc.RegisterInterface((*exported.ProofI)(nil), nil)

	cdc.RegisterConcrete(Root{}, "ibc/commitment/merkle/Root", nil)
	cdc.RegisterConcrete(Prefix{}, "ibc/commitment/merkle/Prefix", nil)
	cdc.RegisterConcrete(Proof{}, "ibc/commitment/merkle/Proof", nil)
}
