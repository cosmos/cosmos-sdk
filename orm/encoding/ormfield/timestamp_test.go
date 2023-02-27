package ormfield_test

import (
	"bytes"
	"testing"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormfield"
)

func TestTimestamp(t *testing.T) {
	cdc := ormfield.TimestampCodec{}

	// nil value
	buf := &bytes.Buffer{}
	assert.NilError(t, cdc.Encode(protoreflect.Value{}, buf))
	assert.Equal(t, 1, len(buf.Bytes()))
	val, err := cdc.Decode(buf)
	assert.NilError(t, err)
	assert.Assert(t, !val.IsValid())

	// no nanos
	ts := timestamppb.New(time.Date(2022, 1, 1, 12, 30, 15, 0, time.UTC))
	val = protoreflect.ValueOfMessage(ts.ProtoReflect())
	buf = &bytes.Buffer{}
	assert.NilError(t, cdc.Encode(val, buf))
	assert.Equal(t, 6, len(buf.Bytes()))
	val2, err := cdc.Decode(buf)
	assert.NilError(t, err)
	assert.Equal(t, 0, cdc.Compare(val, val2))

	// nanos
	ts = timestamppb.New(time.Date(2022, 1, 1, 12, 30, 15, 235809753, time.UTC))
	val = protoreflect.ValueOfMessage(ts.ProtoReflect())
	buf = &bytes.Buffer{}
	assert.NilError(t, cdc.Encode(val, buf))
	assert.Equal(t, 9, len(buf.Bytes()))
	val2, err = cdc.Decode(buf)
	assert.NilError(t, err)
	assert.Equal(t, 0, cdc.Compare(val, val2))

	// min value
	ts = timestamppb.New(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC))
	val = protoreflect.ValueOfMessage(ts.ProtoReflect())
	buf = &bytes.Buffer{}
	assert.NilError(t, cdc.Encode(val, buf))
	assert.Equal(t, 6, len(buf.Bytes()))
	assert.Assert(t, bytes.Equal(buf.Bytes(), []byte{0, 0, 0, 0, 0, 0})) // the minimum value should be all zeros
	val2, err = cdc.Decode(buf)
	assert.NilError(t, err)
	assert.Equal(t, 0, cdc.Compare(val, val2))

	// max value
	ts = timestamppb.New(time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC))
	val = protoreflect.ValueOfMessage(ts.ProtoReflect())
	buf = &bytes.Buffer{}
	assert.NilError(t, cdc.Encode(val, buf))
	assert.Equal(t, 9, len(buf.Bytes()))
	val2, err = cdc.Decode(buf)
	assert.NilError(t, err)
	assert.Equal(t, 0, cdc.Compare(val, val2))
}
