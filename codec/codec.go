package codec

import (
	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

type (
	// Codec defines the interface module codecs must implement in order to support
	// backwards compatibility with Amino while allowing custom Protobuf-based
	// serialization. Note, Amino can still be used without any dependency on
	// Protobuf. There are two typical implementations that fulfill this contract:
	//
	// 1. AminoCodec: Provides full Amino serialization compatibility.
	// 2. ProtoCodec: Provides full Protobuf serialization compatibility.
	Codec interface {
		BinaryCodec
		JSONCodec
	}

	BinaryCodec interface {
		Marshal(o ProtoMarshaler) ([]byte, error)
		MustMarshal(o ProtoMarshaler) []byte

		MarshalLengthPrefixed(o ProtoMarshaler) ([]byte, error)
		MustMarshalLengthPrefixed(o ProtoMarshaler) []byte

		Unmarshal(bz []byte, ptr ProtoMarshaler) error
		MustUnmarshal(bz []byte, ptr ProtoMarshaler)

		UnmarshalLengthPrefixed(bz []byte, ptr ProtoMarshaler) error
		MustUnmarshalLengthPrefixed(bz []byte, ptr ProtoMarshaler)

		MarshalInterface(i proto.Message) ([]byte, error)
		UnmarshalInterface(bz []byte, ptr interface{}) error

		types.AnyUnpacker
	}

	JSONCodec interface {
		MarshalJSON(o proto.Message) ([]byte, error)
		MustMarshalJSON(o proto.Message) []byte
		MarshalInterfaceJSON(i proto.Message) ([]byte, error)
		UnmarshalInterfaceJSON(bz []byte, ptr interface{}) error

		UnmarshalJSON(bz []byte, ptr proto.Message) error
		MustUnmarshalJSON(bz []byte, ptr proto.Message)
	}

	// ProtoMarshaler defines an interface a type must implement to serialize itself
	// as protocol buffer defined message.
	ProtoMarshaler interface {
		proto.Message // for JSON serialization

		Marshal() ([]byte, error)
		MarshalTo(data []byte) (n int, err error)
		MarshalToSizedBuffer(dAtA []byte) (int, error)
		Size() int
		Unmarshal(data []byte) error
	}

	// AminoMarshaler defines an interface a type must implement to serialize itself
	// for Amino codec.
	AminoMarshaler interface {
		MarshalAmino() ([]byte, error)
		UnmarshalAmino([]byte) error
		MarshalAminoJSON() ([]byte, error)
		UnmarshalAminoJSON([]byte) error
	}
)
