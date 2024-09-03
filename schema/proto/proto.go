package schemaproto

import (
	"fmt"

	cosmosproto "github.com/cosmos/cosmos-proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/schema"
)

func FieldDescriptorToField(descriptor protoreflect.FieldDescriptor) (schema.Field, error) {
	kind, err := kindToDefaultKind(descriptor.Kind())
	if err != nil {
		return schema.Field{}, err
	}
	field := schema.Field{
		Name:             string(descriptor.Name()),
		Kind:             kind,
		ProtoFieldNumber: int32(descriptor.Number()),
	}

	scalarType := proto.GetExtension(descriptor.Options(), cosmosproto.E_Scalar).(string)
	switch scalarType {
	case "cosmos.AddressString":
		if kind != schema.StringKind {
			return schema.Field{}, fmt.Errorf("scalar type %q is incompatible with kind %q", scalarType, kind)
		}

		field.ProtoType = "string"
		field.Kind = schema.AddressKind
	case "cosmos.AddressBytes":
		if kind != schema.BytesKind {
			return schema.Field{}, fmt.Errorf("scalar type %q is incompatible with kind %q", scalarType, kind)
		}

		field.ProtoType = "bytes"
		field.Kind = schema.AddressKind
	case "cosmos.Int":
		if kind != schema.StringKind {
			return schema.Field{}, fmt.Errorf("scalar type %q is incompatible with kind %q", scalarType, kind)
		}

		field.Kind = schema.IntegerStringKind
	case "cosmos.Dec":
		if kind != schema.StringKind {
			return schema.Field{}, fmt.Errorf("scalar type %q is incompatible with kind %q", scalarType, kind)
		}

		field.Kind = schema.DecimalStringKind
	default:
		if scalarType != "" {
			return schema.Field{}, fmt.Errorf("unknown scalar type %q", scalarType)
		}
	}

	switch descriptor.Kind() {
	case protoreflect.MessageKind:
		field.ProtoType = string(descriptor.Message().FullName())
	case protoreflect.EnumKind:
		field.ReferencedType = string(descriptor.Enum().FullName())
		// TODO create enum reference type
	default:
	}

	if descriptor.IsList() {
		field.Kind = schema.JSONKind
		field.ProtoType = "repeated"
	}

	return field, nil
}

func kindToDefaultKind(kind protoreflect.Kind) (schema.Kind, error) {
	switch kind {
	case protoreflect.BoolKind:
		return schema.BoolKind, nil
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return schema.Int32Kind, nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return schema.Uint32Kind, nil
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return schema.Int64Kind, nil
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return schema.Uint64Kind, nil
	case protoreflect.StringKind:
		return schema.StringKind, nil
	case protoreflect.BytesKind:
		return schema.BytesKind, nil
	case protoreflect.MessageKind:
		return schema.JSONKind, nil
	case protoreflect.EnumKind:
		return schema.EnumKind, nil
	case protoreflect.FloatKind:
		return schema.Float32Kind, nil
	case protoreflect.DoubleKind:
		return schema.Float64Kind, nil
	case protoreflect.GroupKind:
		return schema.InvalidKind, fmt.Errorf("unsupported kind %v", kind)
	default:
		return schema.InvalidKind, fmt.Errorf("unknown kind %v", kind)
	}
}
