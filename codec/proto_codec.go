package codec

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/gogo/protobuf/jsonpb"
	gogoproto "github.com/gogo/protobuf/proto"
	protov2 "google.golang.org/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

// ProtoCodecMarshaler defines an interface for codecs that utilize Protobuf for both
// binary and JSON encoding.
type ProtoCodecMarshaler interface {
	Codec
	InterfaceRegistry() types.InterfaceRegistry
}

// ProtoCodec defines a codec that utilizes Protobuf for both binary and JSON
// encoding.
type ProtoCodec struct {
	interfaceRegistry types.InterfaceRegistry
}

var _ Codec = &ProtoCodec{}
var _ ProtoCodecMarshaler = &ProtoCodec{}

// NewProtoCodec returns a reference to a new ProtoCodec
func NewProtoCodec(interfaceRegistry types.InterfaceRegistry) *ProtoCodec {
	return &ProtoCodec{interfaceRegistry: interfaceRegistry}
}

// Marshal implements BinaryMarshaler.Marshal method.
// NOTE: this function must be used with a concrete type which
// implements proto.Message. For interface please use the codec.MarshalInterface
func (pc *ProtoCodec) Marshal(o interface{}) ([]byte, error) {
	switch o := o.(type) {
	case ProtoMarshaler:
		return o.Marshal()
	case protov2.Message:
		return protov2.MarshalOptions{Deterministic: true}.Marshal(o.ProtoReflect().Interface())
	default:
		return nil, fmt.Errorf("%T is not a v1 or v2 message", o)
	}
}

// MustMarshal implements BinaryMarshaler.MustMarshal method.
// NOTE: this function must be used with a concrete type which
// implements proto.Message. For interface please use the codec.MarshalInterface
func (pc *ProtoCodec) MustMarshal(o interface{}) []byte {
	bz, err := pc.Marshal(o)
	if err != nil {
		panic(err)
	}

	return bz
}

// MarshalLengthPrefixed implements BinaryMarshaler.MarshalLengthPrefixed method.
func (pc *ProtoCodec) MarshalLengthPrefixed(o interface{}) ([]byte, error) {
	bz, err := pc.Marshal(o)
	if err != nil {
		return nil, err
	}

	var sizeBuf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(sizeBuf[:], uint64(len(bz)))
	return append(sizeBuf[:n], bz...), nil
}

// MustMarshalLengthPrefixed implements BinaryMarshaler.MustMarshalLengthPrefixed method.
func (pc *ProtoCodec) MustMarshalLengthPrefixed(o interface{}) []byte {
	bz, err := pc.MarshalLengthPrefixed(o)
	if err != nil {
		panic(err)
	}

	return bz
}

// Unmarshal implements BinaryMarshaler.Unmarshal method.
// NOTE: this function must be used with a concrete type which
// implements proto.Message. For interface please use the codec.UnmarshalInterface
func (pc *ProtoCodec) Unmarshal(bz []byte, ptr interface{}) error {
	switch ptr := ptr.(type) {
	case protov2.Message:
		err := protov2.Unmarshal(bz, ptr.ProtoReflect().Interface())
		if err != nil {
			return err
		}
	case ProtoMarshaler:
		err := ptr.Unmarshal(bz)
		if err == nil {
			return err
		} else {
			return err
		}
	default:
		return fmt.Errorf("%T is not a v1 or v2 message", ptr)
	}

	err := types.UnpackInterfaces(ptr, pc.interfaceRegistry)
	if err != nil {
		return err
	}
	return nil
}

// MustUnmarshal implements BinaryMarshaler.MustUnmarshal method.
// NOTE: this function must be used with a concrete type which
// implements proto.Message. For interface please use the codec.UnmarshalInterface
func (pc *ProtoCodec) MustUnmarshal(bz []byte, ptr interface{}) {
	if err := pc.Unmarshal(bz, ptr); err != nil {
		panic(err)
	}
}

// UnmarshalLengthPrefixed implements BinaryMarshaler.UnmarshalLengthPrefixed method.
func (pc *ProtoCodec) UnmarshalLengthPrefixed(bz []byte, ptr interface{}) error {
	size, n := binary.Uvarint(bz)
	if n < 0 {
		return fmt.Errorf("invalid number of bytes read from length-prefixed encoding: %d", n)
	}

	if size > uint64(len(bz)-n) {
		return fmt.Errorf("not enough bytes to read; want: %v, got: %v", size, len(bz)-n)
	} else if size < uint64(len(bz)-n) {
		return fmt.Errorf("too many bytes to read; want: %v, got: %v", size, len(bz)-n)
	}

	bz = bz[n:]
	return pc.Unmarshal(bz, ptr)
}

// MustUnmarshalLengthPrefixed implements BinaryMarshaler.MustUnmarshalLengthPrefixed method.
func (pc *ProtoCodec) MustUnmarshalLengthPrefixed(bz []byte, ptr interface{}) {
	if err := pc.UnmarshalLengthPrefixed(bz, ptr); err != nil {
		panic(err)
	}
}

// MarshalJSON implements JSONCodec.MarshalJSON method,
// it marshals to JSON using proto codec.
// NOTE: this function must be used with a concrete type which
// implements proto.Message. For interface please use the codec.MarshalInterfaceJSON
func (pc *ProtoCodec) MarshalJSON(o interface{}) ([]byte, error) {
	return ProtoMarshalJSON(o, pc.interfaceRegistry)
}

// MustMarshalJSON implements JSONCodec.MustMarshalJSON method,
// it executes MarshalJSON except it panics upon failure.
// NOTE: this function must be used with a concrete type which
// implements proto.Message. For interface please use the codec.MarshalInterfaceJSON
func (pc *ProtoCodec) MustMarshalJSON(o interface{}) []byte {
	bz, err := pc.MarshalJSON(o)
	if err != nil {
		panic(err)
	}

	return bz
}

// UnmarshalJSON implements JSONCodec.UnmarshalJSON method,
// it unmarshals from JSON using proto codec.
// NOTE: this function must be used with a concrete type which
// implements proto.Message. For interface please use the codec.UnmarshalInterfaceJSON
func (pc *ProtoCodec) UnmarshalJSON(bz []byte, ptr interface{}) error {
	switch ptr := ptr.(type) {
	case gogoproto.Message:
		unmarshaler := jsonpb.Unmarshaler{AnyResolver: pc.interfaceRegistry}
		err := unmarshaler.Unmarshal(strings.NewReader(string(bz)), ptr)
		if err != nil {
			return err
		}
	case protov2.Message:
		err := protojson.UnmarshalOptions{Resolver: pc.interfaceRegistry}.Unmarshal(bz, ptr)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("%T is not a v1 or v2 message", ptr)
	}

	return types.UnpackInterfaces(ptr, pc.interfaceRegistry)
}

// MustUnmarshalJSON implements JSONCodec.MustUnmarshalJSON method,
// it executes UnmarshalJSON except it panics upon failure.
// NOTE: this function must be used with a concrete type which
// implements proto.Message. For interface please use the codec.UnmarshalInterfaceJSON
func (pc *ProtoCodec) MustUnmarshalJSON(bz []byte, ptr interface{}) {
	if err := pc.UnmarshalJSON(bz, ptr); err != nil {
		panic(err)
	}
}

// MarshalInterface is a convenience function for proto marshalling interfaces. It packs
// the provided value, which must be an interface, in an Any and then marshals it to bytes.
// NOTE: to marshal a concrete type, you should use Marshal instead
func (pc *ProtoCodec) MarshalInterface(i interface{}) ([]byte, error) {
	if err := assertNotNil(i); err != nil {
		return nil, err
	}
	any, err := types.NewAnyWithValue(i)
	if err != nil {
		return nil, err
	}

	return pc.Marshal(any)
}

// UnmarshalInterface is a convenience function for proto unmarshaling interfaces. It
// unmarshals an Any from bz bytes and then unpacks it to the `ptr`, which must
// be a pointer to a non empty interface with registered implementations.
// NOTE: to unmarshal a concrete type, you should use Unmarshal instead
//
// Example:
//    var x MyInterface
//    err := cdc.UnmarshalInterface(bz, &x)
func (pc *ProtoCodec) UnmarshalInterface(bz []byte, ptr interface{}) error {
	any := &types.Any{}
	err := pc.Unmarshal(bz, any)
	if err != nil {
		return err
	}

	return pc.UnpackAny(any, ptr)
}

// MarshalInterfaceJSON is a convenience function for proto marshalling interfaces. It
// packs the provided value in an Any and then marshals it to bytes.
// NOTE: to marshal a concrete type, you should use MarshalJSON instead
func (pc *ProtoCodec) MarshalInterfaceJSON(x interface{}) ([]byte, error) {
	any, err := types.NewAnyWithValue(x)
	if err != nil {
		return nil, err
	}
	return pc.MarshalJSON(any)
}

// UnmarshalInterfaceJSON is a convenience function for proto unmarshaling interfaces.
// It unmarshals an Any from bz bytes and then unpacks it to the `iface`, which must
// be a pointer to a non empty interface, implementing proto.Message with registered implementations.
// NOTE: to unmarshal a concrete type, you should use UnmarshalJSON instead
//
// Example:
//    var x MyInterface  // must implement proto.Message
//    err := cdc.UnmarshalInterfaceJSON(&x, bz)
func (pc *ProtoCodec) UnmarshalInterfaceJSON(bz []byte, iface interface{}) error {
	any := &types.Any{}
	err := pc.UnmarshalJSON(bz, any)
	if err != nil {
		return err
	}
	return pc.UnpackAny(any, iface)
}

// UnpackAny implements AnyUnpacker.UnpackAny method,
// it unpacks the value in any to the interface pointer passed in as
// iface.
func (pc *ProtoCodec) UnpackAny(any *types.Any, iface interface{}) error {
	return pc.interfaceRegistry.UnpackAny(any, iface)
}

// InterfaceRegistry returns InterfaceRegistry
func (pc *ProtoCodec) InterfaceRegistry() types.InterfaceRegistry {
	return pc.interfaceRegistry
}

func assertNotNil(i interface{}) error {
	if i == nil {
		return errors.New("can't marshal <nil> value")
	}
	return nil
}
