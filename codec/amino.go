package codec

import (
	"bytes"
	"encoding/json"
	"fmt"

	amino "github.com/tendermint/go-amino"
	cryptoamino "github.com/tendermint/tendermint/crypto/encoding/amino"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

// Cdc defines a global generic sealed Amino codec to be used throughout sdk. It
// has all Tendermint crypto and evidence types registered.
//
// TODO: Consider removing this global.
var Cdc *Codec

func init() {
	cdc := New()
	RegisterCrypto(cdc)
	RegisterEvidences(cdc)
	Cdc = cdc.Seal()
}

// Codec defines a type alias for an Amino codec.
type Codec struct {
	Amino *amino.Codec
}

var _ JSONMarshaler = &Codec{}

func (cdc *Codec) Seal() *Codec {
	return &Codec{cdc.Amino.Seal()}
}

func New() *Codec {
	return &Codec{amino.NewCodec()}
}

// RegisterCrypto registers all crypto dependency types with the provided Amino
// codec.
func RegisterCrypto(cdc *Codec) {
	cryptoamino.RegisterAmino(cdc.Amino)
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

func (ac *Codec) marshalAnys(o interface{}) error {
	return types.UnpackInterfaces(o, types.AminoPacker{Cdc: ac.Amino})
}

func (ac *Codec) unmarshalAnys(o interface{}) error {
	return types.UnpackInterfaces(o, types.AminoUnpacker{Cdc: ac.Amino})
}

func (ac *Codec) jsonMarshalAnys(o interface{}) error {
	return types.UnpackInterfaces(o, types.AminoJSONPacker{Cdc: ac.Amino})
}

func (ac *Codec) jsonUnmarshalAnys(o interface{}) error {
	return types.UnpackInterfaces(o, types.AminoJSONUnpacker{Cdc: ac.Amino})
}

func (ac *Codec) MarshalBinaryBare(o interface{}) ([]byte, error) {
	err := ac.marshalAnys(o)
	if err != nil {
		return nil, err
	}
	return ac.Amino.MarshalBinaryBare(o)
}

func (ac *Codec) MustMarshalBinaryBare(o interface{}) []byte {
	err := ac.marshalAnys(o)
	if err != nil {
		panic(err)
	}
	return ac.Amino.MustMarshalBinaryBare(o)
}

func (ac *Codec) MarshalBinaryLengthPrefixed(o interface{}) ([]byte, error) {
	err := ac.marshalAnys(o)
	if err != nil {
		return nil, err
	}
	return ac.Amino.MarshalBinaryLengthPrefixed(o)
}

func (ac *Codec) MustMarshalBinaryLengthPrefixed(o interface{}) []byte {
	err := ac.marshalAnys(o)
	if err != nil {
		panic(err)
	}
	return ac.Amino.MustMarshalBinaryLengthPrefixed(o)
}

func (ac *Codec) UnmarshalBinaryBare(bz []byte, ptr interface{}) error {
	err := ac.Amino.UnmarshalBinaryBare(bz, ptr)
	if err != nil {
		return err
	}
	return ac.unmarshalAnys(ptr)
}

func (ac *Codec) MustUnmarshalBinaryBare(bz []byte, ptr interface{}) {
	ac.Amino.MustUnmarshalBinaryBare(bz, ptr)
	err := ac.unmarshalAnys(ptr)
	if err != nil {
		panic(err)
	}
}

func (ac *Codec) UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	err := ac.Amino.UnmarshalBinaryLengthPrefixed(bz, ptr)
	if err != nil {
		return err
	}
	return ac.unmarshalAnys(ptr)
}

func (ac *Codec) MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) {
	ac.Amino.MustUnmarshalBinaryLengthPrefixed(bz, ptr)
	err := ac.unmarshalAnys(ptr)
	if err != nil {
		panic(err)
	}
}

func (ac *Codec) MarshalJSON(o interface{}) ([]byte, error) {
	err := ac.jsonMarshalAnys(o)
	if err != nil {
		return nil, err
	}
	return ac.Amino.MarshalJSON(o)
}

func (ac *Codec) MustMarshalJSON(o interface{}) []byte {
	err := ac.jsonMarshalAnys(o)
	if err != nil {
		panic(err)
	}
	return ac.Amino.MustMarshalJSON(o)
}

func (ac *Codec) UnmarshalJSON(bz []byte, ptr interface{}) error {
	err := ac.Amino.UnmarshalJSON(bz, ptr)
	if err != nil {
		return err
	}
	return ac.jsonUnmarshalAnys(ptr)
}

func (ac *Codec) MustUnmarshalJSON(bz []byte, ptr interface{}) {
	ac.Amino.MustUnmarshalJSON(bz, ptr)
	err := ac.jsonUnmarshalAnys(ptr)
	if err != nil {
		panic(err)
	}
}

func (*Codec) UnpackAny(*types.Any, interface{}) error {
	return fmt.Errorf("AminoCodec can't handle unpack protobuf Any's")
}

func (cdc *Codec) RegisterInterface(ptr interface{}, iopts *amino.InterfaceOptions) {
	cdc.Amino.RegisterInterface(ptr, iopts)
}

func (cdc *Codec) RegisterConcrete(o interface{}, name string, copts *amino.ConcreteOptions) {
	cdc.Amino.RegisterConcrete(o, name, copts)
}
