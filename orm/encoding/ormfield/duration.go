package ormfield

import (
	"fmt"
	io "io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	durationNilValue   = 0x40
	durationSecondsMin = -315576000000
	durationSecondsMax = 315576000000
	durationNanosMin   = -999999999
	durationNanosMax   = 999999999
)

var (
	durationNilBz = []byte{durationNilValue}
)

type DurationCodec struct{}

func (d DurationCodec) Encode(value protoreflect.Value, w io.Writer) error {
	// nil case
	if !value.IsValid() {
		_, err := w.Write(durationNilBz)
		return err
	}

	seconds, nanos := getDurationSecondsAndNanos(value)
	secondsInt := seconds.Int()
	if secondsInt < durationSecondsMin || secondsInt > durationSecondsMax {
		return fmt.Errorf("duration seconds is out of range %d, must be between %d and %d", secondsInt, durationSecondsMin, durationSecondsMax)
	}

	// TODO: implement me

	nanosInt := nanos.Int()
	if nanosInt < durationNanosMin || nanosInt > durationNanosMax {
		return fmt.Errorf("duration nanos is out of range %d, must be between %d and %d", secondsInt, durationNanosMin, durationNanosMax)
	}

	panic("implement me")
}

func (d DurationCodec) Decode(r Reader) (protoreflect.Value, error) {
	//TODO implement me
	panic("implement me")
}

func (d DurationCodec) Compare(v1, v2 protoreflect.Value) int {
	//TODO implement me
	panic("implement me")
}

func (d DurationCodec) IsOrdered() bool {
	//TODO implement me
	panic("implement me")
}

func (d DurationCodec) FixedBufferSize() int {
	//TODO implement me
	panic("implement me")
}

func (d DurationCodec) ComputeBufferSize(value protoreflect.Value) (int, error) {
	//TODO implement me
	panic("implement me")
}

var (
	durationSecondsField = durationMsgType.Descriptor().Fields().ByName("seconds")
	durationNanosField   = durationMsgType.Descriptor().Fields().ByName("nanos")
)

func getDurationSecondsAndNanos(value protoreflect.Value) (protoreflect.Value, protoreflect.Value) {
	msg := value.Message()
	return msg.Get(durationSecondsField), msg.Get(durationNanosField)
}

// DurationV0Codec encodes a google.protobuf.Duration value as 12 bytes using
// Int64Codec for seconds followed by Int32Codec for nanos. This allows for
// sorted iteration.
type DurationV0Codec struct{}

func (d DurationV0Codec) Decode(r Reader) (protoreflect.Value, error) {
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

func (d DurationV0Codec) Encode(value protoreflect.Value, w io.Writer) error {
	seconds, nanos := getDurationSecondsAndNanos(value)
	err := int64Codec.Encode(seconds, w)
	if err != nil {
		return err
	}
	return int32Codec.Encode(nanos, w)
}

func (d DurationV0Codec) Compare(v1, v2 protoreflect.Value) int {
	s1, n1 := getDurationSecondsAndNanos(v1)
	s2, n2 := getDurationSecondsAndNanos(v2)
	c := compareInt(s1, s2)
	if c != 0 {
		return c
	} else {
		return compareInt(n1, n2)
	}
}

func (d DurationV0Codec) IsOrdered() bool {
	return true
}

func (d DurationV0Codec) FixedBufferSize() int {
	return 12
}

func (d DurationV0Codec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return d.FixedBufferSize(), nil
}
