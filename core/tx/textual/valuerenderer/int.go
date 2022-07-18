package valuerenderer

import (
	"context"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type intValueRenderer struct{}

var _ ValueRenderer = intValueRenderer{}

func (r intValueRenderer) Format(ctx context.Context, v protoreflect.Value) ([]string, error) {
	formatted, err := formatInteger(v.String())
	if err != nil {
		return nil, err
	}

	return []string{formatted}, nil
}

func (r intValueRenderer) Parse(ctx context.Context, s []string) (protoreflect.Value, error) {
	panic("implement me")
}

// formatInteger formats an integer into a value-rendered string. This function
// operates with string manipulation (instead of manipulating the int or sdk.Int
// object).
func formatInteger(v string) (string, error) {
	if v[0] == '-' {
		v = v[1:]
	}
	if len(v) > 1 {
		v = strings.TrimLeft(v, "0")

	}

	startOffset := 3
	for outputIndex := len(v); outputIndex > startOffset; {
		outputIndex -= 3
		v = v[:outputIndex] + thousandSeparator + v[outputIndex:]
	}
	return v, nil
}
