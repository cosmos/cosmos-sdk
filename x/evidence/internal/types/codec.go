package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
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

// ModuleCdc defines the evidence module's codec. The codec is not sealed as to
// allow other modules to register their concrete Evidence types.
var ModuleCdc *Codec

// RegisterCodec registers all the necessary types and interfaces for the
// evidence module.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.EvidenceI)(nil), nil)
	cdc.RegisterConcrete(MsgSubmitEvidence{}, "cosmos-sdk/MsgSubmitEvidence", nil)
	cdc.RegisterConcrete(Equivocation{}, "cosmos-sdk/Equivocation", nil)
}

// RegisterEvidenceTypeCodec registers an external concrete Evidence type defined
// in another module for the internal ModuleCdc. This allows the MsgSubmitEvidence
// to be correctly Amino encoded and decoded.
func RegisterEvidenceTypeCodec(o interface{}, name string) {
	ModuleCdc.amino.RegisterConcrete(o, name, nil)
}

func init() {
	ModuleCdc = NewCodec(codec.New())
	RegisterCodec(ModuleCdc.amino)
}
