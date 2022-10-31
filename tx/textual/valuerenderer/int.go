package valuerenderer

import (
	"context"
	"fmt"

	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
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

	parsedInt, err := math.ParseInt(screens[0].Text)
	if err != nil {
		return protoreflect.Value{}, err
	}

	i := basev1beta1.IntProto{
		Int: parsedInt.String(),
	}

	msg := i.ProtoReflect()
	return protoreflect.ValueOfMessage(msg), nil
}
