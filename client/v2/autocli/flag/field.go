package flag

import (
	"context"
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	cosmos_proto "github.com/cosmos/cosmos-proto"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/client/v2/internal/util"
)

// FieldValueBinder wraps a flag value in a way that allows it to be bound
// to a particular field in a protobuf message.
type FieldValueBinder interface {
	Bind(message protoreflect.Message, field protoreflect.FieldDescriptor)
}

// Options specifies options for specific flags.
type Options struct {
	// Prefix is a prefix to prepend to all flags.
	Prefix string
}

// AddFieldFlag adds a flag for the provided field to the flag set.
func (b *Builder) AddFieldFlag(ctx context.Context, flagSet *pflag.FlagSet, field protoreflect.FieldDescriptor, opts *autocliv1.FlagOptions, options Options) (FieldValueBinder, error) {
	if field.Kind() == protoreflect.MessageKind && field.Message().FullName() == "cosmos.base.query.v1beta1.PageRequest" {
		return b.bindPageRequest(ctx, flagSet, field)
	}

	name := opts.Name
	if name == "" {
		name = options.Prefix + util.DescriptorKebabName(field)
	}

	usage := opts.Usage
	if usage == "" {
		usage = util.DescriptorDocs(field)
	}

	shorthand := opts.Shorthand

	typ := b.resolveFlagType(field)
	if typ == nil {
		return nil, fmt.Errorf("unable to bind field %v", field)
	}

	defaultValue := opts.DefaultValue
	if defaultValue == "" {
		defaultValue = typ.DefaultValue()
	}

	val := typ.NewValue(ctx, b)
	flagSet.AddFlag(&pflag.Flag{
		Name:                name,
		Shorthand:           shorthand,
		Usage:               usage,
		DefValue:            defaultValue,
		Deprecated:          opts.Deprecated,
		ShorthandDeprecated: opts.ShorthandDeprecated,
		Hidden:              opts.Hidden,
		NoOptDefVal:         opts.NoOptDefaultValue,
		Value:               val,
	})

	return val, nil
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

type simpleValueBinder struct {
	SimpleValue
}

func (s simpleValueBinder) Bind(message protoreflect.Message, field protoreflect.FieldDescriptor) {
	val := s.Get()
	if val.IsValid() {
		message.Set(field, val)
	} else {
		message.Clear(field)
	}
}

type listValueBinder struct {
	ListValue
}

func (s listValueBinder) Bind(message protoreflect.Message, field protoreflect.FieldDescriptor) {
	s.AppendTo(message.NewField(field).List())
}
