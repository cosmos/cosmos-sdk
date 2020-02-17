package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
)

// EvidenceCodec defines the interface required to serialize evidence
type EvidenceCodec interface {
	codec.Marshaler
	MarshalEvidence(acc exported.EvidenceI) ([]byte, error)
	UnmarshalEvidence(bz []byte) (exported.EvidenceI, error)
	MarshalEvidenceJSON(acc exported.EvidenceI) ([]byte, error)
	UnmarshalEvidenceJSON(bz []byte) (exported.EvidenceI, error)
}

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
var ModuleCdc = NewCodec(codec.New())

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
	RegisterCodec(ModuleCdc.amino)
}

// MarshalEvidence marshals an EvidenceI interface.
func (c *Codec) MarshalEvidence(eviI exported.EvidenceI) ([]byte, error) {
	evi := &Evidence{}
	evi.SetEvidenceI(eviI)
	return c.MarshalBinaryLengthPrefixed(evi)
}

// UnmarshalEvidence returnes an EvidenceI interface.
func (c *Codec) UnmarshalEvidence(bz []byte) (exported.EvidenceI, error) {
	evi := &Evidence{}
	if err := c.UnmarshalBinaryLengthPrefixed(bz, evi); err != nil {
		return nil, err
	}
	return evi.GetEvidenceI(), nil
}

// MarshalEvidenceJSON JSON encodes an evidence object implementating the EvidenceI interface
func (c *Codec) MarshalEvidenceJSON(evi exported.EvidenceI) ([]byte, error) {
	return c.MarshalJSON(evi)
}

// UnmarshalEvidenceJSON returns an EvidenceI from JSON encoded bytes
func (c *Codec) UnmarshalEvidenceJSON(bz []byte) (exported.EvidenceI, error) {
	evi := &Evidence{}
	if err := c.UnmarshalJSON(bz, evi); err != nil {
		return nil, err
	}
	return evi.GetEvidenceI(), nil
}
