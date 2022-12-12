package valuerenderer

import (
	"context"
	"crypto/sha256"
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
	bz := v.Bytes()

	if len(bz) <= 32 {
		text := strings.ToUpper(hex.EncodeToString(v.Bytes()))
		return []Screen{{Text: text}}, nil
	}

	// For long bytes, we show the hash.
	hasher := sha256.New()
	_, err := hasher.Write(bz)
	if err != nil {
		return nil, err
	}
	h := hasher.Sum(nil)

	text := strings.ToUpper(hex.EncodeToString(h))
	return []Screen{{Text: text}}, nil
}

func (vr bytesValueRenderer) Parse(_ context.Context, screens []Screen) (protoreflect.Value, error) {
	if len(screens) != 1 {
		return protoreflect.ValueOfBytes([]byte{}), fmt.Errorf("expected single screen: %v", screens)
	}
	formatted := screens[0].Text

	// If the formatted string's length is 32 bytes, there's actually no way to
	// know if it was a hash of even longer bytes, or those bytes as-is. We
	// simply return the bytes as-is.
	data, err := hex.DecodeString(string(formatted))
	if err != nil {
		return protoreflect.ValueOfBytes([]byte{}), err
	}

	return protoreflect.ValueOfBytes(data), nil
}
