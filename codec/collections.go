package codec

import (
	"fmt"
	"reflect"

	gogotypes "github.com/cosmos/gogoproto/types"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"

	"github.com/cosmos/gogoproto/proto"
)

// BoolValue implements a ValueCodec that saves the bool value
// as if it was a prototypes.BoolValue. Required for backwards
// compatibility of state.
var BoolValue collcodec.ValueCodec[bool] = boolValue{}

type boolValue struct{}

func (boolValue) Encode(value bool) ([]byte, error) {
	return (&gogotypes.BoolValue{Value: value}).Marshal()
}

func (boolValue) Decode(b []byte) (bool, error) {
	v := new(gogotypes.BoolValue)
	err := v.Unmarshal(b)
	return v.Value, err
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

// CollInterfaceValue instantiates a new collections.ValueCodec for a generic
// interface value. The codec must be able to marshal and unmarshal the
// interface.
func CollInterfaceValue[T proto.Message](codec BinaryCodec) collcodec.ValueCodec[T] {
	var x T // assertion
	if reflect.TypeOf(&x).Elem().Kind() != reflect.Interface {
		panic("CollInterfaceValue can only be used with interface types")
	}
	return collInterfaceValue[T]{codec.(Codec)}
}

type collInterfaceValue[T proto.Message] struct {
	codec Codec
}

func (c collInterfaceValue[T]) Encode(value T) ([]byte, error) {
	return c.codec.MarshalInterface(value)
}

func (c collInterfaceValue[T]) Decode(b []byte) (T, error) {
	var value T
	err := c.codec.UnmarshalInterface(b, &value)
	return value, err
}

func (c collInterfaceValue[T]) EncodeJSON(value T) ([]byte, error) {
	return c.codec.MarshalInterfaceJSON(value)
}

func (c collInterfaceValue[T]) DecodeJSON(b []byte) (T, error) {
	var value T
	err := c.codec.UnmarshalInterfaceJSON(b, &value)
	return value, err
}

func (c collInterfaceValue[T]) Stringify(value T) string {
	return value.String()
}

func (c collInterfaceValue[T]) ValueType() string {
	var t T
	return fmt.Sprintf("%T", t)
}
