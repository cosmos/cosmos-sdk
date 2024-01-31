package flag

import (
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func bindSimpleFlag(flagSet *pflag.FlagSet, kind protoreflect.Kind, name, shorthand, usage string) SimpleValue {
	switch kind {
	case protoreflect.BytesKind:
		val := flagSet.BytesBase64P(name, shorthand, nil, usage)
		return simpleValue(func() protoreflect.Value {
			return protoreflect.ValueOfBytes(*val)
		})
	case protoreflect.StringKind:
		val := flagSet.StringP(name, shorthand, "", usage)
		return simpleValue(func() protoreflect.Value {
			return protoreflect.ValueOfString(*val)
		})
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		val := flagSet.Uint32P(name, shorthand, 0, usage)
		return simpleValue(func() protoreflect.Value {
			return protoreflect.ValueOfUint32(*val)
		})
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		val := flagSet.Uint64P(name, shorthand, 0, usage)
		return simpleValue(func() protoreflect.Value {
			return protoreflect.ValueOfUint64(*val)
		})
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		val := flagSet.Int32P(name, shorthand, 0, usage)
		return simpleValue(func() protoreflect.Value {
			return protoreflect.ValueOfInt32(*val)
		})
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		val := flagSet.Int64P(name, shorthand, 0, usage)
		return simpleValue(func() protoreflect.Value {
			return protoreflect.ValueOfInt64(*val)
		})
	case protoreflect.BoolKind:
		val := flagSet.BoolP(name, shorthand, false, usage)
		return simpleValue(func() protoreflect.Value {
			return protoreflect.ValueOfBool(*val)
		})
	default:
		return nil
	}
}

type simpleValue func() protoreflect.Value

func (f simpleValue) Get() protoreflect.Value {
	return f()
}
