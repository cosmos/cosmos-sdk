package textual

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
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
func (vr timestampValueRenderer) Format(_ context.Context, v protoreflect.Value) ([]Screen, error) {
	ts, err := toTimestamp(v.Message().Interface())
	if err != nil {
		return nil, err
	}
	t := ts.AsTime()

	// Format the Go Time as RFC 3339.
	s := t.Format(time.RFC3339Nano)
	return []Screen{{Content: s}}, nil
}

// Parse implements the ValueRenderer interface.
func (vr timestampValueRenderer) Parse(_ context.Context, screens []Screen) (protoreflect.Value, error) {
	// Parse the RFC 3339 input as a Go Time.
	if len(screens) != 1 {
		return protoreflect.Value{}, fmt.Errorf("expected single screen: %v", screens)
	}
	t, err := time.Parse(time.RFC3339Nano, screens[0].Content)
	if err != nil {
		return protoreflect.Value{}, err
	}

	// Convert Go Time to a proto Timestamp.
	timestamp := tspb.New(t)

	// Reflect the proto Timestamp.
	msg := timestamp.ProtoReflect()
	return protoreflect.ValueOfMessage(msg), nil
}

// convertToGoTime converts the proto Message to a timestamppb.Timestamp.
// The input msg can be:
// - either a timestamppb.Timestamp (in which case there's nothing to do),
// - or a dynamicpb.Message.
func toTimestamp(msg protoreflect.ProtoMessage) (*tspb.Timestamp, error) {
	switch msg := msg.(type) {
	case *tspb.Timestamp:
		return msg, nil
	case *dynamicpb.Message:
		s, n := getFieldValue(msg, "seconds").Int(), getFieldValue(msg, "nanos").Int()
		return &tspb.Timestamp{Seconds: s, Nanos: int32(n)}, nil
	default:
		return nil, fmt.Errorf("expected timestamppb.Timestamp or dynamicpb.Message, got %T", msg)
	}
}
