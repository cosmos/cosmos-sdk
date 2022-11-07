package valuerenderer

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/math"
)

// NewDecValueRenderer returns a ValueRenderer for encoding sdk.Dec cosmos
// scalars.
func NewDecValueRenderer() ValueRenderer {
	return decValueRenderer{}
}

type decValueRenderer struct{}

var _ ValueRenderer = decValueRenderer{}

func (vr decValueRenderer) Format(_ context.Context, v protoreflect.Value) ([]Screen, error) {
	formatted, err := math.FormatDec(v.String())
	if err != nil {
		return nil, err
	}
	return []Screen{{Text: formatted}}, nil
}

func (vr decValueRenderer) Parse(_ context.Context, screens []Screen) (protoreflect.Value, error) {
	if len(screens) != 1 {
		return protoreflect.Value{}, fmt.Errorf("expected single screen: %v", screens)
	}

	parsed, err := parseDec(screens[0].Text)
	if err != nil {
		return protoreflect.Value{}, err
	}

	return protoreflect.ValueOfString(parsed), nil
}

func parseDec(v string) (string, error) {
	parts := strings.Split(v, ".")
	if len(parts) > 2 {
		return "", fmt.Errorf("invalid decimal: too many points in %s", v)
	}

	intPart, err := parseInt(parts[0])
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

	if !math.HasOnlyDigits(decPart) {
		return "", fmt.Errorf("non-digits detected after decimal point in: %q", parts[1])
	}

	return intPart + "." + decPart, nil
}
