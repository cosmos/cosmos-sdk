package valuerenderer

import (
	"context"
	"fmt"
	"io"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const thousandSeparator string = "'"

type decValueRenderer struct{}

var _ ValueRenderer = decValueRenderer{}

func (vr decValueRenderer) Format(_ context.Context, v protoreflect.Value, w io.Writer) error {
	formatted, err := formatDecimal(v.String())
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, formatted)
	return err
}

func (vr decValueRenderer) Parse(_ context.Context, r io.Reader) (protoreflect.Value, error) {
	panic("implement me")
}

// formatDecimal formats a decimal into a value-rendered string. This function
// operates with string manipulation (instead of manipulating the sdk.Dec
// object).
func formatDecimal(v string) (string, error) {
	parts := strings.Split(v, ".")
	if len(parts) > 2 {
		return "", fmt.Errorf("invalid decimal: too many points in %s", v)
	}

	intPart, err := formatInteger(parts[0])
	if err != nil {
		return "", err
	}

	if len(parts) == 1 {
		return intPart, nil
	}

	decPart := strings.TrimRight(parts[1], "0")
	if len(decPart) == 0 {
		return intPart, nil
	}

	// Ensure that the decimal part has only digits.
	// https://github.com/cosmos/cosmos-sdk/issues/12811
	if !hasOnlyDigits(decPart) {
		return "", fmt.Errorf("non-digits detected after decimal point in: %q", decPart)
	}

	return intPart + "." + decPart, nil
}
