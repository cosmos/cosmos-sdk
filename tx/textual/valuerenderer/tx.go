package valuerenderer

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	tspb "google.golang.org/protobuf/types/known/timestamppb"

	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
)

type txValueRenderer struct{}

// NewTimestampValueRenderer returns a ValueRenderer for the protobuf Tx type,
// as called the transaction envelope. It follows the specification defined
// in ADR-050.
func NewTxValueRenderer() ValueRenderer {
	return txValueRenderer{}
}

// Format implements the ValueRenderer interface.
func (vr txValueRenderer) Format(_ context.Context, v protoreflect.Value) ([]Screen, error) {
	// Reify the reflected message as a proto Timestamp
	msg := v.Message().Interface()
	timestamp, ok := msg.(*txv1beta1.Tx)
	if !ok {
		return nil, fmt.Errorf("expected Tx, got %T", msg)
	}

	// Convert proto timestamp to a Go Time.
	t := timestamp.AsTime()

	// Format the Go Time as RFC 3339.
	s := t.Format(time.RFC3339Nano)
	return []Screen{{Text: s}}, nil
}

// Parse implements the ValueRenderer interface.
func (vr txValueRenderer) Parse(_ context.Context, screens []Screen) (protoreflect.Value, error) {
	// Parse the RFC 3339 input as a Go Time.
	if len(screens) != 1 {
		return protoreflect.Value{}, fmt.Errorf("expected single screen: %v", screens)
	}
	t, err := time.Parse(time.RFC3339Nano, screens[0].Text)
	if err != nil {
		return protoreflect.Value{}, err
	}

	// Convert Go Time to a proto Timestamp.
	timestamp := tspb.New(t)

	// Reflect the proto Timestamp.
	msg := timestamp.ProtoReflect()
	return protoreflect.ValueOfMessage(msg), nil
}
