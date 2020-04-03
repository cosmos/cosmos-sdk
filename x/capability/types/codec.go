package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers all the necessary types and interfaces for the
// capability module.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Capability)(nil), nil)
	cdc.RegisterConcrete(&CapabilityKey{}, "cosmos-sdk/CapabilityKey", nil)
	cdc.RegisterConcrete(Owner{}, "cosmos-sdk/Owner", nil)
	cdc.RegisterConcrete(&CapabilityOwners{}, "cosmos-sdk/CapabilityOwners", nil)
}

// RegisterCapabilityTypeCodec registers an external concrete Capability type
// defined in another module for the internal ModuleCdc.
func RegisterCapabilityTypeCodec(o interface{}, name string) {
	amino.RegisterConcrete(o, name, nil)
	ModuleCdc = codec.NewHybridCodec(amino)
}

var (
// The amino codec is not sealed as to
// allow other modules to register their concrete Capability types.
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
}
