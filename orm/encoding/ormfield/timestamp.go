package ormfield

import (
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// TimestampCodec DurationCodec encodes a google.protobuf.Timestamp value as 12 bytes using
// Int64Codec for seconds followed by Int32Codec for nanos. This allows for
// sorted iteration.
type TimestampCodec struct{}

var (
	timestampSecondsField = timestampMsgType.Descriptor().Fields().ByName("seconds")
	timestampNanosField   = timestampMsgType.Descriptor().Fields().ByName("nanos")
)

func getTimestampSecondsAndNanos(value protoreflect.Value) (protoreflect.Value, protoreflect.Value) {
	msg := value.Message()
	return msg.Get(timestampSecondsField), msg.Get(timestampNanosField)
}

func (t TimestampCodec) Decode(r Reader) (protoreflect.Value, error) {
	seconds, err := int64Codec.Decode(r)
	if err != nil {
		return protoreflect.Value{}, err
	}
	nanos, err := int32Codec.Decode(r)
	if err != nil {
		return protoreflect.Value{}, err
	}
	msg := timestampMsgType.New()
	msg.Set(timestampSecondsField, seconds)
	msg.Set(timestampNanosField, nanos)
	return protoreflect.ValueOfMessage(msg), nil
}

func (t TimestampCodec) Encode(value protoreflect.Value, w io.Writer) error {
	seconds, nanos := getTimestampSecondsAndNanos(value)
	err := int64Codec.Encode(seconds, w)
	if err != nil {
		return err
	}
	return int32Codec.Encode(nanos, w)
}

func (t TimestampCodec) Compare(v1, v2 protoreflect.Value) int {
	s1, n1 := getTimestampSecondsAndNanos(v1)
	s2, n2 := getTimestampSecondsAndNanos(v2)
	c := compareInt(s1, s2)
	if c != 0 {
		return c
	} else {
		return compareInt(n1, n2)
	}
}

func (t TimestampCodec) IsOrdered() bool {
	return true
}

func (t TimestampCodec) FixedBufferSize() int {
	return 12
}

func (t TimestampCodec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return t.FixedBufferSize(), nil
}
