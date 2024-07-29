package ormfield

import (
	"fmt"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// TimestampCodec encodes google.protobuf.Timestamp values with the following
// encoding:
//   - nil is encoded as []byte{0xFF}
//   - seconds (which can range from 0001-01-01T00:00:00Z to 9999-12-31T23:59:59Z) is encoded as 5 fixed bytes
//   - nanos (which can range from 0 to 999,999,999 or -999,999,999 to 0 if seconds is negative) are encoded such
//     that 999,999,999 is always added to nanos. This ensures that the encoded nanos are always >= 0. Additionally,
//     by adding 999,999,999 to both positive and negative nanos, we guarantee that the lexicographical order is
//     preserved when comparing the encoded values of two Timestamps.
//
// When iterating over timestamp indexes, nil values will always be ordered last.
//
// Values for seconds and nanos outside the ranges specified by google.protobuf.Timestamp will be rejected.
type TimestampCodec struct{}

const (
	timestampDurationNilValue             = 0xFF
	timestampDurationZeroNanosValue       = 0x0
	timestampDurationBufferSize           = 9
	TimestampSecondsMin             int64 = -62135596800
	TimestampSecondsMax             int64 = 253402300799
	TimestampNanosMax                     = 999999999
)

var (
	timestampDurationNilBz = []byte{timestampDurationNilValue}
	timestampZeroNanosBz   = []byte{timestampDurationZeroNanosValue}
)

func (t TimestampCodec) Encode(value protoreflect.Value, w io.Writer) error {
	// nil case
	if !value.IsValid() {
		_, err := w.Write(timestampDurationNilBz)
		return err
	}

	seconds, nanos := getTimestampSecondsAndNanos(value)
	secondsInt := seconds.Int()
	if secondsInt < TimestampSecondsMin || secondsInt > TimestampSecondsMax {
		return fmt.Errorf("timestamp seconds is out of range %d, must be between %d and %d", secondsInt, TimestampSecondsMin, TimestampSecondsMax)
	}
	secondsInt -= TimestampSecondsMin
	err := encodeSeconds(secondsInt, w)
	if err != nil {
		return err
	}

	nanosInt := nanos.Int()
	if nanosInt == 0 {
		_, err = w.Write(timestampZeroNanosBz)
		return err
	}

	if nanosInt < 0 || nanosInt > TimestampNanosMax {
		return fmt.Errorf("timestamp nanos is out of range %d, must be between %d and %d", secondsInt, 0, TimestampNanosMax)
	}

	return encodeNanos(nanosInt, w)
}

func encodeSeconds(secondsInt int64, w io.Writer) error {
	var secondsBz [5]byte
	// write the seconds buffer from the end to the front
	for i := 4; i >= 0; i-- {
		secondsBz[i] = byte(secondsInt)
		secondsInt >>= 8
	}
	_, err := w.Write(secondsBz[:])
	return err
}

func encodeNanos(nanosInt int64, w io.Writer) error {
	var nanosBz [4]byte
	for i := 3; i >= 0; i-- {
		nanosBz[i] = byte(nanosInt)
		nanosInt >>= 8
	}

	// This condition is crucial to ensure the function's correct behavior when dealing with a Timestamp or Duration encoding.
	// Specifically, this function is bypassed for Timestamp values when their nanoseconds part is zero.
	// In the decodeNanos function, there's a preliminary check for a zero first byte, which represents all values ≤ 16777215 (00000000 11111111 11111111 11111111).
	// Without this adjustment (setting the first byte to 0x80 with is 10000000 in binary format), decodeNanos would incorrectly return 0 for any number ≤ 16777215,
	// leading to inaccurate decoding of nanoseconds.
	nanosBz[0] |= 0x80
	_, err := w.Write(nanosBz[:])
	return err
}

func (t TimestampCodec) Decode(r Reader) (protoreflect.Value, error) {
	isNil, seconds, err := decodeSeconds(r)
	if isNil || err != nil {
		return protoreflect.Value{}, err
	}

	seconds += TimestampSecondsMin

	msg := timestampMsgType.New()
	msg.Set(timestampSecondsField, protoreflect.ValueOfInt64(seconds))

	nanos, err := decodeNanos(r)
	if err != nil {
		return protoreflect.Value{}, err
	}

	if nanos == 0 {
		return protoreflect.ValueOfMessage(msg), nil
	}

	msg.Set(timestampNanosField, protoreflect.ValueOfInt32(nanos))
	return protoreflect.ValueOfMessage(msg), nil
}

func decodeSeconds(r Reader) (isNil bool, seconds int64, err error) {
	b0, err := r.ReadByte()
	if err != nil {
		return false, 0, err
	}

	if b0 == timestampDurationNilValue {
		return true, 0, nil
	}

	var secondsBz [4]byte
	n, err := r.Read(secondsBz[:])
	if err != nil {
		return false, 0, err
	}
	if n < 4 {
		return false, 0, io.EOF
	}

	seconds = int64(b0)
	for i := 0; i < 4; i++ {
		seconds <<= 8
		seconds |= int64(secondsBz[i])
	}

	return false, seconds, nil
}

func decodeNanos(r Reader) (int32, error) {
	b0, err := r.ReadByte()
	if err != nil {
		return 0, err
	}

	if b0 == timestampDurationZeroNanosValue {
		return 0, nil
	}

	var nanosBz [3]byte
	n, err := r.Read(nanosBz[:])
	if err != nil {
		return 0, err
	}
	if n < 3 {
		return 0, io.EOF
	}

	// Clear the first bit, previously set in encodeNanos, to ensure this logic is applied
	// and for numbers ≤ 16777215. This adjustment guarantees that we accurately interpret
	// the value as intended when encoding smaller numbers.
	nanos := int32(b0) & 0x7F

	for i := 0; i < 3; i++ {
		nanos <<= 8
		nanos |= int32(nanosBz[i])
	}

	return nanos, nil
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
	return timestampDurationBufferSize
}

func (t TimestampCodec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return timestampDurationBufferSize, nil
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

type TimestampV1Codec struct{}

func (t TimestampV1Codec) Encode(value protoreflect.Value, w io.Writer) error {
	// nil case
	if !value.IsValid() {
		_, err := w.Write(timestampDurationNilBz)
		return err
	}

	seconds, nanos := getTimestampSecondsAndNanos(value)
	secondsInt := seconds.Int()
	if secondsInt < TimestampSecondsMin || secondsInt > TimestampSecondsMax {
		return fmt.Errorf("timestamp seconds is out of range %d, must be between %d and %d", secondsInt, TimestampSecondsMin, TimestampSecondsMax)
	}
	secondsInt -= TimestampSecondsMin
	err := encodeSeconds(secondsInt, w)
	if err != nil {
		return err
	}

	nanosInt := nanos.Int()
	if nanosInt == 0 {
		_, err = w.Write(timestampZeroNanosBz)
		return err
	}

	if nanosInt < 0 || nanosInt > TimestampNanosMax {
		return fmt.Errorf("timestamp nanos is out of range %d, must be between %d and %d", secondsInt, 0, TimestampNanosMax)
	}

	return encodeNanosV1(nanosInt, w)
}

func (t TimestampV1Codec) Decode(r Reader) (protoreflect.Value, error) {
	isNil, seconds, err := decodeSeconds(r)
	if isNil || err != nil {
		return protoreflect.Value{}, err
	}

	seconds += TimestampSecondsMin

	msg := timestampMsgType.New()
	msg.Set(timestampSecondsField, protoreflect.ValueOfInt64(seconds))

	nanos, err := decodeNanosV1(r)
	if err != nil {
		return protoreflect.Value{}, err
	}

	if nanos == 0 {
		return protoreflect.ValueOfMessage(msg), nil
	}

	msg.Set(timestampNanosField, protoreflect.ValueOfInt32(nanos))
	return protoreflect.ValueOfMessage(msg), nil
}

func (t TimestampV1Codec) Compare(v1, v2 protoreflect.Value) int {
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

func (t TimestampV1Codec) IsOrdered() bool {
	return true
}

func (t TimestampV1Codec) FixedBufferSize() int {
	return timestampDurationBufferSize
}

func (t TimestampV1Codec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return timestampDurationBufferSize, nil
}

func encodeNanosV1(nanosInt int64, w io.Writer) error {
	var nanosBz [4]byte
	for i := 3; i >= 0; i-- {
		nanosBz[i] = byte(nanosInt)
		nanosInt >>= 8
	}
	nanosBz[0] |= 0xC0
	_, err := w.Write(nanosBz[:])
	return err
}

func decodeNanosV1(r Reader) (int32, error) {
	b0, err := r.ReadByte()
	if err != nil {
		return 0, err
	}

	if b0 == timestampDurationZeroNanosValue {
		return 0, nil
	}

	var nanosBz [3]byte
	n, err := r.Read(nanosBz[:])
	if err != nil {
		return 0, err
	}
	if n < 3 {
		return 0, io.EOF
	}

	nanos := int32(b0) & 0x3F // clear first two bits
	for i := 0; i < 3; i++ {
		nanos <<= 8
		nanos |= int32(nanosBz[i])
	}

	return nanos, nil
}
