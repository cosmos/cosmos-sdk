package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// ModuleCdc defines the capability module's codec. The codec is not sealed as to
// allow other modules to register their concrete Capability types.
var ModuleCdc = codec.New()

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
	ModuleCdc.RegisterConcrete(o, name, nil)
}

func init() {
	RegisterCodec(ModuleCdc)
}
