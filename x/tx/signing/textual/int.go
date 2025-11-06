// Package textual provides value renderers for basic protobuf scalar types.
// This file implements the integer value renderer for SIGN_MODE_TEXTUAL, handling
// signed and unsigned integers of various sizes, including Cosmos SDK math.Int scalars.
package textual

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/math"
)

// NewIntValueRenderer creates a ValueRenderer for protobuf integer types (uint32, uint64,
// int32, int64) and Cosmos SDK math.Int scalars. The renderer formats integers with
// thousand separators and parses them back to the appropriate integer type.
func NewIntValueRenderer(fd protoreflect.FieldDescriptor) ValueRenderer {
	return intValueRenderer{fd}
}

type intValueRenderer struct {
	fd protoreflect.FieldDescriptor
}

// Format converts a protobuf integer value to a single screen for display.
// The integer is formatted with thousand separators using math.FormatInt.
func (vr intValueRenderer) Format(_ context.Context, v protoreflect.Value) ([]Screen, error) {
	formatted, err := math.FormatInt(v.String())
	if err != nil {
		return nil, err
	}
	return []Screen{{Content: formatted}}, nil
}

// Parse converts a single screen back to a protobuf integer value.
// It expects exactly one screen, removes thousand separators, and parses the value
// according to the field descriptor's integer type (uint32, uint64, int32, int64, or string for math.Int).
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
		return protoreflect.ValueOfUint32(uint32(value)), nil

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
		return protoreflect.ValueOfInt32(int32(value)), nil

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

// parseInt parses a value-rendered string into an integer by removing thousand separators
// and preserving the sign. It handles both positive and negative integers formatted with
// apostrophe separators (e.g., "1'000'000" or "-1'000'000").
func parseInt(v string) (string, error) {
	if len(v) == 0 {
		return "", errors.New("expecting a non-empty string")
	}

	sign := ""
	if v[0] == '-' {
		sign = "-"
		v = v[1:]
	}

	// remove the 1000 separators (ex: 1'000'000 -> 1000000)
	v = strings.ReplaceAll(v, "'", "")

	return sign + v, nil
}
