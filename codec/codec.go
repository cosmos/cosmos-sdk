package codec

import (
	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

type (
	// Marshaler defines the interface module codecs must implement in order to support
	// backwards compatibility with Amino while allowing custom Protobuf-based
	// serialization. Note, Amino can still be used without any dependency on
	// Protobuf. There are three typical implementations that fulfill this contract:
	//
	// 1. AminoCodec: Provides full Amino serialization compatibility.
	// 2. ProtoCodec: Provides full Protobuf serialization compatibility.
	Marshaler interface {
		BinaryMarshaler
		JSONMarshaler
	}

	BinaryMarshaler interface {
		MarshalBinaryBare(o ProtoMarshaler) ([]byte, error)
		MustMarshalBinaryBare(o ProtoMarshaler) []byte

		MarshalBinaryLengthPrefixed(o ProtoMarshaler) ([]byte, error)
		MustMarshalBinaryLengthPrefixed(o ProtoMarshaler) []byte

		UnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler) error
		MustUnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler)

		UnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler) error
		MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler)

		types.AnyUnpacker
	}

	JSONMarshaler interface {
		MarshalJSON(o interface{}) ([]byte, error)
		MustMarshalJSON(o interface{}) []byte

		UnmarshalJSON(bz []byte, ptr interface{}) error
		MustUnmarshalJSON(bz []byte, ptr interface{})
	}

	// ProtoMarshaler defines an interface a type must implement as protocol buffer
	// defined message.
	ProtoMarshaler interface {
		proto.Message // for JSON serialization

		Marshal() ([]byte, error)
		MarshalTo(data []byte) (n int, err error)
		MarshalToSizedBuffer(dAtA []byte) (int, error)
		Size() int
		Unmarshal(data []byte) error
	}
)
