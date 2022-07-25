package valuerenderer

import (
	"context"
	"io"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type intValueRenderer struct{}

var _ ValueRenderer = intValueRenderer{}

func (vr intValueRenderer) Format(_ context.Context, v protoreflect.Value, w io.Writer) error {
	formatted, err := formatInteger(v.String())
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(formatted))
	return err
}

func (vr intValueRenderer) Parse(_ context.Context, r io.Reader) (protoreflect.Value, error) {
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
