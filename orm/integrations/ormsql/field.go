package ormsql

import (
	"fmt"
	"reflect"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type fieldCodec struct {
	valueCodec
	structField reflect.StructField
	protoField  protoreflect.FieldDescriptor
	name        string
}

func (b *schema) makeFieldCodec(descriptor protoreflect.FieldDescriptor, isPrimaryKey bool) (*fieldCodec, error) {
	valCdc, err := b.getValueCodec(descriptor)
	if err != nil {
		return nil, err
	}

	name := descriptor.Name()
	tag := fmt.Sprintf(`gorm:"column:%s`, name)
	if isPrimaryKey {
		tag = tag + fmt.Sprintf(`;primaryKey;autoIncrement:false`)
	}
	var fieldName = strings.ToTitle(string(name))
	structField := reflect.StructField{
		Name: fieldName,
		Type: valCdc.goType(),
		Tag:  reflect.StructTag(tag + `"`),
	}

	return &fieldCodec{
		valueCodec:  valCdc,
		structField: structField,
		protoField:  descriptor,
		name:        string(name),
	}, nil
}

func (f fieldCodec) encode(message protoreflect.Message, structValue reflect.Value) error {
	if !message.Has(f.protoField) {
		return nil
	}

	protoVal := message.Get(f.protoField)
	goField := structValue.FieldByName(f.structField.Name)
	return f.valueCodec.encode(protoVal, goField)
}

func (f fieldCodec) decode(structValue reflect.Value, message protoreflect.Message) error {
	goField := structValue.FieldByName(f.structField.Name)
	protoVal, err := f.valueCodec.decode(goField)
	if err != nil {
		return err
	}
	message.Set(f.protoField, protoVal)
	return nil
}
