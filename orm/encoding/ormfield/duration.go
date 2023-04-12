package ormfield

import (
	io "io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	durationSecondsField = durationMsgType.Descriptor().Fields().ByName("seconds")
	durationNanosField   = durationMsgType.Descriptor().Fields().ByName("nanos")
)

func getDurationSecondsAndNanos(value protoreflect.Value) (protoreflect.Value, protoreflect.Value) {
	msg := value.Message()
	return msg.Get(durationSecondsField), msg.Get(durationNanosField)
}

// DurationCodec encodes a google.protobuf.Duration value as 12 bytes using
// Int64Codec for seconds followed by Int32Codec for nanos. This allows for
// sorted iteration.
type DurationCodec struct{}

func (d DurationCodec) Decode(r Reader) (protoreflect.Value, error) {
	seconds, err := int64Codec.Decode(r)
	if err != nil {
		return protoreflect.Value{}, err
	}
	nanos, err := int32Codec.Decode(r)
	if err != nil {
		return protoreflect.Value{}, err
	}
	msg := durationMsgType.New()
	msg.Set(durationSecondsField, seconds)
	msg.Set(durationNanosField, nanos)
	return protoreflect.ValueOfMessage(msg), nil
}

func (d DurationCodec) Encode(value protoreflect.Value, w io.Writer) error {
	seconds, nanos := getDurationSecondsAndNanos(value)
	err := int64Codec.Encode(seconds, w)
	if err != nil {
		return err
	}
	return int32Codec.Encode(nanos, w)
}

func (d DurationCodec) Compare(v1, v2 protoreflect.Value) int {
	s1, n1 := getDurationSecondsAndNanos(v1)
	s2, n2 := getDurationSecondsAndNanos(v2)
	c := compareInt(s1, s2)
	if c != 0 {
		return c
	}

	return compareInt(n1, n2)
}

func (d DurationCodec) IsOrdered() bool {
	return true
}

func (d DurationCodec) FixedBufferSize() int {
	return 12
}

func (d DurationCodec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return d.FixedBufferSize(), nil
}
