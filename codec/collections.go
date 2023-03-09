package codec

import (
	"bytes"
	"fmt"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"

	"github.com/cosmos/gogoproto/proto"
)

// BoolValue implements a ValueCodec that saves the bool value
// as if it was a prototypes.BoolValue. Required for backwards
// compatibility of state.
var BoolValue collcodec.ValueCodec[bool] = boolValue{}

var (
	boolValueTrueBytes  = []byte{0x8, 0x1}
	boolValueFalseBytes = []byte{}
)

type boolValue struct{}

func (boolValue) Encode(value bool) ([]byte, error) {
	if value {
		return boolValueTrueBytes, nil
	}
	return boolValueFalseBytes, nil
}

func (boolValue) Decode(b []byte) (bool, error) {
	switch {
	case bytes.Equal(b, boolValueFalseBytes):
		return false, nil
	case bytes.Equal(b, boolValueTrueBytes):
		return true, nil
	default:
		return false, fmt.Errorf("%w: %s", collcodec.ErrEncoding, "invalid bool value bytes")
	}
}

func (boolValue) EncodeJSON(value bool) ([]byte, error) {
	return collections.BoolValue.EncodeJSON(value)
}

func (boolValue) DecodeJSON(b []byte) (bool, error) {
	return collections.BoolValue.DecodeJSON(b)
}

func (boolValue) Stringify(value bool) string {
	return collections.BoolValue.Stringify(value)
}

func (boolValue) ValueType() string {
	return "protobuf/bool"
}

type protoMessage[T any] interface {
	*T
	proto.Message
}

// CollValue inits a collections.ValueCodec for a generic gogo protobuf message.
func CollValue[T any, PT protoMessage[T]](cdc BinaryCodec) collcodec.ValueCodec[T] {
	return &collValue[T, PT]{cdc.(Codec)}
}

type collValue[T any, PT protoMessage[T]] struct{ cdc Codec }

func (c collValue[T, PT]) Encode(value T) ([]byte, error) {
	return c.cdc.Marshal(PT(&value))
}

func (c collValue[T, PT]) Decode(b []byte) (value T, err error) {
	err = c.cdc.Unmarshal(b, PT(&value))
	return value, err
}

func (c collValue[T, PT]) EncodeJSON(value T) ([]byte, error) {
	return c.cdc.MarshalJSON(PT(&value))
}

func (c collValue[T, PT]) DecodeJSON(b []byte) (value T, err error) {
	err = c.cdc.UnmarshalJSON(b, PT(&value))
	return
}

func (c collValue[T, PT]) Stringify(value T) string {
	return PT(&value).String()
}

func (c collValue[T, PT]) ValueType() string {
	return "gogoproto/" + proto.MessageName(PT(new(T)))
}
