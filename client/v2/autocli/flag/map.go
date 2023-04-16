package flag

import (
	"context"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
	"strings"
)

func bindSimpleMapFlag(flagSet *pflag.FlagSet, keyKind protoreflect.Kind, valueKind protoreflect.Kind, name, shorthand, usage string) HasValue {
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

type compositeMapType struct {
	keyType, valueType Type
}

type compositeMapValue struct {
	keyType   Type
	valueType Type
	values    protoreflect.Map
	ctx       context.Context
	opts      *Builder
}

func (m compositeMapType) DefaultValue() string {
	return ""
}

func (m compositeMapType) NewValue(ctx context.Context, opts *Builder) Value {
	return &compositeMapValue{
		keyType:   m.keyType,
		valueType: m.valueType,
		ctx:       ctx,
		opts:      opts,
		values:    nil,
	}
}

func (m *compositeMapValue) Set(s string) error {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) != 2 {
		return errors.New("invalid format, expected key=value")
	}
	key, val := parts[0], parts[1]

	keySimpleVal := m.keyType.NewValue(m.ctx, m.opts)
	err := keySimpleVal.Set(key)
	if err != nil {
		return err
	}
	keyValue, err := keySimpleVal.Get(protoreflect.Value{})
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

	}
	m.values.Set(keyValue.MapKey(), protoValue)
	return nil
}

func (m *compositeMapValue) Get(mutable protoreflect.Value) (protoreflect.Value, error) {
	protoMap := mutable.Map()
	m.values.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		protoMap.Set(k, v)
		return true
	})

	return protoreflect.ValueOfMap(protoMap), nil
}

func (m *compositeMapValue) String() string {
	if m.values == nil {
		return ""
	}

	return fmt.Sprintf("%+v", m.values)
}

func (m *compositeMapValue) Type() string {
	return fmt.Sprintf("map[%s]%s", m.keyType.NewValue(m.ctx, m.opts).Type(), m.valueType.NewValue(m.ctx, m.opts).Type())
}
