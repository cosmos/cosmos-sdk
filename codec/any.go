package codec

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

// TODO: For better UX this functions should be moved to the Marshaler interface
// - user should have a consisten way how to unpack and serialize data without thinking
// to use a helper function or an interface.

// MarshalIfc is a convenience function for proto marshalling interfaces. It
// packs the provided value in an Any and then marshals it to bytes.
// NOTE: if you use a concert type, then you should use BinaryMarshaler.MarshalBinaryBare directly
func MarshalIfc(m BinaryMarshaler, x interface{}) ([]byte, error) {
	msg, ok := x.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("can't proto marshal %T; expecting proto.Message", x)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return m.MarshalBinaryBare(any)
}

// UnmarshalIfc is a convenience function for proto unmarshaling interfaces. It
// unmarshals an Any from bz and then unpacks it to the `iface`, which must
// be a pointer to a non empty interface with registered implementations.
// NOTE: if you use a concert type, then you should use BinaryMarshaler.UnarshalBinaryBare directly
//
// Ex:
//		var x MyInterface
//		err := UnmarshalIfc(unpacker, &x, bz)
func UnmarshalIfc(m BinaryMarshaler, iface interface{}, bz []byte) error {
	any := &types.Any{}

	err := m.UnmarshalBinaryBare(bz, any)
	if err != nil {
		return err
	}

	return m.UnpackAny(any, iface)
}

// MarshalIfcJSON is a convenience function for proto marshalling interfaces. It
// packs the provided value in an Any and then marshals it to bytes.
// NOTE: if you use a concert type, then you should use JSONMarshaler.MarshalJSON directly
func MarshalIfcJSON(m JSONMarshaler, x proto.Message) ([]byte, error) {
	any, err := types.NewAnyWithValue(x)
	if err != nil {
		return nil, err
	}
	return m.MarshalJSON(any)
}

// UnmarshalIfcJSON is a convenience function for proto unmarshaling interfaces.
// It unmarshals an Any from bz and then unpacks it to the `iface`, which must
// be a pointer to a non empty interface with registered implementations.
// NOTE: if you use a concert type, then you should use JSONMarshaler.UnarshalJSON directly
//
// Ex:
//		var x MyInterface
//		err := UnmarshalAny(unpacker, &x, bz)
func UnmarshalIfcJSON(m IfcJSONMarshaler, iface interface{}, bz []byte) error {
	any := &types.Any{}
	err := m.UnmarshalJSON(bz, any)
	if err != nil {
		return err
	}
	return m.UnpackAny(any, iface)
}
