package codec

import (
	"github.com/gogo/protobuf/proto"
)

// AminoCodec defines a codec that utilizes Codec for both binary and JSON
// encoding.
type AminoCodec struct {
	*LegacyAmino
}

var _ Marshaler = &AminoCodec{}

// NewAminoCodec returns a reference to a new AminoCodec
func NewAminoCodec(codec *LegacyAmino) *AminoCodec {
	return &AminoCodec{LegacyAmino: codec}
}

// MarshalBinaryBare implements BinaryMarshaler.MarshalBinaryBare method.
func (ac *AminoCodec) MarshalBinaryBare(o ProtoMarshaler) ([]byte, error) {
	return ac.LegacyAmino.MarshalBinaryBare(o)
}

// MustMarshalBinaryBare implements BinaryMarshaler.MustMarshalBinaryBare method.
func (ac *AminoCodec) MustMarshalBinaryBare(o ProtoMarshaler) []byte {
	return ac.LegacyAmino.MustMarshalBinaryBare(o)
}

// MarshalBinaryLengthPrefixed implements BinaryMarshaler.MarshalBinaryLengthPrefixed method.
func (ac *AminoCodec) MarshalBinaryLengthPrefixed(o ProtoMarshaler) ([]byte, error) {
	return ac.LegacyAmino.MarshalBinaryLengthPrefixed(o)
}

// MustMarshalBinaryLengthPrefixed implements BinaryMarshaler.MustMarshalBinaryLengthPrefixed method.
func (ac *AminoCodec) MustMarshalBinaryLengthPrefixed(o ProtoMarshaler) []byte {
	return ac.LegacyAmino.MustMarshalBinaryLengthPrefixed(o)
}

// UnmarshalBinaryBare implements BinaryMarshaler.UnmarshalBinaryBare method.
func (ac *AminoCodec) UnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler) error {
	return ac.LegacyAmino.UnmarshalBinaryBare(bz, ptr)
}

// MustUnmarshalBinaryBare implements BinaryMarshaler.MustUnmarshalBinaryBare method.
func (ac *AminoCodec) MustUnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler) {
	ac.LegacyAmino.MustUnmarshalBinaryBare(bz, ptr)
}

// UnmarshalBinaryLengthPrefixed implements BinaryMarshaler.UnmarshalBinaryLengthPrefixed method.
func (ac *AminoCodec) UnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler) error {
	return ac.LegacyAmino.UnmarshalBinaryLengthPrefixed(bz, ptr)
}

// MustUnmarshalBinaryLengthPrefixed implements BinaryMarshaler.MustUnmarshalBinaryLengthPrefixed method.
func (ac *AminoCodec) MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler) {
	ac.LegacyAmino.MustUnmarshalBinaryLengthPrefixed(bz, ptr)
}

// MarshalJSON implements JSONMarshaler.MarshalJSON method,
// it marshals to JSON using legacy amino codec.
func (ac *AminoCodec) MarshalJSON(o proto.Message) ([]byte, error) {
	return ac.LegacyAmino.MarshalJSON(o)
}

// MustMarshalJSON implements JSONMarshaler.MustMarshalJSON method,
// it executes MarshalJSON except it panics upon failure.
func (ac *AminoCodec) MustMarshalJSON(o proto.Message) []byte {
	return ac.LegacyAmino.MustMarshalJSON(o)
}

// UnmarshalJSON implements JSONMarshaler.UnmarshalJSON method,
// it unmarshals from JSON using legacy amino codec.
func (ac *AminoCodec) UnmarshalJSON(bz []byte, ptr proto.Message) error {
	return ac.LegacyAmino.UnmarshalJSON(bz, ptr)
}

// MustUnmarshalJSON implements JSONMarshaler.MustUnmarshalJSON method,
// it executes UnmarshalJSON except it panics upon failure.
func (ac *AminoCodec) MustUnmarshalJSON(bz []byte, ptr proto.Message) {
	ac.LegacyAmino.MustUnmarshalJSON(bz, ptr)
}

// MarshalInterface is a convenience function for amino marshaling interfaces.
// The `i` must be an interface.
// NOTE: to marshal a concrete type, you should use MarshalBinaryBare instead
func (ac *AminoCodec) MarshalInterface(i proto.Message) ([]byte, error) {
	if err := assertNotNil(i); err != nil {
		return nil, err
	}
	return ac.LegacyAmino.MarshalBinaryBare(i)
}

// UnmarshalInterface is a convenience function for amino unmarshaling interfaces.
// `ptr` must be a pointer to an interface.
// NOTE: to unmarshal a concrete type, you should use UnmarshalBinaryBare instead
//
// Example:
//   var x MyInterface
//   err := cdc.UnmarshalInterface(bz, &x)
func (ac *AminoCodec) UnmarshalInterface(bz []byte, ptr interface{}) error {
	return ac.LegacyAmino.UnmarshalBinaryBare(bz, ptr)
}

// MarshalInterfaceJSON is a convenience function for amino marshaling interfaces.
// The `i` must be an interface.
// NOTE: to marshal a concrete type, you should use MarshalJSON instead
func (ac *AminoCodec) MarshalInterfaceJSON(i proto.Message) ([]byte, error) {
	if err := assertNotNil(i); err != nil {
		return nil, err
	}
	return ac.LegacyAmino.MarshalJSON(i)
}

// UnmarshalInterfaceJSON is a convenience function for amino unmarshaling interfaces.
// `ptr` must be a pointer to an interface.
// NOTE: to unmarshal a concrete type, you should use UnmarshalJSON instead
//
// Example:
//   var x MyInterface
//   err := cdc.UnmarshalInterfaceJSON(bz, &x)
func (ac *AminoCodec) UnmarshalInterfaceJSON(bz []byte, ptr interface{}) error {
	return ac.LegacyAmino.UnmarshalJSON(bz, ptr)
}
