package valuerenderer

import (
	"context"

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
	panic("implement me")
}
