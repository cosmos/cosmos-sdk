package ormkv

import (
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type Fields struct {
	fields string
}

func CommaSeparatedFields(fields string) (Fields, error) {
	return Fields{fields: fields}, nil
}

func FieldsFromNames(fieldNames []protoreflect.Name) Fields {
	var names []string
	for _, name := range fieldNames {
		names = append(names, string(name))
	}
	return Fields{fields: strings.Join(names, ",")}
}

func (f Fields) Names() []protoreflect.Name {
	fields := strings.Split(f.fields, ",")
	names := make([]protoreflect.Name, len(fields))
	for i, field := range fields {
		names[i] = protoreflect.Name(field)
	}
	return names
}

func (f Fields) String() string {
	return f.fields
}
