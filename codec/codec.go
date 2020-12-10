package codec

import (
	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

type (
	// Marshaler defines the interface module codecs must implement in order to support
	// backwards compatibility with Amino while allowing custom Protobuf-based
	// serialization. Note, Amino can still be used without any dependency on
	// Protobuf. There are two typical implementations that fulfill this contract:
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

		MarshalInterface(i proto.Message) ([]byte, error)
		UnmarshalInterface(bz []byte, ptr interface{}) error

		types.AnyUnpacker
	}

	JSONMarshaler interface {
		MarshalJSON(o proto.Message) ([]byte, error)
		MustMarshalJSON(o proto.Message) []byte
		MarshalInterfaceJSON(i proto.Message) ([]byte, error)
		UnmarshalInterfaceJSON(bz []byte, ptr interface{}) error

		UnmarshalJSON(bz []byte, ptr proto.Message) error
		MustUnmarshalJSON(bz []byte, ptr proto.Message)
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

	// AminoMarshaler defines an interface where Amino marshalling can be
	// overridden by custom marshalling.
	AminoMarshaler interface {
		MarshalAmino() ([]byte, error)
		UnmarshalAmino([]byte) error
		MarshalAminoJSON() ([]byte, error)
		UnmarshalAminoJSON([]byte) error
	}
)
