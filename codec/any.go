package codec

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

// MarshalAny is a convenience function for packing the provided value in an
// Any and then proto marshaling it to bytes
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

// UnmarshalAny is a convenience function for proto unmarshaling an Any from
// bz and then unpacking it to the interface pointer passed in as iface using
// the provided AnyUnpacker or returning an error
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

// MarshalAnyJSON is a convenience function for packing the provided value in an
// Any and then proto marshaling into JSON
func MarshalAnyJSON(m JSONMarshaler, x proto.Message) ([]byte, error) {
	any, err := types.NewAnyWithValue(x)
	if err != nil {
		return nil, err
	}
	return m.MarshalJSON(any)
}

// UnmarshalAnyJSON is a convenience function for unmarshaling an Any from
// JSON bytes and then unpacking it to the `iface` pointer using the provided
// AnyUnpacker or returning an error
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
