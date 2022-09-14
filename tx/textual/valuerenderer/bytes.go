package valuerenderer

import (
	"context"
	"encoding/hex"
	"io"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// NewBytesValueRenderer returns a ValueRenderer for Protobuf bytes, which are
// encoded as capital-letter hexadecimal, without the '0x' prefix.
func NewBytesValueRenderer() ValueRenderer {
	return bytesValueRenderer{}
}

type bytesValueRenderer struct{}

var _ ValueRenderer = bytesValueRenderer{}

func (vr bytesValueRenderer) Format(ctx context.Context, v protoreflect.Value, w io.Writer) error {
	_, err := io.WriteString(w, strings.ToUpper(hex.EncodeToString(v.Bytes())))
	return err
}

func (vr bytesValueRenderer) Parse(_ context.Context, r io.Reader) (protoreflect.Value, error) {
	formatted, err := io.ReadAll(r)
	if err != nil {
		return protoreflect.ValueOfBytes([]byte{}), err
	}

	data, err := hex.DecodeString(string(formatted))
	if err != nil {
		return protoreflect.ValueOfBytes([]byte{}), err
	}

	return protoreflect.ValueOfBytes(data), nil
}
