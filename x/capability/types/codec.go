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

	// ModuleCdc references the global x/capability module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/capability and
	// defined at the application level.
	ModuleCdc = codec.NewHybridCodec(amino)
)

func init() {
	RegisterCodec(amino)
	amino.Seal()
}
