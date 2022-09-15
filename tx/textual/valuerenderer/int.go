package valuerenderer

import (
	"context"
	"io"

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

func (vr intValueRenderer) Format(_ context.Context, v protoreflect.Value, w io.Writer) error {
	formatted, err := math.FormatInteger(v.String())
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, formatted)
	return err
}

func (vr intValueRenderer) Parse(_ context.Context, r io.Reader) (protoreflect.Value, error) {
	panic("implement me")
}
