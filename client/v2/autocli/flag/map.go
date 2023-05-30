package flag

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func bindSimpleMapFlag(flagSet *pflag.FlagSet, keyKind, valueKind protoreflect.Kind, name, shorthand, usage string) HasValue {
	switch keyKind {
	case protoreflect.StringKind:
		switch valueKind {
		case protoreflect.StringKind:
			val := flagSet.StringToStringP(name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfString)
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
			val := StringToInt32P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfInt32)
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			val := flagSet.StringToInt64P(name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfInt64)
		case protoreflect.Uint32Kind:
			val := StringToUint32P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfUint32)
		case protoreflect.Uint64Kind:
			val := StringToUint64P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfUint64)
		case protoreflect.BoolKind:
			val := StringToBoolP(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfBool)
		}

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		switch valueKind {
		case protoreflect.StringKind:
			val := Int32ToStringP(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfString)
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
			val := Int32ToInt32P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfInt32)
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			val := Int32ToInt64P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfInt64)
		case protoreflect.Uint32Kind:
			val := Int32ToUint32P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfUint32)
		case protoreflect.Uint64Kind:
			val := Int32ToUint64P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfUint64)
		case protoreflect.BoolKind:
			val := Int32ToBoolP(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfBool)
		}

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		switch valueKind {
		case protoreflect.StringKind:
			val := Int64ToStringP(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfString)
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
			val := Int64ToInt32P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfInt32)
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			val := Int64ToInt64P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfInt64)
		case protoreflect.Uint32Kind:
			val := Int64ToUint32P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfUint32)
		case protoreflect.Uint64Kind:
			val := Int64ToUint64P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfUint64)
		case protoreflect.BoolKind:
			val := Int64ToBoolP(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfBool)
		}
	case protoreflect.Uint32Kind:
		switch valueKind {
		case protoreflect.StringKind:
			val := Uint32ToStringP(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfString)
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
			val := Uint32ToInt32P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfInt32)
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			val := Uint32ToInt64P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfInt64)
		case protoreflect.Uint32Kind:
			val := Uint32ToUint32P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfUint32)
		case protoreflect.Uint64Kind:
			val := Uint32ToUint64P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfUint64)
		case protoreflect.BoolKind:
			val := Uint32ToBoolP(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfBool)
		}
	case protoreflect.Uint64Kind:
		switch valueKind {
		case protoreflect.StringKind:
			val := Uint64ToStringP(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfString)
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
			val := Uint64ToInt32P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfInt32)
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			val := Uint64ToInt64P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfInt64)
		case protoreflect.Uint32Kind:
			val := Uint64ToUint32P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfUint32)
		case protoreflect.Uint64Kind:
			val := Uint64ToUint64P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfUint64)
		case protoreflect.BoolKind:
			val := Uint64ToBoolP(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfBool)
		}
	case protoreflect.BoolKind:
		switch valueKind {
		case protoreflect.StringKind:
			val := BoolToStringP(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfString)
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
			val := BoolToInt32P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfInt32)
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			val := BoolToInt64P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfInt64)
		case protoreflect.Uint32Kind:
			val := BoolToUint32P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfUint32)
		case protoreflect.Uint64Kind:
			val := BoolToUint64P(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfUint64)
		case protoreflect.BoolKind:
			val := BoolToBoolP(flagSet, name, shorthand, nil, usage)
			return newMapValue(val, protoreflect.ValueOfBool)
		}

	}
	return nil
}

type mapValue[K comparable, V any] struct {
	value               *map[K]V
	toProtoreflectValue func(V) protoreflect.Value
}

func newMapValue[K comparable, V any](mapV *map[K]V, toProtoreflectValue func(V) protoreflect.Value) mapValue[K, V] {
	return mapValue[K, V]{value: mapV, toProtoreflectValue: toProtoreflectValue}
}

func (v mapValue[K, V]) Get(mutable protoreflect.Value) (protoreflect.Value, error) {
	protoMap := mutable.Map()
	for k, val := range *v.value {
		protoMap.Set(protoreflect.MapKey(protoreflect.ValueOf(k)), v.toProtoreflectValue(val))
	}
	return mutable, nil
}

// keyValueResolver is a function that converts a string to a key that is primitive Type T
type keyValueResolver[T comparable] func(string) (T, error)

// compositeMapType is a map type that is composed of a key and value type that are both primitive types
type compositeMapType[T comparable] struct {
	keyValueResolver keyValueResolver[T]
	keyType          string
	valueType        Type
}

// compositeMapValue is a map value that is composed of a key and value type that are both primitive types
type compositeMapValue[T comparable] struct {
	keyValueResolver keyValueResolver[T]
	keyType          string
	valueType        Type
	values           map[T]protoreflect.Value
	ctx              context.Context
	opts             *Builder
}

func (m compositeMapType[T]) DefaultValue() string {
	return ""
}

func (m compositeMapType[T]) NewValue(ctx context.Context, opts *Builder) Value {
	return &compositeMapValue[T]{
		keyValueResolver: m.keyValueResolver,
		valueType:        m.valueType,
		keyType:          m.keyType,
		ctx:              ctx,
		opts:             opts,
		values:           nil,
	}
}

func (m *compositeMapValue[T]) Set(s string) error {
	comaArgs := strings.Split(s, ",")
	for _, arg := range comaArgs {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return errors.New("invalid format, expected key=value")
		}
		key, val := parts[0], parts[1]

		keyValue, err := m.keyValueResolver(key)
		if err != nil {
			return err
		}

		simpleVal := m.valueType.NewValue(m.ctx, m.opts)
		err = simpleVal.Set(val)
		if err != nil {
			return err
		}
		protoValue, err := simpleVal.Get(protoreflect.Value{})
		if err != nil {
			return err
		}
		if m.values == nil {
			m.values = make(map[T]protoreflect.Value)
		}

		m.values[keyValue] = protoValue
	}

	return nil
}

func (m *compositeMapValue[T]) Get(mutable protoreflect.Value) (protoreflect.Value, error) {
	protoMap := mutable.Map()
	for key, value := range m.values {
		keyVal := protoreflect.ValueOf(key)
		protoMap.Set(keyVal.MapKey(), value)
	}
	return protoreflect.ValueOfMap(protoMap), nil
}

func (m *compositeMapValue[T]) String() string {
	if m.values == nil {
		return ""
	}

	return fmt.Sprintf("%+v", m.values)
}

func (m *compositeMapValue[T]) Type() string {
	return fmt.Sprintf("map[%s]%s", m.keyType, m.valueType.NewValue(m.ctx, m.opts).Type())
}
