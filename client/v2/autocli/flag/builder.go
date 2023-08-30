package flag

import (
	"context"
	"fmt"
	"strconv"

	cosmos_proto "github.com/cosmos/cosmos-proto"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli/keyring"
	"cosmossdk.io/client/v2/internal/util"
	"cosmossdk.io/core/address"
)

// Builder manages options for building pflag flags for protobuf messages.
type Builder struct {
	// TypeResolver specifies how protobuf types will be resolved. If it is
	// nil protoregistry.GlobalTypes will be used.
	TypeResolver interface {
		protoregistry.MessageTypeResolver
		protoregistry.ExtensionTypeResolver
	}

	// FileResolver specifies how protobuf file descriptors will be resolved. If it is
	// nil protoregistry.GlobalFiles will be used.
	FileResolver protodesc.Resolver

	messageFlagTypes map[protoreflect.FullName]Type
	scalarFlagTypes  map[string]Type

	// AddressCodec is the address codec used for the address flag
	AddressCodec          address.Codec
	ValidatorAddressCodec address.Codec
	ConsensusAddressCodec address.Codec

	// Keyring implementation
	Keyring keyring.Keyring
}

func (b *Builder) init() {
	if b.messageFlagTypes == nil {
		b.messageFlagTypes = map[protoreflect.FullName]Type{}
		b.messageFlagTypes["google.protobuf.Timestamp"] = timestampType{}
		b.messageFlagTypes["google.protobuf.Duration"] = durationType{}
		b.messageFlagTypes["cosmos.base.v1beta1.Coin"] = coinType{}
	}

	if b.scalarFlagTypes == nil {
		b.scalarFlagTypes = map[string]Type{}
		b.scalarFlagTypes["cosmos.AddressString"] = addressStringType{}
		b.scalarFlagTypes["cosmos.ValidatorAddressString"] = validatorAddressStringType{}
		b.scalarFlagTypes["cosmos.ConsensusAddressString"] = consensusAddressStringType{}
	}
}

func (b *Builder) DefineMessageFlagType(messageName protoreflect.FullName, flagType Type) {
	b.init()
	b.messageFlagTypes[messageName] = flagType
}

func (b *Builder) DefineScalarFlagType(scalarName string, flagType Type) {
	b.init()
	b.scalarFlagTypes[scalarName] = flagType
}

func (b *Builder) AddMessageFlags(ctx context.Context, flagSet *pflag.FlagSet, messageType protoreflect.MessageType, commandOptions *autocliv1.RpcCommandOptions) (*MessageBinder, error) {
	return b.addMessageFlags(ctx, flagSet, messageType, commandOptions, namingOptions{})
}

// AddMessageFlags adds flags for each field in the message to the flag set.
func (b *Builder) addMessageFlags(ctx context.Context, flagSet *pflag.FlagSet, messageType protoreflect.MessageType, commandOptions *autocliv1.RpcCommandOptions, options namingOptions) (*MessageBinder, error) {
	fields := messageType.Descriptor().Fields()
	numFields := fields.Len()
	handler := &MessageBinder{
		messageType: messageType,
	}

	isPositional := map[string]bool{}
	hasVarargs := false
	hasOptional := false
	n := len(commandOptions.PositionalArgs)
	// positional args are also parsed using a FlagSet so that we can reuse all the same parsers
	handler.positionalFlagSet = pflag.NewFlagSet("positional", pflag.ContinueOnError)
	for i, arg := range commandOptions.PositionalArgs {
		isPositional[arg.ProtoField] = true

		field := fields.ByName(protoreflect.Name(arg.ProtoField))
		if field == nil {
			return nil, fmt.Errorf("can't find field %s on %s", arg.ProtoField, messageType.Descriptor().FullName())
		}

		if arg.Optional && arg.Varargs {
			return nil, fmt.Errorf("positional argument %s can't be both optional and varargs", arg.ProtoField)
		}

		if arg.Varargs {
			if i != n-1 {
				return nil, fmt.Errorf("varargs positional argument %s must be the last argument", arg.ProtoField)
			}

			hasVarargs = true
		}

		if arg.Optional {
			if i != n-1 {
				return nil, fmt.Errorf("optional positional argument %s must be the last argument", arg.ProtoField)
			}

			hasOptional = true
		}

		_, hasValue, err := b.addFieldFlag(
			ctx,
			handler.positionalFlagSet,
			field,
			&autocliv1.FlagOptions{Name: fmt.Sprintf("%d", i)},
			namingOptions{},
		)
		if err != nil {
			return nil, err
		}

		handler.positionalArgs = append(handler.positionalArgs, fieldBinding{
			field:    field,
			hasValue: hasValue,
		})
	}

	if hasVarargs {
		handler.CobraArgs = cobra.MinimumNArgs(n - 1)
		handler.hasVarargs = true
	} else if hasOptional {
		handler.CobraArgs = cobra.RangeArgs(n-1, n)
		handler.hasOptional = true
	} else {
		handler.CobraArgs = cobra.ExactArgs(n)
	}

	// validate flag options
	for name := range commandOptions.FlagOptions {
		if fields.ByName(protoreflect.Name(name)) == nil {
			return nil, fmt.Errorf("can't find field %s on %s specified as a flag", name, messageType.Descriptor().FullName())
		}
	}

	flagOptsByFlagName := map[string]*autocliv1.FlagOptions{}
	for i := 0; i < numFields; i++ {
		field := fields.Get(i)
		if isPositional[string(field.Name())] {
			continue
		}

		flagOpts := commandOptions.FlagOptions[string(field.Name())]
		name, hasValue, err := b.addFieldFlag(ctx, flagSet, field, flagOpts, options)
		flagOptsByFlagName[name] = flagOpts
		if err != nil {
			return nil, err
		}

		handler.flagBindings = append(handler.flagBindings, fieldBinding{
			hasValue: hasValue,
			field:    field,
		})
	}

	flagSet.VisitAll(func(flag *pflag.Flag) {
		opts := flagOptsByFlagName[flag.Name]
		if opts != nil {
			// This is a bit of hacking around the pflag API, but
			// we need to set these options here using Flag.VisitAll because the flag
			// constructors that pflag gives us (StringP, Int32P, etc.) do not
			// actually return the *Flag instance
			flag.Deprecated = opts.Deprecated
			flag.ShorthandDeprecated = opts.ShorthandDeprecated
			flag.Hidden = opts.Hidden
		}
	})

	return handler, nil
}

// bindPageRequest create a flag for pagination
func (b *Builder) bindPageRequest(ctx context.Context, flagSet *pflag.FlagSet, field protoreflect.FieldDescriptor) (HasValue, error) {
	return b.addMessageFlags(
		ctx,
		flagSet,
		util.ResolveMessageType(b.TypeResolver, field.Message()),
		&autocliv1.RpcCommandOptions{},
		namingOptions{Prefix: "page-"},
	)
}

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
