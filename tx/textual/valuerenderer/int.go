package valuerenderer

import (
	"context"
	"fmt"
	"strings"

	"cosmossdk.io/math"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// NewIntValueRenderer returns a ValueRenderer for uint32, uint64, int32 and
// int64, and sdk.Int scalars.
func NewIntValueRenderer() ValueRenderer {
	return intValueRenderer{}
}

type intValueRenderer struct{}

var _ ValueRenderer = intValueRenderer{}

func (vr intValueRenderer) Format(_ context.Context, v protoreflect.Value) ([]Screen, error) {
	formatted, err := math.FormatInt(v.String())
	if err != nil {
		return nil, err
	}
	return []Screen{{Text: formatted}}, nil
}

func (vr intValueRenderer) Parse(_ context.Context, screens []Screen) (protoreflect.Value, error) {
	if len(screens) != 1 {
		return protoreflect.Value{}, fmt.Errorf("expected single screen: %v", screens)
	}

	parsedInt, err := parseInt(screens[0].Text)
	if err != nil {
		return protoreflect.Value{}, err
	}

	return protoreflect.ValueOfString(parsedInt), nil
}

// parseInt parses a value-rendered string into an integer
func parseInt(v string) (string, error) {
	sign := ""
	if v[0] == '-' {
		sign = "-"
		v = v[1:]
	}

	// remove the 1000 separators (ex: 1'000'000 -> 1000000)
	v = strings.Replace(v, "'", "", -1)

	return sign + v, nil
}
