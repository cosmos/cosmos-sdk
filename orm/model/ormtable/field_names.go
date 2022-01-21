package ormtable

import (
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// fieldNames abstractly represents a list of fields with a comparable type which
// can be used as a map key. It is used primarily to lookup indexes.
type fieldNames struct {
	fields string
}

// commaSeparatedFieldNames creates a fieldNames instance from a list of comma-separated
// fields.
func commaSeparatedFieldNames(fields string) fieldNames {
	// normalize cases where there are spaces
	if strings.IndexByte(fields, ' ') >= 0 {
		parts := strings.Split(fields, ",")
		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}
		fields = strings.Join(parts, ",")
	}
	return fieldNames{fields: fields}
}

// fieldsFromNames creates a fieldNames instance from an array of field
// names.
func fieldsFromNames(fnames []protoreflect.Name) fieldNames {
	var names []string
	for _, name := range fnames {
		names = append(names, string(name))
	}
	return fieldNames{fields: strings.Join(names, ",")}
}

// Names returns the array of names this fieldNames instance represents.
func (f fieldNames) Names() []protoreflect.Name {
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

func (f fieldNames) String() string {
	return f.fields
}
