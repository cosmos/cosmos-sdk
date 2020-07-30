package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers all the necessary types and interfaces for the
// capability module.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(&Capability{}, "cosmos-sdk/Capability", nil)
	cdc.RegisterConcrete(Owner{}, "cosmos-sdk/Owner", nil)
	cdc.RegisterConcrete(&CapabilityOwners{}, "cosmos-sdk/CapabilityOwners", nil)
}

var (
	amino = codec.New()
)

func init() {
	RegisterCodec(amino)
	amino.Seal()
}
