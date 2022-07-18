package valuerenderer

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const thousandSeparator string = "'"

type decValueRenderer struct{}

var _ ValueRenderer = decValueRenderer{}

func (r decValueRenderer) Format(_ context.Context, v protoreflect.Value) ([]string, error) {
	formatted, err := formatDecimal(v.String())
	if err != nil {
		return nil, err
	}

	return []string{formatted}, nil
}

func (r decValueRenderer) Parse(_ context.Context, s []string) (protoreflect.Value, error) {
	panic("implement me")
}

// formatDecimal formats a decimal into a value-rendered string. This function
// operates with string manipulation (instead of manipulating the sdk.Dec
// object).
func formatDecimal(v string) (string, error) {
	parts := strings.Split(v, ".")
	intPart, err := formatInteger(parts[0])
	if err != nil {
		return "", err
	}

	if len(parts) > 2 {
		return "", fmt.Errorf("invalid decimal %s", v)
	}

	if len(parts) == 1 {
		return intPart, nil
	}

	decPart := strings.TrimRight(parts[1], "0")
	if len(decPart) == 0 {
		return intPart, nil
	}

	return intPart + "." + decPart, nil
}
