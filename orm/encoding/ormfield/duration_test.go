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

func TestDurationNil(t *testing.T) {
	t.Parallel()

	cdc := ormfield.DurationCodec{}
	buf := &bytes.Buffer{}
	assert.NilError(t, cdc.Encode(protoreflect.Value{}, buf))
	assert.Equal(t, 1, len(buf.Bytes()))
	val, err := cdc.Decode(buf)
	assert.NilError(t, err)
	assert.Assert(t, !val.IsValid())
}

func TestDuration(t *testing.T) {
	t.Parallel()
	cdc := ormfield.DurationCodec{}

	tt := []struct {
		name    string
		seconds int64
		nanos   int32
		wantLen int
	}{
		{
			"no nanos",
			100,
			0,
			9,
		},
		{
			"with nanos",
			3,
			879468295,
			9,
		},
		{
			"min seconds, -1 nanos",
			-315576000000,
			-1,
			9,
		},
		{
			"min value",
			-315576000000,
			-999999999,
			9,
		},
		{
			"max value",
			315576000000,
			999999999,
			9,
		},
		{
			"max seconds, 1 nanos",
			315576000000,
			1,
			9,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			durPb := &durationpb.Duration{
				Seconds: tc.seconds,
				Nanos:   tc.nanos,
			}
			val := protoreflect.ValueOfMessage(durPb.ProtoReflect())
			buf := &bytes.Buffer{}
			assert.NilError(t, cdc.Encode(val, buf))
			assert.Equal(t, tc.wantLen, len(buf.Bytes()))
			val2, err := cdc.Decode(buf)
			assert.NilError(t, err)
			assert.Equal(t, 0, cdc.Compare(val, val2))
		})
	}
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
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			val := protoreflect.ValueOfMessage(tc.dur.ProtoReflect())
			buf := &bytes.Buffer{}
			err := cdc.Encode(val, buf)
			assert.ErrorContains(t, err, tc.expectErr)
		})
	}
}

func TestDurationCompare(t *testing.T) {
	t.Parallel()
	cdc := ormfield.DurationCodec{}

	tt := []struct {
		name string
		dur1 *durationpb.Duration
		dur2 *durationpb.Duration
		want int
	}{
		{
			name: "equal",
			dur1: &durationpb.Duration{
				Seconds: 1,
				Nanos:   1,
			},
			dur2: &durationpb.Duration{
				Seconds: 1,
				Nanos:   1,
			},
			want: 0,
		},
		{
			name: "seconds equal, dur1 nanos less than dur2 nanos",
			dur1: &durationpb.Duration{
				Seconds: 1,
				Nanos:   1,
			},
			dur2: &durationpb.Duration{
				Seconds: 1,
				Nanos:   2,
			},
			want: -1,
		},
		{
			name: "seconds equal, dur1 nanos greater than dur2 nanos",
			dur1: &durationpb.Duration{
				Seconds: 1,
				Nanos:   2,
			},
			dur2: &durationpb.Duration{
				Seconds: 1,
				Nanos:   1,
			},
			want: 1,
		},
		{
			name: "seconds less than",
			dur1: &durationpb.Duration{
				Seconds: 1,
				Nanos:   1,
			},
			dur2: &durationpb.Duration{
				Seconds: 2,
				Nanos:   1,
			},
			want: -1,
		},
		{
			name: "seconds greater than",
			dur1: &durationpb.Duration{
				Seconds: 2,
				Nanos:   1,
			},
			dur2: &durationpb.Duration{
				Seconds: 1,
				Nanos:   1,
			},
			want: 1,
		},
		{
			name: "negative seconds equal, dur1 nanos less than dur2 nanos",
			dur1: &durationpb.Duration{
				Seconds: -1,
				Nanos:   -2,
			},
			dur2: &durationpb.Duration{
				Seconds: -1,
				Nanos:   -1,
			},
			want: -1,
		},
		{
			name: "negative seconds equal, dur1 nanos zero",
			dur1: &durationpb.Duration{
				Seconds: -1,
				Nanos:   0,
			},
			dur2: &durationpb.Duration{
				Seconds: -1,
				Nanos:   -1,
			},
			want: 1,
		},
		{
			name: "negative seconds equal, dur2 nanos zero",
			dur1: &durationpb.Duration{
				Seconds: -1,
				Nanos:   -1,
			},
			dur2: &durationpb.Duration{
				Seconds: -1,
				Nanos:   0,
			},
			want: -1,
		},
		{
			name: "seconds equal and dur1 nanos min values",
			dur1: &durationpb.Duration{
				Seconds: ormfield.DurationSecondsMin,
				Nanos:   ormfield.DurationNanosMin,
			},
			dur2: &durationpb.Duration{
				Seconds: ormfield.DurationSecondsMin,
				Nanos:   -1,
			},
			want: -1,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			val1 := protoreflect.ValueOfMessage(tc.dur1.ProtoReflect())
			val2 := protoreflect.ValueOfMessage(tc.dur2.ProtoReflect())
			got := cdc.Compare(val1, val2)
			assert.Equal(t, tc.want, got, "Compare(%v, %v)", tc.dur1, tc.dur2)

			bz1 := encodeValue(t, cdc, val1)
			bz2 := encodeValue(t, cdc, val2)
			assert.Equal(t, tc.want, bytes.Compare(bz1, bz2), "bytes.Compare(%v, %v)", bz1, bz2)
		})
	}

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
}

func encodeValue(t *testing.T, cdc ormfield.Codec, val protoreflect.Value) []byte {
	t.Helper()
	buf := &bytes.Buffer{}
	assert.NilError(t, cdc.Encode(val, buf))
	return buf.Bytes()
}
