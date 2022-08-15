package valuerenderer

import (
	"context"
	"encoding/base64"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// bytesValueRenderer implements ValueRenderer for bytes
type bytesValueRenderer struct{}

var _ ValueRenderer = bytesValueRenderer{}

func (vr bytesValueRenderer) Format(ctx context.Context, v protoreflect.Value, w io.Writer) error {
	_, err := io.WriteString(w, base64.StdEncoding.EncodeToString(v.Bytes()))
	return err
}

func (vr bytesValueRenderer) Parse(_ context.Context, r io.Reader) (protoreflect.Value, error) {
	formatted, err := io.ReadAll(r)
	if err != nil {
		return protoreflect.ValueOfBytes([]byte{}), err
	}

	data, err := base64.StdEncoding.DecodeString(string(formatted))
	if err != nil {
		return protoreflect.ValueOfBytes([]byte{}), err
	}

	return protoreflect.ValueOfBytes(data), nil
}
