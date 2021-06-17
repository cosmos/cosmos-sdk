package codec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	amino "github.com/tendermint/go-amino"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

// Deprecated: LegacyAmino defines a wrapper for an Amino codec that properly handles protobuf
// types with Any's
type LegacyAmino struct {
	Amino *amino.Codec
}

func (cdc *LegacyAmino) Seal() {
	cdc.Amino.Seal()
}

func NewLegacyAmino() *LegacyAmino {
	return &LegacyAmino{amino.NewCodec()}
}

// RegisterEvidences registers Tendermint evidence types with the provided Amino
// codec.
func RegisterEvidences(cdc *LegacyAmino) {
	cdc.Amino.RegisterInterface((*tmtypes.Evidence)(nil), nil)
	cdc.Amino.RegisterConcrete(&tmtypes.DuplicateVoteEvidence{}, "tendermint/DuplicateVoteEvidence", nil)
}

// MarshalJSONIndent provides a utility for indented JSON encoding of an object
// via an Amino codec. It returns an error if it cannot serialize or indent as
// JSON.
func MarshalJSONIndent(cdc *LegacyAmino, obj interface{}) ([]byte, error) {
	bz, err := cdc.MarshalJSON(obj)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err = json.Indent(&out, bz, "", "  "); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

// MustMarshalJSONIndent executes MarshalJSONIndent except it panics upon failure.
func MustMarshalJSONIndent(cdc *LegacyAmino, obj interface{}) []byte {
	bz, err := MarshalJSONIndent(cdc, obj)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %s", err))
	}

	return bz
}

func (cdc *LegacyAmino) marshalAnys(o interface{}) error {
	return types.UnpackInterfaces(o, types.AminoPacker{Cdc: cdc.Amino})
}

func (cdc *LegacyAmino) unmarshalAnys(o interface{}) error {
	return types.UnpackInterfaces(o, types.AminoUnpacker{Cdc: cdc.Amino})
}

func (cdc *LegacyAmino) jsonMarshalAnys(o interface{}) error {
	return types.UnpackInterfaces(o, types.AminoJSONPacker{Cdc: cdc.Amino})
}

func (cdc *LegacyAmino) jsonUnmarshalAnys(o interface{}) error {
	return types.UnpackInterfaces(o, types.AminoJSONUnpacker{Cdc: cdc.Amino})
}

func (cdc *LegacyAmino) MarshalBinaryBare(o interface{}) ([]byte, error) {
	err := cdc.marshalAnys(o)
	if err != nil {
		return nil, err
	}
	return cdc.Amino.MarshalBinaryBare(o)
}

func (cdc *LegacyAmino) MustMarshalBinaryBare(o interface{}) []byte {
	bz, err := cdc.MarshalBinaryBare(o)
	if err != nil {
		panic(err)
	}
	return bz
}

func (cdc *LegacyAmino) MarshalBinaryLengthPrefixed(o interface{}) ([]byte, error) {
	err := cdc.marshalAnys(o)
	if err != nil {
		return nil, err
	}
	return cdc.Amino.MarshalBinaryLengthPrefixed(o)
}

func (cdc *LegacyAmino) MustMarshalBinaryLengthPrefixed(o interface{}) []byte {
	bz, err := cdc.MarshalBinaryLengthPrefixed(o)
	if err != nil {
		panic(err)
	}
	return bz
}

func (cdc *LegacyAmino) UnmarshalBinaryBare(bz []byte, ptr interface{}) error {
	err := cdc.Amino.UnmarshalBinaryBare(bz, ptr)
	if err != nil {
		return err
	}
	return cdc.unmarshalAnys(ptr)
}

func (cdc *LegacyAmino) MustUnmarshalBinaryBare(bz []byte, ptr interface{}) {
	err := cdc.UnmarshalBinaryBare(bz, ptr)
	if err != nil {
		panic(err)
	}
}

func (cdc *LegacyAmino) UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	err := cdc.Amino.UnmarshalBinaryLengthPrefixed(bz, ptr)
	if err != nil {
		return err
	}
	return cdc.unmarshalAnys(ptr)
}

func (cdc *LegacyAmino) MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) {
	err := cdc.UnmarshalBinaryLengthPrefixed(bz, ptr)
	if err != nil {
		panic(err)
	}
}

// MarshalJSON implements codec.Marshaler interface
func (cdc *LegacyAmino) MarshalJSON(o interface{}) ([]byte, error) {
	err := cdc.jsonMarshalAnys(o)
	if err != nil {
		return nil, err
	}
	return cdc.Amino.MarshalJSON(o)
}

func (cdc *LegacyAmino) MustMarshalJSON(o interface{}) []byte {
	bz, err := cdc.MarshalJSON(o)
	if err != nil {
		panic(err)
	}
	return bz
}

// UnmarshalJSON implements codec.Marshaler interface
func (cdc *LegacyAmino) UnmarshalJSON(bz []byte, ptr interface{}) error {
	err := cdc.Amino.UnmarshalJSON(bz, ptr)
	if err != nil {
		return err
	}
	return cdc.jsonUnmarshalAnys(ptr)
}

func (cdc *LegacyAmino) MustUnmarshalJSON(bz []byte, ptr interface{}) {
	err := cdc.UnmarshalJSON(bz, ptr)
	if err != nil {
		panic(err)
	}
}

func (*LegacyAmino) UnpackAny(*types.Any, interface{}) error {
	return errors.New("AminoCodec can't handle unpack protobuf Any's")
}

func (cdc *LegacyAmino) RegisterInterface(ptr interface{}, iopts *amino.InterfaceOptions) {
	cdc.Amino.RegisterInterface(ptr, iopts)
}

func (cdc *LegacyAmino) RegisterConcrete(o interface{}, name string, copts *amino.ConcreteOptions) {
	cdc.Amino.RegisterConcrete(o, name, copts)
}

func (cdc *LegacyAmino) MarshalJSONIndent(o interface{}, prefix, indent string) ([]byte, error) {
	err := cdc.jsonMarshalAnys(o)
	if err != nil {
		panic(err)
	}
	return cdc.Amino.MarshalJSONIndent(o, prefix, indent)
}

func (cdc *LegacyAmino) PrintTypes(out io.Writer) error {
	return cdc.Amino.PrintTypes(out)
}
