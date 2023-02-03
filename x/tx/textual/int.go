package textual

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"cosmossdk.io/math"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// NewIntValueRenderer returns a ValueRenderer for uint32, uint64, int32 and
// int64, and sdk.Int scalars.
func NewIntValueRenderer(fd protoreflect.FieldDescriptor) ValueRenderer {
	return intValueRenderer{fd}
}

type intValueRenderer struct {
	fd protoreflect.FieldDescriptor
}

var _ ValueRenderer = intValueRenderer{}

func (vr intValueRenderer) Format(_ context.Context, v protoreflect.Value) ([]Screen, error) {
	formatted, err := math.FormatInt(v.String())
	if err != nil {
		return nil, err
	}
	return []Screen{{Content: formatted}}, nil
}

func (vr intValueRenderer) Parse(_ context.Context, screens []Screen) (protoreflect.Value, error) {
	if n := len(screens); n != 1 {
		return nilValue, fmt.Errorf("expected 1 screen, got: %d", n)
	}

	parsedInt, err := parseInt(screens[0].Content)
	if err != nil {
		return nilValue, err
	}

	switch vr.fd.Kind() {
	case protoreflect.Uint32Kind:
		value, err := strconv.ParseUint(parsedInt, 10, 32)
		if err != nil {
			return nilValue, err
		}
		return protoreflect.ValueOfUint32(uint32(value)), nil //nolint:gosec

	case protoreflect.Uint64Kind:
		value, err := strconv.ParseUint(parsedInt, 10, 64)
		if err != nil {
			return nilValue, err
		}
		return protoreflect.ValueOfUint64(value), nil

	case protoreflect.Int32Kind:
		value, err := strconv.ParseInt(parsedInt, 10, 32)
		if err != nil {
			return nilValue, err
		}
		return protoreflect.ValueOfInt32(int32(value)), nil //nolint:gosec

	case protoreflect.Int64Kind:
		value, err := strconv.ParseInt(parsedInt, 10, 64)
		if err != nil {
			return nilValue, err
		}
		return protoreflect.ValueOfInt64(value), nil

	case protoreflect.StringKind:
		return protoreflect.ValueOfString(parsedInt), nil

	default:
		return nilValue, fmt.Errorf("parsing integers into a %s field is not supported", vr.fd.Kind())
	}
}

// parseInt parses a value-rendered string into an integer
func parseInt(v string) (string, error) {
	sign := ""
	if v[0] == '-' {
		sign = "-"
		v = v[1:]
	}

	// remove the 1000 separators (ex: 1'000'000 -> 1000000)
	v = strings.ReplaceAll(v, "'", "")

	return sign + v, nil
}
