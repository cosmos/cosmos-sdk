package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgUnjail{}, "cosmos-sdk/MsgUnjail", nil)
}

// ModuleCdc defines the module codec
//var ModuleCdc *codec.Codec

type Codec struct {
	codec.Marshaler
	// Keep reference to the amino codec to allow backwards compatibility along
	// with type, and interface registration.
	amino *codec.Codec
}

func NewCodec(amino *codec.Codec) *Codec {
	return &Codec{Marshaler: codec.NewHybridCodec(amino), amino: amino}
}

var ModuleCdc *Codec

func init() {
	ModuleCdc = NewCodec(codec.New())
	//ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc.amino)
	//codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.amino.Seal()
}
