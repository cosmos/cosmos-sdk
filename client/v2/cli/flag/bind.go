package flag

import (
	"context"

	cosmos_proto "github.com/cosmos/cosmos-proto"
	"github.com/iancoleman/strcase"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Binder interface {
	Bind(message protoreflect.Message, field protoreflect.FieldDescriptor)
}

type funcBinder func(message protoreflect.Message, field protoreflect.FieldDescriptor)

func (f funcBinder) Bind(message protoreflect.Message, field protoreflect.FieldDescriptor) {
	f(message, field)
}

func (b *Options) BindFieldFlag(ctx context.Context, flagSet *pflag.FlagSet, field protoreflect.FieldDescriptor) Binder {
	name := descriptorKebabName(field)
	usage := descriptorDocs(field)
	shorthand := ""

	if typ := b.resolveFlagType(field); typ != nil {
		switch typ := typ.(type) {
		case Type:
			val := typ.NewValue(ctx, b)
			flagSet.AddFlag(&pflag.Flag{
				Name:      name,
				Shorthand: shorthand,
				Usage:     usage,
				DefValue:  typ.DefaultValue(),
				Value:     val,
			})
			return funcBinder(func(message protoreflect.Message, field protoreflect.FieldDescriptor) {
				message.Set(field, val.Get())
			})
		}
	}

	if field.IsList() {
		if value := bindSimpleListFlag(flagSet, field.Kind(), name, shorthand, usage); value != nil {
			return funcBinder(func(message protoreflect.Message, field protoreflect.FieldDescriptor) {
				value.AppendTo(message.NewField(field).List())
			})
		}
		return nil
	}

	if value := bindSimpleFlag(flagSet, field.Kind(), name, shorthand, usage); value != nil {
		return funcBinder(func(message protoreflect.Message, field protoreflect.FieldDescriptor) {
			message.Set(field, value.Get())
		})
	}

	return nil
}

func (b *Options) resolveFlagType(field protoreflect.FieldDescriptor) interface{} {
	typ := b.resolveFlagTypeBasic(field)
	if !field.IsList() {
		return typ
	}

	if typ != nil {
		//if simple, ok Value:= typ.(Type); ok {
		//return compositeListType{simpleType: simple}
		//}
		panic("TODO")
	}

	return nil
}

func (b *Options) resolveFlagTypeBasic(field protoreflect.FieldDescriptor) interface{} {
	scalar := proto.GetExtension(field.Options(), cosmos_proto.E_Scalar)
	if scalar != nil {
		b.init()
		if typ, ok := b.scalarFlagTypes[scalar.(string)]; ok {
			return typ
		}
	}

	switch field.Kind() {
	case protoreflect.EnumKind:
		return enumType{enum: field.Enum()}
	case protoreflect.MessageKind:
		b.init()
		if flagType, ok := b.messageFlagTypes[field.Message().FullName()]; ok {
			return flagType
		}
		return jsonMessageFlagType{
			messageDesc: field.Message(),
		}
	default:
		return nil
	}
}

func descriptorKebabName(descriptor protoreflect.Descriptor) string {
	return strcase.ToKebab(string(descriptor.Name()))
}

func descriptorDocs(descriptor protoreflect.Descriptor) string {
	return descriptor.ParentFile().SourceLocations().ByDescriptor(descriptor).LeadingComments
}
