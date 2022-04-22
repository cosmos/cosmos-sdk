package flag

import (
	"context"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type enumType struct {
	enum protoreflect.EnumDescriptor
}

func (b enumType) NewValue(ctx context.Context, builder *Options) SimpleValue {
	return &enumValue{
		enum: b.enum,
	}
}

func (b enumType) DefaultValue() string {
	defValue := ""
	if def := b.enum.Values().ByNumber(0); def != nil {
		defValue = enumValueName(b.enum, def)
	}
	return defValue
}

type enumValue struct {
	enum  protoreflect.EnumDescriptor
	value protoreflect.EnumNumber
}

func (e enumValue) Get() protoreflect.Value {
	return protoreflect.ValueOfEnum(e.value)
}

func enumValueName(enum protoreflect.EnumDescriptor, enumValue protoreflect.EnumValueDescriptor) string {
	name := string(enumValue.Name())
	name = strings.TrimPrefix(name, strcase.ToScreamingSnake(string(enum.Name()))+"_")
	return strcase.ToKebab(name)
}

func (e enumValue) String() string {
	return enumValueName(e.enum, e.enum.Values().ByNumber(e.value))
}

func (e *enumValue) Set(s string) error {
	valDesc := e.enum.Values().ByName(protoreflect.Name(s))
	if valDesc == nil {
		return fmt.Errorf("%s is not a valid value for enum %s", s, e.enum.FullName())
	}
	e.value = valDesc.Number()
	return nil
}

func (e enumValue) Type() string {
	var vals []string
	n := e.enum.Values().Len()
	for i := 0; i < n; i++ {
		vals = append(vals, enumValueName(e.enum, e.enum.Values().Get(i)))
	}
	return fmt.Sprintf("%s (%s)", e.enum.Name(), strings.Join(vals, " | "))
}
