package ormfield_test

import (
	"bytes"
	"testing"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"gotest.tools/v3/assert"

	"cosmossdk.io/orm/encoding/ormfield"
)

func TestDuration(t *testing.T) {
	t.Parallel()
	cdc := ormfield.DurationCodec{}

	// nil value
	t.Run("nil value", func(t *testing.T) {
		t.Parallel()
		buf := &bytes.Buffer{}
		assert.NilError(t, cdc.Encode(protoreflect.Value{}, buf))
		assert.Equal(t, 1, len(buf.Bytes()))
		val, err := cdc.Decode(buf)
		assert.NilError(t, err)
		assert.Assert(t, !val.IsValid())
	})

	// no nanos
	t.Run("no nanos", func(t *testing.T) {
		t.Parallel()
		dur, err := time.ParseDuration("100s")
		assert.NilError(t, err)
		durPb := durationpb.New(dur)
		val := protoreflect.ValueOfMessage(durPb.ProtoReflect())
		buf := &bytes.Buffer{}
		assert.NilError(t, cdc.Encode(val, buf))
		assert.Equal(t, 6, len(buf.Bytes()))
		val2, err := cdc.Decode(buf)
		assert.NilError(t, err)
		assert.Equal(t, 0, cdc.Compare(val, val2))
	})

	t.Run("nanos", func(t *testing.T) {
		t.Parallel()
		dur, err := time.ParseDuration("3879468295ns")
		assert.NilError(t, err)
		durPb := durationpb.New(dur)
		val := protoreflect.ValueOfMessage(durPb.ProtoReflect())
		buf := &bytes.Buffer{}
		assert.NilError(t, cdc.Encode(val, buf))
		assert.Equal(t, 9, len(buf.Bytes()))
		val2, err := cdc.Decode(buf)
		assert.NilError(t, err)
		assert.Equal(t, 0, cdc.Compare(val, val2))
	})

	t.Run("min value", func(t *testing.T) {
		t.Parallel()
		durPb := &durationpb.Duration{
			Seconds: -315576000000,
			Nanos:   -999999999,
		}
		val := protoreflect.ValueOfMessage(durPb.ProtoReflect())
		buf := &bytes.Buffer{}
		assert.NilError(t, cdc.Encode(val, buf))
		assert.Equal(t, 9, len(buf.Bytes()))
		val2, err := cdc.Decode(buf)
		assert.NilError(t, err)
		assert.Equal(t, 0, cdc.Compare(val, val2))
	})

	t.Run("max value", func(t *testing.T) {
		t.Parallel()
		durPb := &durationpb.Duration{
			Seconds: 315576000000,
			Nanos:   999999999,
		}
		val := protoreflect.ValueOfMessage(durPb.ProtoReflect())
		buf := &bytes.Buffer{}
		assert.NilError(t, cdc.Encode(val, buf))
		assert.Equal(t, 9, len(buf.Bytes()))
		val2, err := cdc.Decode(buf)
		assert.NilError(t, err)
		assert.Equal(t, 0, cdc.Compare(val, val2))
	})
}

func TestDurationOutOfRange(t *testing.T) {
	t.Parallel()
	cdc := ormfield.DurationCodec{}

	tt := []struct {
		name      string
		dur       *durationpb.Duration
		expectErr string
	}{
		{
			name: "seconds too small",
			dur: &durationpb.Duration{
				Seconds: -315576000001,
				Nanos:   0,
			},
			expectErr: "seconds is out of range",
		},
		{
			name: "seconds too big",
			dur: &durationpb.Duration{
				Seconds: 315576000001,
				Nanos:   0,
			},
			expectErr: "seconds is out of range",
		},
		{
			name: "positive seconds negative nanos",
			dur: &durationpb.Duration{
				Seconds: 0,
				Nanos:   -1,
			},
			expectErr: "nanos is out of range",
		},
		{
			name: "positive seconds nanos too big",
			dur: &durationpb.Duration{
				Seconds: 0,
				Nanos:   1000000000,
			},
			expectErr: "nanos is out of range",
		},
		{
			name: "negative seconds positive nanos",
			dur: &durationpb.Duration{
				Seconds: -1,
				Nanos:   1,
			},
			expectErr: "negative duration nanos is out of range",
		},
		{
			name: "negative seconds nanos too small",
			dur: &durationpb.Duration{
				Seconds: -1,
				Nanos:   -1000000000,
			},
			expectErr: "negative duration nanos is out of range",
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			val := protoreflect.ValueOfMessage(tc.dur.ProtoReflect())
			buf := &bytes.Buffer{}
			err := cdc.Encode(val, buf)
			assert.ErrorContains(t, err, tc.expectErr)
		})
	}
}
