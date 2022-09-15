package valuerenderer

import (
	"context"
	"io"

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

func (vr decValueRenderer) Format(_ context.Context, v protoreflect.Value, w io.Writer) error {
	formatted, err := math.FormatDecimal(v.String())
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, formatted)
	return err
}

func (vr decValueRenderer) Parse(_ context.Context, r io.Reader) (protoreflect.Value, error) {
	panic("implement me")
}
