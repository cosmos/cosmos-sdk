package valuerenderer

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type enumValueRenderer struct {
	fd protoreflect.FieldDescriptor
}

func NewEnumValueRenderer(fd protoreflect.FieldDescriptor) ValueRenderer {
	return enumValueRenderer{fd: fd}
}

var _ ValueRenderer = (*enumValueRenderer)(nil)

func (er enumValueRenderer) Format(_ context.Context, v protoreflect.Value) ([]Screen, error) {
	ed := er.fd.Enum()
	if ed == nil {
		return nil, fmt.Errorf("expected enum field, got %T", er.fd)
	}

	evd := ed.Values().ByNumber(v.Enum())
	fullName := string(evd.FullName())

	// Transform the Enum name to SNAKE_CASE, and optionally trim if from the
	// enum value name.
	snakeCaseEd := toSnakeCase(string(ed.Name()))
	if strings.HasPrefix(fullName, snakeCaseEd) {
		fullName = strings.TrimPrefix(fullName, snakeCaseEd+"_")
		fullName = strings.ToLower(fullName)
	}

	return []Screen{{Text: formatFieldName(fullName)}}, nil

}

func (er enumValueRenderer) Parse(_ context.Context, screens []Screen) (protoreflect.Value, error) {
	panic("unimplemented")
}

// toSnakeCase converts from PascalCase to capitalized SNAKE_CASE.
func toSnakeCase(s string) string {
	var buf bytes.Buffer
	for _, c := range s {
		if 'A' <= c && c <= 'Z' {
			// just convert [A-Z] to _[A-Z]
			if buf.Len() > 0 {
				buf.WriteRune('_')
			}
			buf.WriteRune(c)
		} else {
			buf.WriteRune(c - 'a' + 'A')
		}
	}

	return buf.String()
}
