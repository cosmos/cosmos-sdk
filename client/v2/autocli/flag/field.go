package flag

import (
	"context"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	cosmos_proto "github.com/cosmos/cosmos-proto"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/client/v2/internal/util"
)

// namingOptions specifies internal naming options for flags.
type namingOptions struct {
	// Prefix is a prefix to prepend to all flags.
	Prefix string
}

// addFieldFlag adds a flag for the provided field to the flag set.
func (b *Builder) addFieldFlag(ctx context.Context, flagSet *pflag.FlagSet, field protoreflect.FieldDescriptor, opts *autocliv1.FlagOptions, options namingOptions) (name string, hasValue HasValue, err error) {
	if opts == nil {
		opts = &autocliv1.FlagOptions{}
	}

	if field.Kind() == protoreflect.MessageKind && field.Message().FullName() == "cosmos.base.query.v1beta1.PageRequest" {
		hasValue, err := b.bindPageRequest(ctx, flagSet, field)
		return "", hasValue, err
	}

	name = opts.Name
	if name == "" {
		name = options.Prefix + util.DescriptorKebabName(field)
	}

	usage := opts.Usage
	if usage == "" {
		usage = util.DescriptorDocs(field)
	}

	shorthand := opts.Shorthand
	defaultValue := opts.DefaultValue

	if typ := b.resolveFlagType(field); typ != nil {
		if defaultValue == "" {
			defaultValue = typ.DefaultValue()
		}

		val := typ.NewValue(ctx, b)
		flagSet.AddFlag(&pflag.Flag{
			Name:      name,
			Shorthand: shorthand,
			Usage:     usage,
			DefValue:  defaultValue,
			Value:     val,
		})
		return name, val, nil
	}

	// use the built-in pflag StringP, Int32P, etc. functions
	var val HasValue
	if field.IsList() {
		val = bindSimpleListFlag(flagSet, field.Kind(), name, shorthand, usage)
	} else {
		val = bindSimpleFlag(flagSet, field.Kind(), name, shorthand, usage)
	}

	// This is a bit of hacking around the pflag API, but the
	// defaultValue is set in this way because this is much easier than trying
	// to parse the string into the types that StringSliceP, Int32P, etc. expect
	if defaultValue != "" {
		err = flagSet.Set(name, defaultValue)
	}
	return name, val, err
}

func (b *Builder) resolveFlagType(field protoreflect.FieldDescriptor) Type {
	typ := b.resolveFlagTypeBasic(field)
	if field.IsList() {
		if typ != nil {
			return compositeListType{simpleType: typ}
		}

		return nil
	}

	return typ
}

func (b *Builder) resolveFlagTypeBasic(field protoreflect.FieldDescriptor) Type {
	scalar := proto.GetExtension(field.Options(), cosmos_proto.E_Scalar)
	if scalar != nil {
		b.init()
		if typ, ok := b.scalarFlagTypes[scalar.(string)]; ok {
			return typ
		}
	}

	switch field.Kind() {
	case protoreflect.BytesKind:
		return binaryType{}
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
