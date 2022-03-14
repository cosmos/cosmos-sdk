package ormsql

import (
	"fmt"
	"reflect"

	"google.golang.org/protobuf/reflect/protoreflect"
	"gorm.io/datatypes"
)

type valueCodec interface {
	goType() reflect.Type
	encode(protoValue protoreflect.Value, goValue reflect.Value) error
	decode(goValue reflect.Value) (protoreflect.Value, error)
}

func (b *schema) getValueCodec(descriptor protoreflect.FieldDescriptor) (valueCodec, error) {
	if descriptor.IsList() {
		return nil, fmt.Errorf("TODO")
	}

	if descriptor.IsMap() {
		return nil, fmt.Errorf("TODO")
	}

	switch descriptor.Kind() {
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return int32Codec{}, nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return uint32Codec{}, nil
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return int64Codec{}, nil
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return nil, fmt.Errorf("TODO")
	case protoreflect.BoolKind:
		return nil, fmt.Errorf("TODO")
	case protoreflect.EnumKind:
		return nil, fmt.Errorf("TODO")
	case protoreflect.FloatKind:
		return nil, fmt.Errorf("TODO")
	case protoreflect.DoubleKind:
		return nil, fmt.Errorf("TODO")
	case protoreflect.BytesKind:
		return nil, fmt.Errorf("TODO")
	case protoreflect.StringKind:
		return stringCodec{}, nil
	case protoreflect.MessageKind:
		switch descriptor.Message().FullName() {
		case timestampDesc.FullName():
			return nil, fmt.Errorf("TODO")
		case durationDesc.FullName():
			return nil, fmt.Errorf("TODO")
		default:
			typ, err := b.resolver.FindMessageByName(descriptor.Message().FullName())
			if err != nil {
				return nil, err
			}
			return newMessageValueCodec(b, typ), nil
		}
	default:
		panic("TODO")
	}
}

type messageValueCodec struct {
	*schema
	protoType protoreflect.MessageType
}

func newMessageValueCodec(builder *schema, protoType protoreflect.MessageType) *messageValueCodec {
	return &messageValueCodec{schema: builder, protoType: protoType}
}

func (m messageValueCodec) goType() reflect.Type {
	return reflect.TypeOf(datatypes.JSON{})
}

func (m messageValueCodec) encode(protoValue protoreflect.Value, goValue reflect.Value) error {
	bz, err := m.jsonMarshalOptions.Marshal(protoValue.Message().Interface())
	if err != nil {
		return err
	}

	goValue.Set(reflect.ValueOf(datatypes.JSON(bz)))
	return nil
}

func (m messageValueCodec) decode(goValue reflect.Value) (protoreflect.Value, error) {
	j := goValue.Interface().(datatypes.JSON)
	msg := m.protoType.New()
	err := m.jsonUnmarshalOptions.Unmarshal(j, msg.Interface())
	if err != nil {
		return protoreflect.Value{}, err
	}

	return protoreflect.ValueOfMessage(msg), nil
}
