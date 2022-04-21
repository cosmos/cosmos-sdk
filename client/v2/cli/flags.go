package cli

import (
	"github.com/iancoleman/strcase"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (b *Builder) addFieldFlag(flags *pflag.FlagSet, field protoreflect.FieldDescriptor) {
	name := strcase.ToKebab(string(field.Name()))
	docs := field.ParentFile().SourceLocations().ByDescriptor(field).LeadingComments
	flags.String(name, "", docs)

	if field.IsList() {
	}

	if field.ContainingOneof() != nil {
	}

	if field.HasOptionalKeyword() {
	}

	switch field.Kind() {
	case protoreflect.BytesKind:
	case protoreflect.StringKind:
	case protoreflect.Uint32Kind:
	case protoreflect.Fixed32Kind:
	case protoreflect.Uint64Kind:
	case protoreflect.Fixed64Kind:
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
	case protoreflect.BoolKind:
	case protoreflect.EnumKind:
	case protoreflect.MessageKind:
	default:
	}
}
