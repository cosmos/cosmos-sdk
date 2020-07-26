package codec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	amino "github.com/tendermint/go-amino"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/KiraCore/cosmos-sdk/codec/types"
)

// deprecated: Codec defines a wrapper for an Amino codec that properly handles protobuf
// types with Any's
type Codec struct {
	Amino *amino.Codec
}

var _ JSONMarshaler = &Codec{}

func (cdc *Codec) Seal() {
	cdc.Amino.Seal()
}

func New() *Codec {
	return &Codec{amino.NewCodec()}
}

// RegisterEvidences registers Tendermint evidence types with the provided Amino
// codec.
func RegisterEvidences(cdc *Codec) {
	tmtypes.RegisterEvidences(cdc.Amino)
}

// MarshalJSONIndent provides a utility for indented JSON encoding of an object
// via an Amino codec. It returns an error if it cannot serialize or indent as
// JSON.
func MarshalJSONIndent(m JSONMarshaler, obj interface{}) ([]byte, error) {
	bz, err := m.MarshalJSON(obj)
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
func MustMarshalJSONIndent(m JSONMarshaler, obj interface{}) []byte {
	bz, err := MarshalJSONIndent(m, obj)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %s", err))
	}

	return bz
}

func (cdc *Codec) marshalAnys(o interface{}) error {
	return types.UnpackInterfaces(o, types.AminoPacker{Cdc: cdc.Amino})
}

func (cdc *Codec) unmarshalAnys(o interface{}) error {
	return types.UnpackInterfaces(o, types.AminoUnpacker{Cdc: cdc.Amino})
}

func (cdc *Codec) jsonMarshalAnys(o interface{}) error {
	return types.UnpackInterfaces(o, types.AminoJSONPacker{Cdc: cdc.Amino})
}

func (cdc *Codec) jsonUnmarshalAnys(o interface{}) error {
	return types.UnpackInterfaces(o, types.AminoJSONUnpacker{Cdc: cdc.Amino})
}

func (cdc *Codec) MarshalBinaryBare(o interface{}) ([]byte, error) {
	err := cdc.marshalAnys(o)
	if err != nil {
		return nil, err
	}
	return cdc.Amino.MarshalBinaryBare(o)
}

func (cdc *Codec) MustMarshalBinaryBare(o interface{}) []byte {
	bz, err := cdc.MarshalBinaryBare(o)
	if err != nil {
		panic(err)
	}
	return bz
}

func (cdc *Codec) MarshalBinaryLengthPrefixed(o interface{}) ([]byte, error) {
	err := cdc.marshalAnys(o)
	if err != nil {
		return nil, err
	}
	return cdc.Amino.MarshalBinaryLengthPrefixed(o)
}

func (cdc *Codec) MustMarshalBinaryLengthPrefixed(o interface{}) []byte {
	bz, err := cdc.MarshalBinaryLengthPrefixed(o)
	if err != nil {
		panic(err)
	}
	return bz
}

func (cdc *Codec) UnmarshalBinaryBare(bz []byte, ptr interface{}) error {
	err := cdc.Amino.UnmarshalBinaryBare(bz, ptr)
	if err != nil {
		return err
	}
	return cdc.unmarshalAnys(ptr)
}

func (cdc *Codec) MustUnmarshalBinaryBare(bz []byte, ptr interface{}) {
	err := cdc.UnmarshalBinaryBare(bz, ptr)
	if err != nil {
		panic(err)
	}
}

func (cdc *Codec) UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	err := cdc.Amino.UnmarshalBinaryLengthPrefixed(bz, ptr)
	if err != nil {
		return err
	}
	return cdc.unmarshalAnys(ptr)
}

func (cdc *Codec) MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) {
	err := cdc.UnmarshalBinaryLengthPrefixed(bz, ptr)
	if err != nil {
		panic(err)
	}
}

func (cdc *Codec) MarshalJSON(o interface{}) ([]byte, error) {
	err := cdc.jsonMarshalAnys(o)
	if err != nil {
		return nil, err
	}
	return cdc.Amino.MarshalJSON(o)
}

func (cdc *Codec) MustMarshalJSON(o interface{}) []byte {
	bz, err := cdc.MarshalJSON(o)
	if err != nil {
		panic(err)
	}
	return bz
}

func (cdc *Codec) UnmarshalJSON(bz []byte, ptr interface{}) error {
	err := cdc.Amino.UnmarshalJSON(bz, ptr)
	if err != nil {
		return err
	}
	return cdc.jsonUnmarshalAnys(ptr)
}

func (cdc *Codec) MustUnmarshalJSON(bz []byte, ptr interface{}) {
	err := cdc.UnmarshalJSON(bz, ptr)
	if err != nil {
		panic(err)
	}
}

func (*Codec) UnpackAny(*types.Any, interface{}) error {
	return errors.New("AminoCodec can't handle unpack protobuf Any's")
}

func (cdc *Codec) RegisterInterface(ptr interface{}, iopts *amino.InterfaceOptions) {
	cdc.Amino.RegisterInterface(ptr, iopts)
}

func (cdc *Codec) RegisterConcrete(o interface{}, name string, copts *amino.ConcreteOptions) {
	cdc.Amino.RegisterConcrete(o, name, copts)
}

func (cdc *Codec) MarshalJSONIndent(o interface{}, prefix, indent string) ([]byte, error) {
	err := cdc.jsonMarshalAnys(o)
	if err != nil {
		panic(err)
	}
	return cdc.Amino.MarshalJSONIndent(o, prefix, indent)
}

func (cdc *Codec) PrintTypes(out io.Writer) error {
	return cdc.Amino.PrintTypes(out)
}
