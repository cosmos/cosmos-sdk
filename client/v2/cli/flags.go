package cli

import (
	"context"

	cosmos_proto "github.com/cosmos/cosmos-proto"
	"github.com/iancoleman/strcase"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

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
	case protoreflect.StringKind:
		return stringFlagType{}
	case protoreflect.Uint32Kind:
	case protoreflect.Fixed32Kind:
	case protoreflect.Uint64Kind:
	case protoreflect.Fixed64Kind:
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
	case protoreflect.BoolKind:
		return boolFlagType{}
	case protoreflect.EnumKind:
	case protoreflect.MessageKind:
		b.init()
		if flagType, ok := b.messageFlagTypes[field.Message().FullName()]; ok {
			return flagType
		}
	default:
	}

	//fmt.Printf("TODO: %v\n", field)
	return nil
}

type flagValueClosure func() protoreflect.Value

func (f flagValueClosure) Get() protoreflect.Value {
	return f()
}

type stringFlagType struct{}

func (s stringFlagType) AddFlag(ctx context.Context, set *pflag.FlagSet, field protoreflect.FieldDescriptor) FlagValue {
	val := set.String(descriptorKebabName(field), "", descriptorDocs(field))
	return flagValueClosure(func() protoreflect.Value {
		return protoreflect.ValueOfString(*val)
	})
}

type boolFlagType struct{}

func (s boolFlagType) AddFlag(ctx context.Context, set *pflag.FlagSet, field protoreflect.FieldDescriptor) FlagValue {
	val := set.Bool(descriptorKebabName(field), false, descriptorDocs(field))
	return flagValueClosure(func() protoreflect.Value {
		return protoreflect.ValueOfBool(*val)
	})
}

func descriptorKebabName(descriptor protoreflect.Descriptor) string {
	return strcase.ToKebab(string(descriptor.Name()))
}

func descriptorDocs(descriptor protoreflect.Descriptor) string {
	return descriptor.ParentFile().SourceLocations().ByDescriptor(descriptor).LeadingComments
}
