package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
)

// EvidenceCodec defines the interface required to serialize evidence
type Codec interface {
	codec.Marshaler

	MarshalEvidence(acc exported.EvidenceI) ([]byte, error)
	UnmarshalEvidence(bz []byte) (exported.EvidenceI, error)
	MarshalEvidenceJSON(acc exported.EvidenceI) ([]byte, error)
	UnmarshalEvidenceJSON(bz []byte) (exported.EvidenceI, error)
}

var (
	// Keep reference to the amino codec to allow backwards compatibility along
	// with type, and interface registration.
	amino *codec.Codec

	ModuleCdc = codec.NewHybridCodec(amino)
)

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
	RegisterCodec(amino)
}
