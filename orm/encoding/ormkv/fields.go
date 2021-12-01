package ormkv

import (
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// Fields abstractly represents a list of fields with a comparable type which
// can be used as a map kep.
type Fields struct {
	fields string
}

// CommaSeparatedFields creates a Fields instance from a list of comma-separated
// fields.
func CommaSeparatedFields(fields string) (Fields, error) {
	// normalize cases where there are spaces
	if strings.IndexByte(fields, ' ') >= 0 {
		parts := strings.Split(fields, ",")
		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}
		fields = strings.Join(parts, ",")
	}
	return Fields{fields: fields}, nil
}

// FieldsFromDescriptors creates a Fields instance from an array of field
// descriptors.
func FieldsFromDescriptors(fieldDescriptors []protoreflect.FieldDescriptor) Fields {
	names := make([]protoreflect.Name, len(fieldDescriptors))
	for i, descriptor := range fieldDescriptors {
		names[i] = descriptor.Name()
	}
	return FieldsFromNames(names)
}

// FieldsFromNames creates a Fields instance from an array of field
// names.
func FieldsFromNames(fieldNames []protoreflect.Name) Fields {
	var names []string
	for _, name := range fieldNames {
		names = append(names, string(name))
	}
	return Fields{fields: strings.Join(names, ",")}
}

// Names returns the array of names this Fields instance represents.
func (f Fields) Names() []protoreflect.Name {
	if f.fields == "" {
		return nil
	}

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
