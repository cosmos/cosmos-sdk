package valuerenderer

import (
	"context"
	"fmt"
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

	_, err = io.WriteString(w, formatted)
	return err
}

func (vr intValueRenderer) Parse(_ context.Context, r io.Reader) (protoreflect.Value, error) {
	panic("implement me")
}

func hasOnlyDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// formatInteger formats an integer into a value-rendered string. This function
// operates with string manipulation (instead of manipulating the int or sdk.Int
// object).
func formatInteger(v string) (string, error) {
	sign := ""
	if v[0] == '-' {
		sign = "-"
		v = v[1:]
	}
	if len(v) > 1 {
		v = strings.TrimLeft(v, "0")
	}

	// Ensure that the string contains only digits at this point.
	if !hasOnlyDigits(v) {
		return "", fmt.Errorf("expecting only digits 0-9, but got non-digits in %q", v)
	}

	startOffset := 3
	for outputIndex := len(v); outputIndex > startOffset; {
		outputIndex -= 3
		v = v[:outputIndex] + thousandSeparator + v[outputIndex:]
	}

	return sign + v, nil
}
