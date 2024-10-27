package valuerenderer

import (
	"context"
	"encoding/hex"
	"fmt"
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

func (vr bytesValueRenderer) Format(ctx context.Context, v protoreflect.Value) ([]Screen, error) {
	text := strings.ToUpper(hex.EncodeToString(v.Bytes()))
	return []Screen{{Text: text}}, nil
}

func (vr bytesValueRenderer) Parse(_ context.Context, screens []Screen) (protoreflect.Value, error) {
	if len(screens) != 1 {
		return protoreflect.ValueOfBytes([]byte{}), fmt.Errorf("expected single screen: %v", screens)
	}
	formatted := screens[0].Text

	data, err := hex.DecodeString(string(formatted))
	if err != nil {
		return protoreflect.ValueOfBytes([]byte{}), err
	}

	return protoreflect.ValueOfBytes(data), nil
}
