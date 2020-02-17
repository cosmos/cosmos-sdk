package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

type Codec struct {
	codec.Marshaler

	// Keep reference to the amino codec to allow backwards compatibility along
	// with type, and interface registration.
	amino *codec.Codec
}

func NewCodec(amino *codec.Codec) *Codec {
	return &Codec{Marshaler: codec.NewHybridCodec(amino), amino: amino}
}

// ----------------------------------------------------------------------------

// RegisterCodec registers all the necessary crisis module concrete types and
// interfaces with the provided codec reference.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgVerifyInvariant{}, "cosmos-sdk/MsgVerifyInvariant", nil)
}

// ModuleCdc defines a crisis module global Amino codec.
var ModuleCdc *Codec

func init() {
	ModuleCdc = NewCodec(codec.New())
	RegisterCodec(ModuleCdc.amino)
	codec.RegisterCrypto(ModuleCdc.amino)
	ModuleCdc.amino.Seal()
}
