// Package textual provides value renderers for basic protobuf scalar types.
// This file implements the string value renderer for SIGN_MODE_TEXTUAL.
package textual

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type stringValueRenderer struct{}

// NewStringValueRenderer creates a ValueRenderer for protobuf string values.
// The renderer displays strings without quotation marks and expects single-screen
// input when parsing, following the ADR-050 specification for string formatting.
func NewStringValueRenderer() ValueRenderer {
	return stringValueRenderer{}
}

// Format converts a protobuf string value to a single screen for display.
// The string content is displayed directly without any additional formatting.
func (sr stringValueRenderer) Format(_ context.Context, v protoreflect.Value) ([]Screen, error) {
	return []Screen{{Content: v.String()}}, nil
}

// Parse converts a single screen back to a protobuf string value.
// It expects exactly one screen and returns an error if multiple screens are provided.
func (sr stringValueRenderer) Parse(_ context.Context, screens []Screen) (protoreflect.Value, error) {
	if len(screens) != 1 {
		return nilValue, fmt.Errorf("expected single screen: %v", screens)
	}
	return protoreflect.ValueOfString(screens[0].Content), nil
}
