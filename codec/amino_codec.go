package codec

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

// AminoCodec defines a codec that utilizes Amino for both binary and JSON
// encoding.
type AminoCodec struct {
	amino *Codec
}

func NewAminoCodec(amino *Codec) Marshaler {
	return &AminoCodec{amino}
}

func (ac *AminoCodec) marshalAnys(o ProtoMarshaler) error {
	return types.UnpackInterfaces(o, types.AminoPacker{Cdc: ac.amino})
}

func (ac *AminoCodec) unmarshalAnys(o ProtoMarshaler) error {
	return types.UnpackInterfaces(o, types.AminoUnpacker{Cdc: ac.amino})
}

func (ac *AminoCodec) jsonMarshalAnys(o interface{}) error {
	return types.UnpackInterfaces(o, types.AminoJSONPacker{Cdc: ac.amino})
}

func (ac *AminoCodec) jsonUnmarshalAnys(o interface{}) error {
	return types.UnpackInterfaces(o, types.AminoJSONUnpacker{Cdc: ac.amino})
}

func (ac *AminoCodec) MarshalBinaryBare(o ProtoMarshaler) ([]byte, error) {
	err := ac.marshalAnys(o)
	if err != nil {
		return nil, err
	}
	return ac.amino.MarshalBinaryBare(o)
}

func (ac *AminoCodec) MustMarshalBinaryBare(o ProtoMarshaler) []byte {
	err := ac.marshalAnys(o)
	if err != nil {
		panic(err)
	}
	return ac.amino.MustMarshalBinaryBare(o)
}

func (ac *AminoCodec) MarshalBinaryLengthPrefixed(o ProtoMarshaler) ([]byte, error) {
	err := ac.marshalAnys(o)
	if err != nil {
		return nil, err
	}
	return ac.amino.MarshalBinaryLengthPrefixed(o)
}

func (ac *AminoCodec) MustMarshalBinaryLengthPrefixed(o ProtoMarshaler) []byte {
	err := ac.marshalAnys(o)
	if err != nil {
		panic(err)
	}
	return ac.amino.MustMarshalBinaryLengthPrefixed(o)
}

func (ac *AminoCodec) UnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler) error {
	err := ac.amino.UnmarshalBinaryBare(bz, ptr)
	if err != nil {
		return err
	}
	return ac.unmarshalAnys(ptr)
}

func (ac *AminoCodec) MustUnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler) {
	ac.amino.MustUnmarshalBinaryBare(bz, ptr)
	err := ac.unmarshalAnys(ptr)
	if err != nil {
		panic(err)
	}
}

func (ac *AminoCodec) UnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler) error {
	err := ac.amino.UnmarshalBinaryLengthPrefixed(bz, ptr)
	if err != nil {
		return err
	}
	return ac.unmarshalAnys(ptr)
}

func (ac *AminoCodec) MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler) {
	ac.amino.MustUnmarshalBinaryLengthPrefixed(bz, ptr)
	err := ac.unmarshalAnys(ptr)
	if err != nil {
		panic(err)
	}
}

func (ac *AminoCodec) MarshalJSON(o interface{}) ([]byte, error) {
	err := ac.jsonMarshalAnys(o)
	if err != nil {
		return nil, err
	}
	return ac.amino.MarshalJSON(o)
}

func (ac *AminoCodec) MustMarshalJSON(o interface{}) []byte {
	err := ac.jsonMarshalAnys(o)
	if err != nil {
		panic(err)
	}
	return ac.amino.MustMarshalJSON(o)
}

func (ac *AminoCodec) UnmarshalJSON(bz []byte, ptr interface{}) error {
	err := ac.amino.UnmarshalJSON(bz, ptr)
	if err != nil {
		return err
	}
	return ac.jsonUnmarshalAnys(ptr)
}

func (ac *AminoCodec) MustUnmarshalJSON(bz []byte, ptr interface{}) {
	ac.amino.MustUnmarshalJSON(bz, ptr)
	err := ac.jsonUnmarshalAnys(ptr)
	if err != nil {
		panic(err)
	}
}

func (*AminoCodec) UnpackAny(*types.Any, interface{}) error {
	return fmt.Errorf("AminoCodec can't handle unpack protobuf Any's")
}
