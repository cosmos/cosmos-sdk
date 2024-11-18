package codec

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/cosmos/gogoproto/proto"
	gogotypes "github.com/cosmos/gogoproto/types"
	gogoprotoany "github.com/cosmos/gogoproto/types/any"
	"google.golang.org/protobuf/encoding/protojson"
	protov2 "google.golang.org/protobuf/proto"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	"cosmossdk.io/schema"
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

type protoCollValueCodec[T any] interface {
	collcodec.HasSchemaCodec[T]
	collcodec.ValueCodec[T]
}

// CollValue inits a collections.ValueCodec for a generic gogo protobuf message.
func CollValue[T any, PT protoMessage[T]](cdc interface {
	Marshal(proto.Message) ([]byte, error)
	Unmarshal([]byte, proto.Message) error
},
) protoCollValueCodec[T] {
	return &collValue[T, PT]{cdc.(Codec), proto.MessageName(PT(new(T)))}
}

type collValue[T any, PT protoMessage[T]] struct {
	cdc         Codec
	messageName string
}

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
	return "github.com/cosmos/gogoproto/" + c.messageName
}

type hasUnpackInterfaces interface {
	UnpackInterfaces(unpacker gogoprotoany.AnyUnpacker) error // Replace `AnyUnpacker` with the actual type from gogoprotoany
}

func (c collValue[T, PT]) SchemaCodec() (collcodec.SchemaCodec[T], error) {
	return FallbackSchemaCodec[T](
		func(v T) error {
			if unpackable, ok := any(v).(hasUnpackInterfaces); ok {
				return unpackable.UnpackInterfaces(c.cdc)
			}
			return nil
		},
	), nil
}

// FallbackSchemaCodec returns a fallback schema codec for T when one isn't explicitly
// specified with HasSchemaCodec. It maps all simple types directly to schema kinds
// and converts everything else to JSON String.
func FallbackSchemaCodec[T any](unpacker func(T) error) collcodec.SchemaCodec[T] {
	var t T
	kind := schema.KindForGoValue(t)
	if err := kind.Validate(); err == nil {
		return collcodec.SchemaCodec[T]{
			Fields: []schema.Field{{
				// we don't set any name so that this can be set to a good default by the caller
				Name: "",
				Kind: kind,
			}},
			// these can be nil because T maps directly to a schema value for this kind
			ToSchemaType:   nil,
			FromSchemaType: nil,
		}
	} else {
		// we default to encoding everything to JSON String
		return collcodec.SchemaCodec[T]{
			Fields: []schema.Field{{Kind: schema.StringKind}},
			ToSchemaType: func(t T) (any, error) {
				fmt.Println("type of t: ", reflect.TypeOf(t))
				if unpacker != nil {
					if err := unpacker(t); err != nil {
						return nil, err
					}
				}
				bz, err := json.Marshal(t)
				return string(json.RawMessage(bz)), err
			},
			FromSchemaType: func(a any) (T, error) {
				var t T
				sz, ok := a.(string)
				if !ok {
					return t, fmt.Errorf("expected string, got %T", a)
				}
				err := json.Unmarshal([]byte(sz), &t)
				return t, err
			},
		}
	}
}

type protoMessageV2[T any] interface {
	*T
	protov2.Message
}

// CollValueV2 is used for protobuf values of the newest google.golang.org/protobuf API.
func CollValueV2[T any, PT protoMessageV2[T]]() collcodec.ValueCodec[PT] {
	return &collValue2[T, PT]{
		messageName: string(PT(new(T)).ProtoReflect().Descriptor().FullName()),
	}
}

type collValue2[T any, PT protoMessageV2[T]] struct {
	messageName string
}

func (c collValue2[T, PT]) Encode(value PT) ([]byte, error) {
	protov2MarshalOpts := protov2.MarshalOptions{Deterministic: true}
	return protov2MarshalOpts.Marshal(value)
}

func (c collValue2[T, PT]) Decode(b []byte) (PT, error) {
	var value T
	err := protov2.Unmarshal(b, PT(&value))
	return &value, err
}

func (c collValue2[T, PT]) EncodeJSON(value PT) ([]byte, error) {
	return protojson.Marshal(value)
}

func (c collValue2[T, PT]) DecodeJSON(b []byte) (PT, error) {
	var value T
	err := protojson.Unmarshal(b, PT(&value))
	return &value, err
}

func (c collValue2[T, PT]) Stringify(value PT) string {
	return fmt.Sprintf("%v", value)
}

func (c collValue2[T, PT]) ValueType() string {
	return "google.golang.org/protobuf/" + c.messageName
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
