package valuerenderer

import (
	"context"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type anyValueRenderer struct {
}

func NewAnyValueRenderer() ValueRenderer {
	return anyValueRenderer{}
}

func (ar anyValueRenderer) Format(ctx context.Context, v protoreflect.Value) ([]Screen, error) {
	return nil, nil
}

func (ar anyValueRenderer) Parse(ctx context.Context, screens []Screen) (protoreflect.Value, error) {
	return protoreflect.Value{}, nil
}
