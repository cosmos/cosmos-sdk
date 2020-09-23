package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// capability module.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&Capability{}, "cosmos-sdk/Capability", nil)
	cdc.RegisterConcrete(Owner{}, "cosmos-sdk/Owner", nil)
	cdc.RegisterConcrete(&CapabilityOwners{}, "cosmos-sdk/CapabilityOwners", nil)
}

var (
	amino = codec.NewLegacyAmino()
)

func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}
