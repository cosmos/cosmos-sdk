package textual

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
	decStr := v.String()

	// If the decimal doesn't contain a point, we assume it's a badly formatted
	// value. So we assume it's a decimal with 18 decimal places and add the decimal
	// in the right place.
	if !strings.Contains(decStr, ".") {
		isNeg := false
		if strings.HasPrefix(decStr, "-") {
			decStr = strings.TrimPrefix(decStr, "-")
			isNeg = true
		}

		if len(decStr) < 19 {
			decStr = "0." + strings.Repeat("0", 18-len(decStr)) + decStr
		} else {
			whole, decimal := decStr[:len(decStr)-18], decStr[18:]
			decStr = whole + "." + decimal
		}

		if isNeg {
			decStr = "-" + decStr
		}
	}

	formatted, err := math.FormatDec(decStr)
	if err != nil {
		return nil, err
	}
	return []Screen{{Content: formatted}}, nil
}

func (vr decValueRenderer) Parse(_ context.Context, screens []Screen) (protoreflect.Value, error) {
	if n := len(screens); n != 1 {
		return nilValue, fmt.Errorf("expected 1 screen, got: %d", n)
	}

	parsed, err := parseDec(screens[0].Content)
	if err != nil {
		return nilValue, err
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

	return intPart + "." + parts[1], nil
}
