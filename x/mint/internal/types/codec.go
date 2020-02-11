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

// generic sealed codec to be used throughout this module
var ModuleCdc *Codec

func init() {
	ModuleCdc = NewCodec(codec.New())
	codec.RegisterCrypto(ModuleCdc.amino)
	ModuleCdc.amino.Seal()
}
