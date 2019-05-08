package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// module codec
var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
}

// RegisterCodec registers all necessary param module types with a given codec.
func RegisterCodec(ModuleCdc *codec.Codec) {
	cdc.RegisterConcrete(ParameterChangeProposal{}, "cosmos-sdk/ParameterChangeProposal", nil)
}
