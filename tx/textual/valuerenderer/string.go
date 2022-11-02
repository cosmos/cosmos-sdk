package valuerenderer

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type stringValueRenderer struct{}

// NewStringValueRenderer returns a ValueRenderer for protocol buffer string values.
// It renders the string as-is without quotation.
func NewStringValueRenderer() ValueRenderer {
	return stringValueRenderer{}
}

func (sr stringValueRenderer) Format(_ context.Context, v protoreflect.Value) ([]Screen, error) {
	return []Screen{{Text: v.String()}}, nil
}

func (sr stringValueRenderer) Parse(_ context.Context, screens []Screen) (protoreflect.Value, error) {
	if len(screens) != 1 {
		return protoreflect.Value{}, fmt.Errorf("expected single screen: %v", screens)
	}
	return protoreflect.ValueOfString(screens[0].Text), nil
}
