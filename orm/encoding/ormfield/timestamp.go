package ormfield

import (
	"fmt"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// TimestampCodec encodes google.protobuf.Timestamp values with the following
// encoding:
// - nil is encoded as []byte{0xFF}
// - seconds (which can range from 0001-01-01T00:00:00Z to 9999-12-31T23:59:59Z) is encoded as 5 fixed bytes
// - nanos (which can range from 0 to 999,999,999) is encoded as:
//   - []byte{0x0} for zero nanos
//   - 4 fixed bytes with the bit mask 0xC0 applied to the first byte
//
// When iterating over timestamp indexes, nil values will always be ordered last.
//
// Values for seconds and nanos outside the ranges specified by google.protobuf.Timestamp will be rejected.
type TimestampCodec struct{}

const (
	timestampNilValue       = 0xFF
	timestampZeroNanosValue = 0x0
	timestampSecondsMin     = -62135579038
	timestampSecondsMax     = 253402318799
	timestampNanosMax       = 999999999
)

var (
	timestampNilBz       = []byte{timestampNilValue}
	timestampZeroNanosBz = []byte{timestampZeroNanosValue}
)

func (t TimestampCodec) Encode(value protoreflect.Value, w io.Writer) error {
	// nil case
	if !value.IsValid() {
		_, err := w.Write(timestampNilBz)
		return err
	}

	seconds, nanos := getTimestampSecondsAndNanos(value)
	secondsInt := seconds.Int()
	if secondsInt < timestampSecondsMin || secondsInt > timestampSecondsMax {
		return fmt.Errorf("seconds is out of range %d, must be between %d and %d", secondsInt, timestampSecondsMin, timestampSecondsMax)
	}
	secondsInt -= timestampSecondsMin
	var secondsBz [5]byte
	// write the seconds buffer from the end to the front
	for i := 4; i >= 0; i-- {
		secondsBz[i] = byte(secondsInt)
		secondsInt >>= 8
	}
	_, err := w.Write(secondsBz[:])
	if err != nil {
		return err
	}

	nanosInt := nanos.Int()
	if nanosInt == 0 {
		_, err = w.Write(timestampZeroNanosBz)
		return err
	}

	if nanosInt < 0 || nanosInt > timestampNanosMax {
		return fmt.Errorf("nanos is out of range %d, must be between %d and %d", secondsInt, 0, timestampNanosMax)
	}

	var nanosBz [4]byte
	for i := 3; i >= 0; i-- {
		nanosBz[i] = byte(nanosInt)
		nanosInt >>= 8
	}
	nanosBz[0] |= 0xC0
	_, err = w.Write(nanosBz[:])
	return err
}

func (t TimestampCodec) Decode(r Reader) (protoreflect.Value, error) {
	b0, err := r.ReadByte()
	if err != nil {
		return protoreflect.Value{}, err
	}

	if b0 == timestampNilValue {
		return protoreflect.Value{}, nil
	}

	var secondsBz [4]byte
	n, err := r.Read(secondsBz[:])
	if err != nil {
		return protoreflect.Value{}, err
	}
	if n < 4 {
		return protoreflect.Value{}, io.EOF
	}

	seconds := int64(b0)
	for i := 0; i < 4; i++ {
		seconds <<= 8
		seconds |= int64(secondsBz[i])
	}
	seconds += timestampSecondsMin

	msg := timestampMsgType.New()
	msg.Set(timestampSecondsField, protoreflect.ValueOfInt64(seconds))

	b0, err = r.ReadByte()
	if err != nil {
		return protoreflect.Value{}, err
	}

	if b0 == timestampZeroNanosValue {
		return protoreflect.ValueOfMessage(msg), nil
	}

	var nanosBz [3]byte
	n, err = r.Read(nanosBz[:])
	if err != nil {
		return protoreflect.Value{}, err
	}
	if n < 3 {
		return protoreflect.Value{}, io.EOF
	}

	nanos := int32(b0) & 0x3F // clear first two bits
	for i := 0; i < 3; i++ {
		nanos <<= 8
		nanos |= int32(nanosBz[i])
	}

	msg.Set(timestampNanosField, protoreflect.ValueOfInt32(nanos))
	return protoreflect.ValueOfMessage(msg), nil
}

func (t TimestampCodec) Compare(v1, v2 protoreflect.Value) int {
	if !v1.IsValid() {
		if !v2.IsValid() {
			return 0
		}
		return 1
	}

	if !v2.IsValid() {
		return -1
	}

	s1, n1 := getTimestampSecondsAndNanos(v1)
	s2, n2 := getTimestampSecondsAndNanos(v2)
	c := compareInt(s1, s2)
	if c != 0 {
		return c
	}

	return compareInt(n1, n2)
}

func (t TimestampCodec) IsOrdered() bool {
	return true
}

func (t TimestampCodec) FixedBufferSize() int {
	return 9
}

func (t TimestampCodec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return 9, nil
}

// TimestampV0Codec encodes a google.protobuf.Timestamp value as 12 bytes using
// Int64Codec for seconds followed by Int32Codec for nanos. This type does not
// encode nil values correctly, but is retained in order to allow users of the
// previous encoding to successfully migrate from this encoding to the new encoding
// specified by TimestampCodec.
type TimestampV0Codec struct{}

var (
	timestampSecondsField = timestampMsgType.Descriptor().Fields().ByName("seconds")
	timestampNanosField   = timestampMsgType.Descriptor().Fields().ByName("nanos")
)

func getTimestampSecondsAndNanos(value protoreflect.Value) (protoreflect.Value, protoreflect.Value) {
	msg := value.Message()
	return msg.Get(timestampSecondsField), msg.Get(timestampNanosField)
}

func (t TimestampV0Codec) Decode(r Reader) (protoreflect.Value, error) {
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

func (t TimestampV0Codec) Encode(value protoreflect.Value, w io.Writer) error {
	seconds, nanos := getTimestampSecondsAndNanos(value)
	err := int64Codec.Encode(seconds, w)
	if err != nil {
		return err
	}
	return int32Codec.Encode(nanos, w)
}

func (t TimestampV0Codec) Compare(v1, v2 protoreflect.Value) int {
	s1, n1 := getTimestampSecondsAndNanos(v1)
	s2, n2 := getTimestampSecondsAndNanos(v2)
	c := compareInt(s1, s2)
	if c != 0 {
		return c
	}

	return compareInt(n1, n2)
}

func (t TimestampV0Codec) IsOrdered() bool {
	return true
}

func (t TimestampV0Codec) FixedBufferSize() int {
	return 12
}

func (t TimestampV0Codec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return t.FixedBufferSize(), nil
}
