package codec

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/cosmos/gogoproto/proto"
	gogotypes "github.com/cosmos/gogoproto/types"
	"google.golang.org/protobuf/encoding/protojson"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

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

func (c collValue[T, PT]) SchemaCodec() (collcodec.SchemaCodec[T], error) {
	var (
		t  T
		pt PT
	)
	msgName := proto.MessageName(pt)
	desc, err := proto.HybridResolver.FindDescriptorByName(protoreflect.FullName(msgName))
	if err != nil {
		return collcodec.SchemaCodec[T]{}, fmt.Errorf("could not find descriptor for %s: %w", msgName, err)
	}
	schemaFields := protoCols(desc.(protoreflect.MessageDescriptor))

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
		}, nil
	} else {
		return collcodec.SchemaCodec[T]{
			Fields: schemaFields,
			ToSchemaType: func(t T) (any, error) {
				values := []interface{}{}
				msgDesc, ok := desc.(protoreflect.MessageDescriptor)
				if !ok {
					return nil, fmt.Errorf("expected message descriptor, got %T", desc)
				}

				nm := dynamicpb.NewMessage(msgDesc)
				bz, err := c.cdc.Marshal(any(&t).(PT))
				if err != nil {
					return nil, err
				}

				err = c.cdc.Unmarshal(bz, nm)
				if err != nil {
					return nil, err
				}

				for _, field := range schemaFields {
					// Find the field descriptor by the Protobuf field name
					fieldDesc := msgDesc.Fields().ByName(protoreflect.Name(field.Name))
					if fieldDesc == nil {
						return nil, fmt.Errorf("field %q not found in message %s", field.Name, desc.FullName())
					}

					val := nm.ProtoReflect().Get(fieldDesc)

					// if the field is a map or list, we need to convert it to a slice of values
					if fieldDesc.IsList() {
						repeatedVals := []interface{}{}
						list := val.List()
						for i := 0; i < list.Len(); i++ {
							repeatedVals = append(repeatedVals, list.Get(i).Interface())
						}
						values = append(values, repeatedVals)
						continue
					}

					switch fieldDesc.Kind() {
					case protoreflect.BoolKind:
						values = append(values, val.Bool())
					case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
						protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
						values = append(values, val.Int())
					case protoreflect.Uint32Kind, protoreflect.Fixed32Kind, protoreflect.Uint64Kind,
						protoreflect.Fixed64Kind:
						values = append(values, val.Uint())
					case protoreflect.FloatKind, protoreflect.DoubleKind:
						values = append(values, val.Float())
					case protoreflect.StringKind:
						values = append(values, val.String())
					case protoreflect.BytesKind:
						values = append(values, val.Bytes())
					case protoreflect.EnumKind:
						// TODO: postgres uses the enum name, not the number
						values = append(values, string(fieldDesc.Enum().Values().ByNumber(val.Enum()).Name()))
					case protoreflect.MessageKind:
						msg := val.Interface().(*dynamicpb.Message)
						msgbz, err := c.cdc.Marshal(msg)
						if err != nil {
							return nil, err
						}

						if field.Kind == schema.TimeKind {
							// make it a time.Time
							ts := &timestamppb.Timestamp{}
							err = c.cdc.Unmarshal(msgbz, ts)
							if err != nil {
								return nil, fmt.Errorf("error unmarshalling timestamp: %w %x %s", err, msgbz, fieldDesc.FullName())
							}
							values = append(values, ts.AsTime())
						} else if field.Kind == schema.DurationKind {
							// make it a time.Duration
							dur := &durationpb.Duration{}
							err = c.cdc.Unmarshal(msgbz, dur)
							if err != nil {
								return nil, fmt.Errorf("error unmarshalling duration: %w", err)
							}
							values = append(values, dur.AsDuration())
						} else {
							// if not a time or duration, just keep it as a JSON object
							// we might want to change this to include the entire object as separate fields
							bz, err := c.cdc.MarshalJSON(msg)
							if err != nil {
								return nil, fmt.Errorf("error marshaling message: %w", err)
							}

							values = append(values, json.RawMessage(bz))
						}
					}

				}

				// if there's only one value, return it directly
				if len(values) == 1 {
					return values[0], nil
				}
				return values, nil
			},
			FromSchemaType: func(a any) (T, error) {
				panic("not implemented")
			},
		}, nil
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

// SchemaCodec returns a schema codec, which will always have a single JSON field
// as there is no way to know in advance the necessary fields for an interface.
func (c collInterfaceValue[T]) SchemaCodec() (collcodec.SchemaCodec[T], error) {
	var pt T

	kind := schema.KindForGoValue(pt)
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
		}, nil
	} else {
		return collcodec.SchemaCodec[T]{
			Fields: []schema.Field{{
				Name: "value",
				Kind: schema.JSONKind,
			}},
			ToSchemaType: func(t T) (any, error) {
				bz, err := c.codec.MarshalInterfaceJSON(t)
				if err != nil {
					return nil, err
				}

				return json.RawMessage(bz), nil
			},
			FromSchemaType: func(a any) (T, error) {
				panic("not implemented")
			},
		}, nil
	}
}

func protoCols(desc protoreflect.MessageDescriptor) []schema.Field {
	nFields := desc.Fields()
	cols := make([]schema.Field, 0, nFields.Len())
	for i := 0; i < nFields.Len(); i++ {
		f := nFields.Get(i)
		cols = append(cols, protoCol(f))
	}
	return cols
}

func protoCol(f protoreflect.FieldDescriptor) schema.Field {
	col := schema.Field{Name: string(f.Name())}
	if f.IsMap() || f.IsList() {
		col.Kind = schema.JSONKind
		col.Nullable = true
	} else {
		switch f.Kind() {
		case protoreflect.BoolKind:
			col.Kind = schema.BoolKind
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
			col.Kind = schema.Int32Kind
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			col.Kind = schema.Int64Kind
		case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
			col.Kind = schema.Int64Kind
		case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
			col.Kind = schema.Uint64Kind
		case protoreflect.FloatKind:
			col.Kind = schema.Float32Kind
		case protoreflect.DoubleKind:
			col.Kind = schema.Float64Kind
		case protoreflect.StringKind:
			col.Kind = schema.StringKind
		case protoreflect.BytesKind:
			col.Kind = schema.BytesKind
			col.Nullable = true
		case protoreflect.EnumKind:
			// TODO: support enums
			col.Kind = schema.EnumKind
			// use the full name to avoid collisions
			col.ReferencedType = string(f.Enum().FullName())
			col.ReferencedType = strings.ReplaceAll(col.ReferencedType, ".", "_")
		case protoreflect.MessageKind:
			col.Nullable = true
			fullName := f.Message().FullName()
			if fullName == "google.protobuf.Timestamp" {
				col.Kind = schema.TimeKind
			} else if fullName == "google.protobuf.Duration" {
				col.Kind = schema.DurationKind
			} else {
				col.Kind = schema.JSONKind
			}
		}
		if f.HasPresence() {
			col.Nullable = true
		}
	}

	return col
}
