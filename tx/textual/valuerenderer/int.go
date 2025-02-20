package valuerenderer

import (
	"context"

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
	panic("implement me")
}
