package valuerenderer

import (
	"context"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type stringValueRenderer struct {
}

var _ ValueRenderer = stringValueRenderer{}

func (sr stringValueRenderer) Format(_ context.Context, v protoreflect.Value, w io.Writer) error {
	_, err := io.WriteString(w, v.String())
	return err
}

func (sr stringValueRenderer) Parse(_ context.Context, r io.Reader) (protoreflect.Value, error) {
	panic("implement me")
}
