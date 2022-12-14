package valuerenderer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

var hashPrefix = "SHA-256="

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

	// For long bytes, we display the hash.
	hasher := sha256.New()
	_, err := hasher.Write(bz)
	if err != nil {
		return nil, err
	}
	h := hasher.Sum(nil)

	text := fmt.Sprintf("%s%s", hashPrefix, strings.ToUpper(hex.EncodeToString(h)))
	return []Screen{{Text: text}}, nil
}

func (vr bytesValueRenderer) Parse(_ context.Context, screens []Screen) (protoreflect.Value, error) {
	if len(screens) != 1 {
		return nilValue, fmt.Errorf("expected single screen: %v", screens)
	}
	formatted := screens[0].Text

	// If the formatted string starts with `SHA-256=`, then we can't actually
	// invert to get the original bytes. In this case, we simply return the
	// hashed bytes.
	// Which means that in the general case, we trim the (optional) prefix, and
	// return whatever is left.
	hexStr := strings.TrimPrefix(formatted, hashPrefix)
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return nilValue, err
	}

	return protoreflect.ValueOfBytes(data), nil
}
