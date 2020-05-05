package types

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
)

// Codec defines the interface required to serialize evidence
type Codec interface {
	codec.Marshaler

	MarshalEvidence(exported.Evidence) ([]byte, error)
	UnmarshalEvidence([]byte) (exported.Evidence, error)
	MarshalEvidenceJSON(exported.Evidence) ([]byte, error)
	UnmarshalEvidenceJSON([]byte) (exported.Evidence, error)
}

// RegisterCodec registers all the necessary types and interfaces for the
// evidence module.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.Evidence)(nil), nil)
	cdc.RegisterConcrete(MsgSubmitEvidence{}, "cosmos-sdk/MsgSubmitEvidence", nil)
	cdc.RegisterConcrete(&Equivocation{}, "cosmos-sdk/Equivocation", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSubmitEvidence{})
	registry.RegisterInterface(
		"cosmos_sdk.evidence.v1.Evidence",
		(*exported.Evidence)(nil),
		&Equivocation{},
	)
}

var (
	amino = codec.New()

	// ModuleCdc references the global x/evidence module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/evidence and
	// defined at the application level.
	ModuleCdc = codec.NewHybridCodec(amino, types.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(amino)
	codec.RegisterCrypto(amino)
	amino.Seal()
}

// AnyCodec is an evidence Codec that marshals evidence using google.protobuf.Any
type AnyCodec struct {
	codec.Marshaler
}

// NewAnyCodec returns a new AnyCodec
func NewAnyCodec(marshaler codec.Marshaler) Codec {
	return AnyCodec{Marshaler: marshaler}
}

// MarshalEvidence marshals an Evidence interface. If the given type implements
// the Marshaler interface, it is treated as a Proto-defined message and
// serialized that way. Otherwise, it falls back on the internal Amino codec.
func (c AnyCodec) MarshalEvidence(evidenceI exported.Evidence) ([]byte, error) {
	return types.MarshalAny(evidenceI)
}

// UnmarshalEvidence returns an Evidence interface from raw encoded evidence
// bytes of a Proto-based Evidence type. An error is returned upon decoding
// failure.
func (c AnyCodec) UnmarshalEvidence(bz []byte) (exported.Evidence, error) {
	var evi exported.Evidence
	err := types.UnmarshalAny(c, &evi, bz)
	if err != nil {
		return nil, err
	}
	return evi, nil
}

// MarshalEvidenceJSON JSON encodes an evidence object implementing the Evidence
// interface.
func (c AnyCodec) MarshalEvidenceJSON(evidence exported.Evidence) ([]byte, error) {
	msg, ok := evidence.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("cannot proto marshal %T", evidence)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}
	return c.MarshalJSON(any)
}

// UnmarshalEvidenceJSON returns an Evidence from JSON encoded bytes
func (c AnyCodec) UnmarshalEvidenceJSON(bz []byte) (exported.Evidence, error) {
	var any types.Any
	if err := c.UnmarshalJSON(bz, &any); err != nil {
		return nil, err
	}

	var evi exported.Evidence
	if err := c.UnpackAny(&any, &evi); err != nil {
		return nil, err
	}
	return evi, nil
}
