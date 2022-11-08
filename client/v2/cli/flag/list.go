package flag

import (
	"context"
	"fmt"

	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func bindSimpleListFlag(flagSet *pflag.FlagSet, kind protoreflect.Kind, name, shorthand, usage string) ListValue {
	switch kind {
	case protoreflect.StringKind:
		val := flagSet.StringSliceP(name, shorthand, nil, usage)
		return listValue(func(list protoreflect.List) {
			for _, x := range *val {
				list.Append(protoreflect.ValueOfString(x))
			}
		})
	case protoreflect.BytesKind:
		// TODO
		return nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		val := flagSet.UintSliceP(name, shorthand, nil, usage)
		return listValue(func(list protoreflect.List) {
			for _, x := range *val {
				list.Append(protoreflect.ValueOfUint64(uint64(x)))
			}
		})
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		val := flagSet.IntSliceP(name, shorthand, nil, usage)
		return listValue(func(list protoreflect.List) {
			for _, x := range *val {
				list.Append(protoreflect.ValueOfInt64(int64(x)))
			}
		})
	case protoreflect.BoolKind:
		val := flagSet.BoolSliceP(name, shorthand, nil, usage)
		return listValue(func(list protoreflect.List) {
			for _, x := range *val {
				list.Append(protoreflect.ValueOfBool(x))
			}
		})
	default:
		return nil
	}
}

type listValue func(protoreflect.List)

func (f listValue) AppendTo(list protoreflect.List) {
	f(list)
}

type compositeListType struct {
	simpleType Type
}

func (t compositeListType) NewValue(ctx context.Context, opts *Builder) pflag.Value {
	return &compositeListValue{
		simpleType: t.simpleType,
		values:     nil,
		ctx:        ctx,
		opts:       opts,
	}
}

func (t compositeListType) DefaultValue() string {
	return ""
}

type compositeListValue struct {
	simpleType Type
	values     []protoreflect.Value
	ctx        context.Context
	opts       *Builder
}

func (c compositeListValue) AppendTo(list protoreflect.List) {
	for _, value := range c.values {
		list.Append(value)
	}
}

func (c compositeListValue) String() string {
	if len(c.values) == 0 {
		return ""
	}

	return fmt.Sprintf("%+v", c.values)
}

func (c *compositeListValue) Set(val string) error {
	simpleVal := c.simpleType.NewValue(c.ctx, c.opts)
	err := simpleVal.Set(val)
	if err != nil {
		return err
	}
	c.values = append(c.values, simpleVal.(SimpleValue).Get())
	return nil
}

func (c compositeListValue) Type() string {
	return fmt.Sprintf("%s (repeated)", c.simpleType.NewValue(c.ctx, c.opts).Type())
}
