package ormfield_test

import (
	"bytes"
	"testing"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/v3/assert"

	"cosmossdk.io/orm/encoding/ormfield"
)

func TestTimestamp(t *testing.T) {
	t.Parallel()
	cdc := ormfield.TimestampCodec{}

	t.Run("nil value", func(t *testing.T) {
		t.Parallel()
		buf := &bytes.Buffer{}
		assert.NilError(t, cdc.Encode(protoreflect.Value{}, buf))
		assert.Equal(t, 1, len(buf.Bytes()))
		val, err := cdc.Decode(buf)
		assert.NilError(t, err)
		assert.Assert(t, !val.IsValid())
	})

	t.Run("no nanos", func(t *testing.T) {
		t.Parallel()
		ts := timestamppb.New(time.Date(2022, 1, 1, 12, 30, 15, 0, time.UTC))
		val := protoreflect.ValueOfMessage(ts.ProtoReflect())
		buf := &bytes.Buffer{}
		assert.NilError(t, cdc.Encode(val, buf))
		assert.Equal(t, 6, len(buf.Bytes()))
		val2, err := cdc.Decode(buf)
		assert.NilError(t, err)
		assert.Equal(t, 0, cdc.Compare(val, val2))
	})

	t.Run("nanos", func(t *testing.T) {
		t.Parallel()
		ts := timestamppb.New(time.Date(2022, 1, 1, 12, 30, 15, 235809753, time.UTC))
		val := protoreflect.ValueOfMessage(ts.ProtoReflect())
		buf := &bytes.Buffer{}
		assert.NilError(t, cdc.Encode(val, buf))
		assert.Equal(t, 9, len(buf.Bytes()))
		val2, err := cdc.Decode(buf)
		assert.NilError(t, err)
		assert.Equal(t, 0, cdc.Compare(val, val2))
	})

	t.Run("min value", func(t *testing.T) {
		t.Parallel()
		ts := timestamppb.New(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC))
		val := protoreflect.ValueOfMessage(ts.ProtoReflect())
		buf := &bytes.Buffer{}
		assert.NilError(t, cdc.Encode(val, buf))
		assert.Equal(t, 6, len(buf.Bytes()))
		assert.Assert(t, bytes.Equal(buf.Bytes(), []byte{0, 0, 0, 0, 0, 0})) // the minimum value should be all zeros
		val2, err := cdc.Decode(buf)
		assert.NilError(t, err)
		assert.Equal(t, 0, cdc.Compare(val, val2))
	})

	t.Run("max value", func(t *testing.T) {
		t.Parallel()
		ts := timestamppb.New(time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC))
		val := protoreflect.ValueOfMessage(ts.ProtoReflect())
		buf := &bytes.Buffer{}
		assert.NilError(t, cdc.Encode(val, buf))
		assert.Equal(t, 9, len(buf.Bytes()))
		val2, err := cdc.Decode(buf)
		assert.NilError(t, err)
		assert.Equal(t, 0, cdc.Compare(val, val2))
	})
}

func TestTimestampOutOfRange(t *testing.T) {
	t.Parallel()
	cdc := ormfield.TimestampCodec{}

	tt := []struct {
		name      string
		ts        *timestamppb.Timestamp
		expectErr string
	}{
		{
			name:      "before min",
			ts:        timestamppb.New(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)),
			expectErr: "timestamp seconds is out of range",
		},
		{
			name:      "after max",
			ts:        timestamppb.New(time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)),
			expectErr: "timestamp seconds is out of range",
		},
		{
			name: "nanos too small",
			ts: &timestamppb.Timestamp{
				Seconds: 0,
				Nanos:   -1,
			},
			expectErr: "timestamp nanos is out of range",
		},

		{
			name: "nanos too big",
			ts: &timestamppb.Timestamp{
				Seconds: 0,
				Nanos:   1000000000,
			},
			expectErr: "timestamp nanos is out of range",
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			val := protoreflect.ValueOfMessage(tc.ts.ProtoReflect())
			buf := &bytes.Buffer{}
			err := cdc.Encode(val, buf)
			assert.ErrorContains(t, err, tc.expectErr)
		})
	}
}
