package codec

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

// MarshalAny is a convenience function for proto marshalling interfaces. It
// packs the provided value in an Any and then marshals it to bytes.
// NOTE: if you use a concret type, then you should use BinaryMarshaler.MarshalBinaryBare directly
func MarshalAny(m BinaryMarshaler, x interface{}) ([]byte, error) {
	msg, ok := x.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("can't proto marshal %T - expecting proto.Message", x)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return m.MarshalBinaryBare(any)
}

// UnmarshalAny is a convenience function for proto unmarshaling interfaces. It
// unmarshals an Any from bz and then unpacks it to the `iface`, which must
// be a pointer to a non empty interface with registered implementations.
// NOTE: if you use a concret type, then you should use BinaryMarshaler.UnarshalBinaryBare directly
//
// Ex:
//		var x MyInterface
//		err := UnmarshalAny(unpacker, &x, bz)
func UnmarshalAny(m BinaryMarshaler, iface interface{}, bz []byte) error {
	any := &types.Any{}

	err := m.UnmarshalBinaryBare(bz, any)
	if err != nil {
		return err
	}

	return m.UnpackAny(any, iface)
}

// MarshalAnyJSON is a convenience function for proto marshalling interfaces. It
// packs the provided value in an Any and then marshals it to bytes.
// NOTE: if you use a concret type, then you should use JSONMarshaler.MarshalJSON directly
func MarshalAnyJSON(m JSONMarshaler, x proto.Message) ([]byte, error) {
	any, err := types.NewAnyWithValue(x)
	if err != nil {
		return nil, err
	}
	return m.MarshalJSON(any)
}

// UnmarshalAnyJSON is a convenience function for proto unmarshaling interfaces.
// It unmarshals an Any from bz and then unpacks it to the `iface`, which must
// be a pointer to a non empty interface with registered implementations.
// NOTE: if you use a concret type, then you should use JSONMarshaler.UnarshalJSON directly
//
// Ex:
//		var x MyInterface
//		err := UnmarshalAny(unpacker, &x, bz)
func UnmarshalAnyJSON(m JSONAnyMarshaler, iface interface{}, bz []byte) error {
	any := &types.Any{}
	err := m.UnmarshalJSON(bz, any)
	if err != nil {
		return err
	}
	return m.UnpackAny(any, iface)
}
