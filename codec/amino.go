package codec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	cmttypes "github.com/cometbft/cometbft/v2/types"
	amino "github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

// LegacyAmino defines a wrapper for an Amino codec that properly
// handles protobuf types with Any's. Deprecated.
type LegacyAmino struct {
	Amino *amino.Codec
}

func (cdc *LegacyAmino) Seal() {
	cdc.Amino.Seal()
}

func NewLegacyAmino() *LegacyAmino {
	return &LegacyAmino{amino.NewCodec()}
}

// RegisterEvidences registers CometBFT evidence types with the provided Amino
// codec.
func RegisterEvidences(cdc *LegacyAmino) {
	cdc.Amino.RegisterInterface((*cmttypes.Evidence)(nil), nil)
	cdc.Amino.RegisterConcrete(&cmttypes.DuplicateVoteEvidence{}, "tendermint/DuplicateVoteEvidence", nil)
}

// MarshalJSONIndent provides a utility for indented JSON encoding of an object
// via an Amino codec. It returns an error if it cannot serialize or indent as
// JSON.
func MarshalJSONIndent(cdc *LegacyAmino, obj any) ([]byte, error) {
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
func MustMarshalJSONIndent(cdc *LegacyAmino, obj any) []byte {
	bz, err := MarshalJSONIndent(cdc, obj)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %s", err))
	}

	return bz
}

func (cdc *LegacyAmino) marshalAnys(o any) error {
	return types.UnpackInterfaces(o, types.AminoPacker{Cdc: cdc.Amino})
}

func (cdc *LegacyAmino) unmarshalAnys(o any) error {
	return types.UnpackInterfaces(o, types.AminoUnpacker{Cdc: cdc.Amino})
}

func (cdc *LegacyAmino) jsonMarshalAnys(o any) error {
	return types.UnpackInterfaces(o, types.AminoJSONPacker{Cdc: cdc.Amino})
}

func (cdc *LegacyAmino) jsonUnmarshalAnys(o any) error {
	return types.UnpackInterfaces(o, types.AminoJSONUnpacker{Cdc: cdc.Amino})
}

func (cdc *LegacyAmino) Marshal(o any) ([]byte, error) {
	err := cdc.marshalAnys(o)
	if err != nil {
		return nil, err
	}
	return cdc.Amino.MarshalBinaryBare(o)
}

func (cdc *LegacyAmino) MustMarshal(o any) []byte {
	bz, err := cdc.Marshal(o)
	if err != nil {
		panic(err)
	}
	return bz
}

func (cdc *LegacyAmino) MarshalLengthPrefixed(o any) ([]byte, error) {
	err := cdc.marshalAnys(o)
	if err != nil {
		return nil, err
	}
	return cdc.Amino.MarshalBinaryLengthPrefixed(o)
}

func (cdc *LegacyAmino) MustMarshalLengthPrefixed(o any) []byte {
	bz, err := cdc.MarshalLengthPrefixed(o)
	if err != nil {
		panic(err)
	}
	return bz
}

func (cdc *LegacyAmino) Unmarshal(bz []byte, ptr any) error {
	err := cdc.Amino.UnmarshalBinaryBare(bz, ptr)
	if err != nil {
		return err
	}
	return cdc.unmarshalAnys(ptr)
}

func (cdc *LegacyAmino) MustUnmarshal(bz []byte, ptr any) {
	err := cdc.Unmarshal(bz, ptr)
	if err != nil {
		panic(err)
	}
}

func (cdc *LegacyAmino) UnmarshalLengthPrefixed(bz []byte, ptr any) error {
	err := cdc.Amino.UnmarshalBinaryLengthPrefixed(bz, ptr)
	if err != nil {
		return err
	}
	return cdc.unmarshalAnys(ptr)
}

func (cdc *LegacyAmino) MustUnmarshalLengthPrefixed(bz []byte, ptr any) {
	err := cdc.UnmarshalLengthPrefixed(bz, ptr)
	if err != nil {
		panic(err)
	}
}

// MarshalJSON implements codec.Codec interface
func (cdc *LegacyAmino) MarshalJSON(o any) ([]byte, error) {
	err := cdc.jsonMarshalAnys(o)
	if err != nil {
		return nil, err
	}
	return cdc.Amino.MarshalJSON(o)
}

func (cdc *LegacyAmino) MustMarshalJSON(o any) []byte {
	bz, err := cdc.MarshalJSON(o)
	if err != nil {
		panic(err)
	}
	return bz
}

// UnmarshalJSON implements codec.Codec interface
func (cdc *LegacyAmino) UnmarshalJSON(bz []byte, ptr any) error {
	err := cdc.Amino.UnmarshalJSON(bz, ptr)
	if err != nil {
		return err
	}
	return cdc.jsonUnmarshalAnys(ptr)
}

func (cdc *LegacyAmino) MustUnmarshalJSON(bz []byte, ptr any) {
	err := cdc.UnmarshalJSON(bz, ptr)
	if err != nil {
		panic(err)
	}
}

func (*LegacyAmino) UnpackAny(*types.Any, any) error {
	return errors.New("AminoCodec can't handle unpack protobuf Any's")
}

func (cdc *LegacyAmino) RegisterInterface(ptr any, iopts *amino.InterfaceOptions) {
	cdc.Amino.RegisterInterface(ptr, iopts)
}

func (cdc *LegacyAmino) RegisterConcrete(o any, name string, copts *amino.ConcreteOptions) {
	cdc.Amino.RegisterConcrete(o, name, copts)
}

func (cdc *LegacyAmino) MarshalJSONIndent(o any, prefix, indent string) ([]byte, error) {
	err := cdc.jsonMarshalAnys(o)
	if err != nil {
		panic(err)
	}
	return cdc.Amino.MarshalJSONIndent(o, prefix, indent)
}

func (cdc *LegacyAmino) PrintTypes(out io.Writer) error {
	return cdc.Amino.PrintTypes(out)
}
