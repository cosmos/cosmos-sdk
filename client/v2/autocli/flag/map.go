package flag

import (
	"context"
	"fmt"
	"github.com/cockroachdb/errors"
	"google.golang.org/protobuf/reflect/protoreflect"
	"strings"
)

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

type mapValue[K comparable, T any] struct {
	value               *map[K]T
	toProtoreflectValue func(T) protoreflect.Value
}

func newMapValue[K comparable, T any](mapV *map[K]T, toProtoreflectValue func(T) protoreflect.Value) mapValue[K, T] {
	return mapValue[K, T]{value: mapV, toProtoreflectValue: toProtoreflectValue}
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
