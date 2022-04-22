package cli

import (
	"context"
	"fmt"
	"strings"

	cosmos_proto "github.com/cosmos/cosmos-proto"
	"github.com/iancoleman/strcase"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type flagValueClosure func() protoreflect.Value

func (f flagValueClosure) Get() protoreflect.Value {
	return f()
}

type stringFlagType struct{}

func (s stringFlagType) AddFlag(ctx context.Context, builder *Builder, set *pflag.FlagSet, field protoreflect.FieldDescriptor) FlagValue {
	val := set.String(descriptorKebabName(field), "", descriptorDocs(field))
	return flagValueClosure(func() protoreflect.Value {
		return protoreflect.ValueOfString(*val)
	})
}

type boolFlagType struct{}

func (s boolFlagType) AddFlag(ctx context.Context, builder *Builder, set *pflag.FlagSet, field protoreflect.FieldDescriptor) FlagValue {
	val := set.Bool(descriptorKebabName(field), false, descriptorDocs(field))
	return flagValueClosure(func() protoreflect.Value {
		return protoreflect.ValueOfBool(*val)
	})
}

type uint32FlagType struct{}

func (u uint32FlagType) AddFlag(ctx context.Context, builder *Builder, set *pflag.FlagSet, descriptor protoreflect.FieldDescriptor) FlagValue {
	val := set.Uint32(descriptorKebabName(descriptor), 0, descriptorDocs(descriptor))
	return flagValueClosure(func() protoreflect.Value {
		return protoreflect.ValueOfUint32(*val)
	})
}

type uint64FlagType struct{}

func (u uint64FlagType) AddFlag(ctx context.Context, builder *Builder, set *pflag.FlagSet, descriptor protoreflect.FieldDescriptor) FlagValue {
	val := set.Uint64(descriptorKebabName(descriptor), 0, descriptorDocs(descriptor))
	return flagValueClosure(func() protoreflect.Value {
		return protoreflect.ValueOfUint64(*val)
	})
}

type int32FlagType struct{}

func (i int32FlagType) AddFlag(ctx context.Context, builder *Builder, set *pflag.FlagSet, descriptor protoreflect.FieldDescriptor) FlagValue {
	val := set.Int32(descriptorKebabName(descriptor), 0, descriptorDocs(descriptor))
	return flagValueClosure(func() protoreflect.Value {
		return protoreflect.ValueOfInt32(*val)
	})
}

type int64FlagType struct{}

func (i int64FlagType) AddFlag(ctx context.Context, builder *Builder, set *pflag.FlagSet, descriptor protoreflect.FieldDescriptor) FlagValue {
	val := set.Int64(descriptorKebabName(descriptor), 0, descriptorDocs(descriptor))
	return flagValueClosure(func() protoreflect.Value {
		return protoreflect.ValueOfInt64(*val)
	})
}

type base64BytesFlagType struct{}

func (b base64BytesFlagType) AddFlag(ctx context.Context, builder *Builder, set *pflag.FlagSet, descriptor protoreflect.FieldDescriptor) FlagValue {
	val := set.BytesBase64(descriptorKebabName(descriptor), nil, descriptorDocs(descriptor))
	return flagValueClosure(func() protoreflect.Value {
		return protoreflect.ValueOfBytes(*val)
	})
}

type enumFlagType struct {
	enum protoreflect.EnumDescriptor
}

func (b enumFlagType) AddFlag(ctx context.Context, builder *Builder, set *pflag.FlagSet, descriptor protoreflect.FieldDescriptor) FlagValue {
	val := &enumFlagValue{
		enum: b.enum,
	}
	defValue := ""
	if def := b.enum.Values().ByNumber(0); def != nil {
		defValue = enumValueName(b.enum, def)
	}
	set.AddFlag(&pflag.Flag{
		Name:     descriptorKebabName(descriptor),
		Usage:    descriptorDocs(descriptor),
		Value:    val,
		DefValue: defValue,
	})
	return val
}

type enumFlagValue struct {
	enum  protoreflect.EnumDescriptor
	value protoreflect.EnumNumber
}

func (e enumFlagValue) Get() protoreflect.Value {
	return protoreflect.ValueOfEnum(e.value)
}

func enumValueName(enum protoreflect.EnumDescriptor, enumValue protoreflect.EnumValueDescriptor) string {
	name := string(enumValue.Name())
	name = strings.TrimPrefix(name, strcase.ToScreamingSnake(string(enum.Name()))+"_")
	return strcase.ToKebab(name)
}

func (e enumFlagValue) String() string {
	return enumValueName(e.enum, e.enum.Values().ByNumber(e.value))
}

func (e *enumFlagValue) Set(s string) error {
	valDesc := e.enum.Values().ByName(protoreflect.Name(s))
	if valDesc == nil {
		return fmt.Errorf("%s is not a valid value for enum %s", s, e.enum.FullName())
	}
	e.value = valDesc.Number()
	return nil
}

func (e enumFlagValue) Type() string {
	var vals []string
	n := e.enum.Values().Len()
	for i := 0; i < n; i++ {
		vals = append(vals, enumValueName(e.enum, e.enum.Values().Get(i)))
	}
	return fmt.Sprintf("%s (%s)", e.enum.Name(), strings.Join(vals, " | "))
}

func descriptorKebabName(descriptor protoreflect.Descriptor) string {
	return strcase.ToKebab(string(descriptor.Name()))
}

func descriptorDocs(descriptor protoreflect.Descriptor) string {
	return descriptor.ParentFile().SourceLocations().ByDescriptor(descriptor).LeadingComments
}

type jsonMessageFlagType struct {
	messageDesc protoreflect.MessageDescriptor
}

func (j jsonMessageFlagType) AddFlag(ctx context.Context, builder *Builder, set *pflag.FlagSet, descriptor protoreflect.FieldDescriptor) FlagValue {
	val := &jsonMessageFlagValue{
		messageType:          builder.resolverMessageType(j.messageDesc),
		jsonMarshalOptions:   builder.JSONMarshalOptions,
		jsonUnmarshalOptions: builder.JSONUnmarshalOptions,
	}
	set.AddFlag(&pflag.Flag{
		Name:  descriptorKebabName(descriptor),
		Usage: descriptorDocs(descriptor),
		Value: val,
	})
	return val
}

type jsonMessageFlagValue struct {
	jsonMarshalOptions   protojson.MarshalOptions
	jsonUnmarshalOptions protojson.UnmarshalOptions
	messageType          protoreflect.MessageType
	message              proto.Message
}

func (j jsonMessageFlagValue) Get() protoreflect.Value {
	return protoreflect.ValueOfMessage(j.message.ProtoReflect())
}

func (j jsonMessageFlagValue) String() string {
	if j.message == nil {
		return ""
	}

	bz, err := j.jsonMarshalOptions.Marshal(j.message)
	if err != nil {
		return err.Error()
	}
	return string(bz)
}

func (j *jsonMessageFlagValue) Set(s string) error {
	j.message = j.messageType.New().Interface()
	return j.jsonUnmarshalOptions.Unmarshal([]byte(s), j.message)
}

func (j jsonMessageFlagValue) Type() string {
	return fmt.Sprintf("%s (json string or file)", j.messageType.Descriptor().FullName())
}

func (b *Builder) getFlagType(field protoreflect.FieldDescriptor) FlagType {
	if field.IsList() {

	}

	if field.ContainingOneof() != nil {
	}

	if field.HasOptionalKeyword() {
	}

	scalar := proto.GetExtension(field.Options(), cosmos_proto.E_Scalar)
	if scalar != nil {
		b.init()
		if flagType, ok := b.scalarFlagTypes[scalar.(string)]; ok {
			return flagType
		}
	}

	switch field.Kind() {
	case protoreflect.BytesKind:
		return base64BytesFlagType{}
	case protoreflect.StringKind:
		return stringFlagType{}
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return uint32FlagType{}
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return uint64FlagType{}
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return int32FlagType{}
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return int64FlagType{}
	case protoreflect.BoolKind:
		return boolFlagType{}
	case protoreflect.EnumKind:
		return enumFlagType{
			enum: field.Enum(),
		}
	case protoreflect.MessageKind:
		b.init()
		if flagType, ok := b.messageFlagTypes[field.Message().FullName()]; ok {
			return flagType
		}
		return jsonMessageFlagType{
			messageDesc: field.Message(),
		}
	default:
	}

	return nil
}
