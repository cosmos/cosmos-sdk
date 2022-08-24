package valuerenderer

import (
	"context"
	"fmt"
	"io"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

type timestampValueRenderer struct{}

// NewTimestampValueRenderer returns a ValueRenderer for protocol buffer Timestamp messages.
// It renders timestamps using the RFC 3339 format, always using UTC as the timezone.
// Fractional seconds are only rendered if nonzero.
func NewTimestampValueRenderer() ValueRenderer {
	return timestampValueRenderer{}
}

// Format implements the ValueRenderer interface.
func (vr timestampValueRenderer) Format(_ context.Context, v protoreflect.Value, w io.Writer) error {
	// Reify the reflected message as a proto Timestamp
	msg := v.Message().Interface()
	timestamp, ok := msg.(*tspb.Timestamp)
	if !ok {
		return fmt.Errorf("expected Timestamp, got %T", msg)
	}

	// Convert proto timestamp to a Go Time.
	t := timestamp.AsTime()

	// Format the Go Time as RFC 3339.
	s := t.Format(time.RFC3339Nano)
	w.Write([]byte(s))
	return nil
}

// Parse implements the ValueRenderer interface.
func (vr timestampValueRenderer) Parse(_ context.Context, r io.Reader) (protoreflect.Value, error) {
	// Parse the RFC 3339 input as a Go Time.
	bz, err := io.ReadAll(r)
	if err != nil {
		return protoreflect.Value{}, err
	}
	t, err := time.Parse(time.RFC3339Nano, string(bz))
	if err != nil {
		return protoreflect.Value{}, err
	}

	// Convert Go Time to a proto Timestamp.
	timestamp := tspb.New(t)

	// Reflect the proto Timestamp.
	msg := timestamp.ProtoReflect()
	return protoreflect.ValueOfMessage(msg), nil
}
