package aminojson

import (
	"bytes"
	"testing"

	"google.golang.org/protobuf/types/known/durationpb"
)

func TestMarshalDuration_Valid(t *testing.T) {
	msg := durationpb.New(0)
	// 1 second and 500 nanos => 1_000_000_000 + 500
	msg.Seconds = 1
	msg.Nanos = 500

	var buf bytes.Buffer
	if err := marshalDuration(nil, msg.ProtoReflect(), &buf); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := buf.String()
	want := "\"1000000500\""
	if got != want {
		t.Fatalf("unexpected output: got %s want %s", got, want)
	}
}

func TestMarshalDuration_InvalidNanosRange(t *testing.T) {
	cases := []int32{1_000_000_000, -1_000_000_000}
	for _, nanos := range cases {
		msg := &durationpb.Duration{Seconds: 0, Nanos: nanos}
		var buf bytes.Buffer
		if err := marshalDuration(nil, msg.ProtoReflect(), &buf); err == nil {
			t.Fatalf("expected error for nanos=%d, got none", nanos)
		}
	}
}

func TestMarshalDuration_SignMismatch(t *testing.T) {
	cases := []struct {
		seconds int64
		nanos   int32
	}{
		{seconds: 1, nanos: -1},
		{seconds: -1, nanos: 1},
	}

	for _, tc := range cases {
		msg := &durationpb.Duration{Seconds: tc.seconds, Nanos: tc.nanos}
		var buf bytes.Buffer
		if err := marshalDuration(nil, msg.ProtoReflect(), &buf); err == nil {
			t.Fatalf("expected error for seconds=%d nanos=%d, got none", tc.seconds, tc.nanos)
		}
	}
}

func TestMarshalDuration_OverflowUnderflow(t *testing.T) {
	// overflow
	msgOverflow := &durationpb.Duration{Seconds: MaxDurationSeconds + 1, Nanos: 0}
	if err := marshalDuration(nil, msgOverflow.ProtoReflect(), &bytes.Buffer{}); err == nil {
		t.Fatalf("expected overflow error, got none")
	}

	// underflow
	msgUnderflow := &durationpb.Duration{Seconds: -(MaxDurationSeconds + 1), Nanos: 0}
	if err := marshalDuration(nil, msgUnderflow.ProtoReflect(), &bytes.Buffer{}); err == nil {
		t.Fatalf("expected underflow error, got none")
	}
}
