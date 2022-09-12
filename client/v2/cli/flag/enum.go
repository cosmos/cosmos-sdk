package flag

import (
	"context"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func enumType(enum protoreflect.EnumDescriptor) Type {
	defValue := ""
	if def := enum.Values().ByNumber(0); def != nil {
		defValue = enumValueName(enum, def)
	}

	return Type{
		NewValue: func(ctx context.Context, builder *Builder) Value {
			val := &enumValue{
				enum:   enum,
				valMap: map[string]protoreflect.EnumValueDescriptor{},
			}
			n := enum.Values().Len()
			for i := 0; i < n; i++ {
				valDesc := enum.Values().Get(i)
				val.valMap[enumValueName(enum, valDesc)] = valDesc
			}
			return val
		},
		DefaultValue: defValue,
	}
}

type enumValue struct {
	enum   protoreflect.EnumDescriptor
	value  protoreflect.EnumNumber
	valMap map[string]protoreflect.EnumValueDescriptor
}

func (e enumValue) Bind(message protoreflect.Message, field protoreflect.FieldDescriptor) {
	message.Set(field, protoreflect.ValueOfEnum(e.value))
}

func (e enumValue) Get() (protoreflect.Value, error) {
	return protoreflect.ValueOfEnum(e.value), nil
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
	valDesc, ok := e.valMap[s]
	if !ok {
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
