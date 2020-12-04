package codec

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

// MarshalInterface is a convenience function for proto marshalling interfaces. It
// packs the provided value, which must implemenet proto.Message,
// in an Any and then marshals it to bytes.
// NOTE: if you use a concret type, then you should use BinaryMarshaler.MarshalBinaryBare instead
func MarshalInterface(m BinaryMarshaler, x interface{}) ([]byte, error) {
	msg, ok := x.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("can't proto marshal %T", x)
	}

	any := &types.Any{}
	err := any.Pack(msg)
	if err != nil {
		return nil, err
	}

	return m.MarshalBinaryBare(any)
}

// UnmarshalInterface is a convenience function for proto unmarshaling interfaces. It
// unmarshals an Any from bz and then unpacks it to the `iface`, which must
// be a pointer to a non empty interface with registered implementations.
// NOTE: if you use a concert type, then you should use BinaryMarshaler.UnarshalBinaryBare instead
//
// Ex:
//		var x MyInterface
//		err := UnmarshalInterface(unpacker, &x, bz)
func UnmarshalInterface(m BinaryMarshaler, iface interface{}, bz []byte) error {
	any := &types.Any{}

	err := m.UnmarshalBinaryBare(bz, any)
	if err != nil {
		return err
	}

	return m.UnpackAny(any, iface)
}
