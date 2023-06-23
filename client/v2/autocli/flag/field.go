package flag

import (
	"context"
	"strconv"

	cosmos_proto "github.com/cosmos/cosmos-proto"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
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
	} else if field.IsMap() {
		keyKind := field.MapKey().Kind()
		valKind := field.MapValue().Kind()
		val = bindSimpleMapFlag(flagSet, keyKind, valKind, name, shorthand, usage)
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
	if field.IsMap() {
		keyKind := field.MapKey().Kind()
		valType := b.resolveFlagType(field.MapValue())
		if valType != nil {
			switch keyKind {
			case protoreflect.StringKind:
				ct := new(compositeMapType[string])
				ct.keyValueResolver = func(s string) (string, error) { return s, nil }
				ct.valueType = valType
				ct.keyType = "string"
				return ct
			case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
				ct := new(compositeMapType[int32])
				ct.keyValueResolver = func(s string) (int32, error) {
					i, err := strconv.ParseInt(s, 10, 32)
					return int32(i), err
				}
				ct.valueType = valType
				ct.keyType = "int32"
				return ct
			case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
				ct := new(compositeMapType[int64])
				ct.keyValueResolver = func(s string) (int64, error) {
					i, err := strconv.ParseInt(s, 10, 64)
					return i, err
				}
				ct.valueType = valType
				ct.keyType = "int64"
				return ct
			case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
				ct := new(compositeMapType[uint32])
				ct.keyValueResolver = func(s string) (uint32, error) {
					i, err := strconv.ParseUint(s, 10, 32)
					return uint32(i), err
				}
				ct.valueType = valType
				ct.keyType = "uint32"
				return ct
			case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
				ct := new(compositeMapType[uint64])
				ct.keyValueResolver = func(s string) (uint64, error) {
					i, err := strconv.ParseUint(s, 10, 64)
					return i, err
				}
				ct.valueType = valType
				ct.keyType = "uint64"
				return ct
			case protoreflect.BoolKind:
				ct := new(compositeMapType[bool])
				ct.keyValueResolver = strconv.ParseBool
				ct.valueType = valType
				ct.keyType = "bool"
				return ct
			}
			return nil

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
