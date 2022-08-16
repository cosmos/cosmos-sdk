package valuerenderer

import (
	"context"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type enumRenderer struct{}

var _ ValueRenderer = (*enumRenderer)(nil)

func (er enumRenderer) Format(_ context.Context, v protoreflect.Value, w io.Writer) error {
	// TODO: Decide if we should be serializing the type and full package
	// path of the enum value so as to unmarshal it its .Parse method.
	formatted, err := formatInteger(v.String())
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, formatted)
	return err
}

func (er enumRenderer) Parse(_ context.Context, r io.Reader) (protoreflect.Value, error) {
	panic("unimplemented")
}
