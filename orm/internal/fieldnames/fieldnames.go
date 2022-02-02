package fieldnames

import (
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// FieldNames abstractly represents a list of fields with a comparable type which
// can be used as a map key. It is used primarily to lookup indexes.
type FieldNames struct {
	fields string
}

// CommaSeparatedFieldNames creates a FieldNames instance from a list of comma-separated
// fields.
func CommaSeparatedFieldNames(fields string) FieldNames {
	// normalize cases where there are spaces
	if strings.IndexByte(fields, ' ') >= 0 {
		parts := strings.Split(fields, ",")
		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}
		fields = strings.Join(parts, ",")
	}
	return FieldNames{fields: fields}
}

// FieldsFromNames creates a FieldNames instance from an array of field
// names.
func FieldsFromNames(fnames []protoreflect.Name) FieldNames {
	var names []string
	for _, name := range fnames {
		names = append(names, string(name))
	}
	return FieldNames{fields: strings.Join(names, ",")}
}

// Names returns the array of names this FieldNames instance represents.
func (f FieldNames) Names() []protoreflect.Name {
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

func (f FieldNames) String() string {
	return f.fields
}
