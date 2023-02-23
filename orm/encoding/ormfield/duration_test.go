package ormfield_test

import (
	"bytes"
	"testing"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormfield"
)

func TestDuration(t *testing.T) {
	cdc := ormfield.DurationCodec{}

	// nil value
	buf := &bytes.Buffer{}
	assert.NilError(t, cdc.Encode(protoreflect.Value{}, buf))
	assert.Equal(t, 1, len(buf.Bytes()))
	val, err := cdc.Decode(buf)
	assert.NilError(t, err)
	assert.Assert(t, !val.IsValid())

	// no nanos
	dur, err := time.ParseDuration("100s")
	assert.NilError(t, err)
	durPb := durationpb.New(dur)
	val = protoreflect.ValueOfMessage(durPb.ProtoReflect())
	buf = &bytes.Buffer{}
	assert.NilError(t, cdc.Encode(val, buf))
	assert.Equal(t, 6, len(buf.Bytes()))
	val2, err := cdc.Decode(buf)
	assert.NilError(t, err)
	assert.Equal(t, 0, cdc.Compare(val, val2))

	// nanos
	dur, err = time.ParseDuration("3879468295ns")
	assert.NilError(t, err)
	durPb = durationpb.New(dur)
	val = protoreflect.ValueOfMessage(durPb.ProtoReflect())
	buf = &bytes.Buffer{}
	assert.NilError(t, cdc.Encode(val, buf))
	assert.Equal(t, 9, len(buf.Bytes()))
	val2, err = cdc.Decode(buf)
	assert.NilError(t, err)
	assert.Equal(t, 0, cdc.Compare(val, val2))

	// min value
	assert.NilError(t, err)
	durPb = &durationpb.Duration{
		Seconds: -315576000000,
		Nanos:   -999999999,
	}
	val = protoreflect.ValueOfMessage(durPb.ProtoReflect())
	buf = &bytes.Buffer{}
	assert.NilError(t, cdc.Encode(val, buf))
	assert.Equal(t, 9, len(buf.Bytes()))
	val2, err = cdc.Decode(buf)
	assert.NilError(t, err)
	assert.Equal(t, 0, cdc.Compare(val, val2))

	// max value
	assert.NilError(t, err)
	durPb = &durationpb.Duration{
		Seconds: 315576000000,
		Nanos:   999999999,
	}
	val = protoreflect.ValueOfMessage(durPb.ProtoReflect())
	buf = &bytes.Buffer{}
	assert.NilError(t, cdc.Encode(val, buf))
	assert.Equal(t, 9, len(buf.Bytes()))
	val2, err = cdc.Decode(buf)
	assert.NilError(t, err)
	assert.Equal(t, 0, cdc.Compare(val, val2))
}
