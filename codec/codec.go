package codec

import (
	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

type (
	// Codec defines a functionality for serializing other objects.
	// Users can define a custom Protobuf-based serialization.
	// Note, Amino can still be used without any dependency on Protobuf.
	// SDK provides to Codec implementations:
	//
	// 1. AminoCodec: Provides full Amino serialization compatibility.
	// 2. ProtoCodec: Provides full Protobuf serialization compatibility.
	Codec interface {
		BinaryCodec
		JSONCodec
	}

	BinaryCodec interface {
		// Marshal returns binary encoding of v.
		Marshal(o interface{}) ([]byte, error)
		// MustMarshal calls Marshal and panics if error is returned.
		MustMarshal(o interface{}) []byte

		// MarshalLengthPrefixed returns binary encoding of v with bytes length prefix.
		MarshalLengthPrefixed(o interface{}) ([]byte, error)
		// MustMarshalLengthPrefixed calls MarshalLengthPrefixed and panics if
		// error is returned.
		MustMarshalLengthPrefixed(o interface{}) []byte

		// Unmarshal parses the data encoded with Marshal method and stores the result
		// in the value pointed to by v.
		Unmarshal(bz []byte, ptr interface{}) error
		// MustUnmarshal calls Unmarshal and panics if error is returned.
		MustUnmarshal(bz []byte, ptr interface{})

		// UnmarshalLengthPrefixed parses the data encoded with UnmarshalLengthPrefixed method and stores
		// the result in the value pointed to by v.
		UnmarshalLengthPrefixed(bz []byte, ptr interface{}) error
		// MustUnmarshalLengthPrefixed calls UnmarshalLengthPrefixed and panics if error
		// is returned.
		MustUnmarshalLengthPrefixed(bz []byte, ptr interface{})

		// MarshalInterface is a helper method which will wrap `i` into `Any` for correct
		// binary interface (de)serialization.
		MarshalInterface(i interface{}) ([]byte, error)
		// UnmarshalInterface is a helper method which will parse binary enoded data
		// into `Any` and unpack any into the `ptr`. It fails if the target interface type
		// is not registered in codec, or is not compatible with the serialized data
		UnmarshalInterface(bz []byte, ptr interface{}) error

		types.AnyUnpacker
	}

	JSONCodec interface {
		// MarshalJSON returns JSON encoding of v.
		MarshalJSON(o interface{}) ([]byte, error)
		// MustMarshalJSON calls MarshalJSON and panics if error is returned.
		MustMarshalJSON(o interface{}) []byte
		// MarshalInterfaceJSON is a helper method which will wrap `i` into `Any` for correct
		// JSON interface (de)serialization.
		MarshalInterfaceJSON(i interface{}) ([]byte, error)
		// UnmarshalInterfaceJSON is a helper method which will parse JSON encoded data
		// into `Any` and unpack any into the `ptr`. It fails if the target interface type
		// is not registered in codec, or is not compatible with the serialized data
		UnmarshalInterfaceJSON(bz []byte, ptr interface{}) error

		// UnmarshalJSON parses the data encoded with MarshalJSON method and stores the result
		// in the value pointed to by v.
		UnmarshalJSON(bz []byte, ptr interface{}) error
		// MustUnmarshalJSON calls Unmarshal and panics if error is returned.
		MustUnmarshalJSON(bz []byte, ptr interface{})
	}

	// ProtoMarshaler defines an interface a type must implement to serialize itself
	// as a protocol buffer defined message.
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
