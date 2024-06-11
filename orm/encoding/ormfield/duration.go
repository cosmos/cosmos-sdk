package ormfield

import (
	"fmt"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	DurationSecondsMin int64 = -315576000000
	DurationSecondsMax int64 = 315576000000
	DurationNanosMin         = -999999999
	DurationNanosMax         = 999999999
)

// DurationCodec encodes google.protobuf.Duration values with the following
// encoding:
//   - nil is encoded as []byte{0xFF}
//   - seconds (which can range from -315,576,000,000 to +315,576,000,000) is encoded as 5 fixed bytes
//   - nanos (which can range from 0 to 999,999,999 or -999,999,999 to 0 if seconds is negative) are encoded such
//     that 999,999,999 is always added to nanos. This ensures that the encoded nanos are always >= 0. Additionally,
//     by adding 999,999,999 to both positive and negative nanos, we guarantee that the lexicographical order is
//     preserved when comparing the encoded values of two Durations:
//   - []byte{0xBB, 0x9A, 0xC9, 0xFF} for zero nanos
//   - 4 fixed bytes with the bit mask 0x80 applied to the first byte, with negative nanos scaled so that -999,999,999
//     is encoded as 0 and -1 is encoded as 999,999,998
//
// When iterating over timestamp indexes, nil values will always be ordered last.
//
// Values for seconds and nanos outside the ranges specified by google.protobuf.Duration will be rejected.
type DurationCodec struct{}

func (d DurationCodec) Encode(value protoreflect.Value, w io.Writer) error {
	// nil case
	if !value.IsValid() {
		_, err := w.Write(timestampDurationNilBz)
		return err
	}

	seconds, nanos := getDurationSecondsAndNanos(value)
	secondsInt := seconds.Int()
	nanosInt := nanos.Int()

	if err := validateDurationRanges(secondsInt, nanosInt); err != nil {
		return err
	}

	// we subtract the min duration value to make sure secondsInt is always non-negative and starts at 0.
	secondsInt -= DurationSecondsMin
	err := encodeSeconds(secondsInt, w)
	if err != nil {
		return err
	}

	// we subtract the min duration value to make sure nanosInt is always non-negative and starts at 0.
	nanosInt -= DurationNanosMin
	return encodeNanos(nanosInt, w)
}

func (d DurationCodec) Decode(r Reader) (protoreflect.Value, error) {
	isNil, seconds, err := decodeSeconds(r)
	if isNil || err != nil {
		return protoreflect.Value{}, err
	}

	// we add the min seconds duration value to get back the original value
	seconds += DurationSecondsMin

	msg := durationMsgType.New()
	msg.Set(durationSecondsField, protoreflect.ValueOfInt64(seconds))

	nanos, err := decodeNanos(r)
	if err != nil {
		return protoreflect.Value{}, err
	}
	// we add the min nanos duration value to get back the original value
	nanos += DurationNanosMin

	msg.Set(durationNanosField, protoreflect.ValueOfInt32(nanos))
	return protoreflect.ValueOfMessage(msg), nil
}

func (d DurationCodec) Compare(v1, v2 protoreflect.Value) int {
	if !v1.IsValid() {
		if !v2.IsValid() {
			return 0
		}
		return 1
	}

	if !v2.IsValid() {
		return -1
	}

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
	return timestampDurationBufferSize
}

func (d DurationCodec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return timestampDurationBufferSize, nil
}

var (
	durationSecondsField = durationMsgType.Descriptor().Fields().ByName("seconds")
	durationNanosField   = durationMsgType.Descriptor().Fields().ByName("nanos")
)

func getDurationSecondsAndNanos(value protoreflect.Value) (protoreflect.Value, protoreflect.Value) {
	msg := value.Message()
	return msg.Get(durationSecondsField), msg.Get(durationNanosField)
}

// validateDurationRanges checks whether seconds and nanoseconds are in valid ranges
// for a protobuf Duration type. It ensures that seconds are within the allowed range
// and, if seconds are zero or negative, verifies that nanoseconds are also within
// the valid range. For negative seconds, nanoseconds must be non-positive.
// Parameters:
//   - seconds: The number of seconds component of the duration.
//   - nanos: The number of nanoseconds component of the duration.
//
// Returns:
//   - error: An error indicating if the duration components are out of range.
func validateDurationRanges(seconds, nanos int64) error {
	if seconds < DurationSecondsMin || seconds > DurationSecondsMax {
		return fmt.Errorf("duration seconds is out of range %d, must be between %d and %d", seconds, DurationSecondsMin, DurationSecondsMax)
	}

	if seconds == 0 {
		if nanos < DurationNanosMin || nanos > DurationNanosMax {
			return fmt.Errorf("duration nanos is out of range %d, must be between %d and %d", nanos, DurationNanosMin, DurationNanosMax)
		}
	} else if seconds < 0 {
		if nanos < DurationNanosMin || nanos > 0 {
			return fmt.Errorf("negative duration nanos is out of range %d, must be between %d and %d", nanos, DurationNanosMin, 0)
		}
	} else if nanos < 0 || nanos > DurationNanosMax {
		return fmt.Errorf("duration nanos is out of range %d, must be between %d and %d", nanos, 0, DurationNanosMax)
	}

	return nil
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
	}
	return compareInt(n1, n2)
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

// DurationV1Codec encodes google.protobuf.Duration values with the following
// encoding:
// - nil is encoded as []byte{0xFF}
// - seconds (which can range from -315,576,000,000 to +315,576,000,000) is encoded as 5 fixed bytes
// - nanos (which can range from 0 to 999,999,999 or -999,999,999 to 0 if seconds is negative) is encoded as:
//   - []byte{0x0} for zero nanos
//   - 4 fixed bytes with the bit mask 0xC0 applied to the first byte, with negative nanos scaled so that -999,999,999
//     is encoded as 1 and -1 is encoded as 999,999,999
//
// When iterating over timestamp indexes, nil values will always be ordered last.
//
// Values for seconds and nanos outside the ranges specified by google.protobuf.Duration will be rejected.
type DurationV1Codec struct{}

func (d DurationV1Codec) Encode(value protoreflect.Value, w io.Writer) error {
	// nil case
	if !value.IsValid() {
		_, err := w.Write(timestampDurationNilBz)
		return err
	}

	seconds, nanos := getDurationSecondsAndNanos(value)
	secondsInt := seconds.Int()
	if secondsInt < DurationSecondsMin || secondsInt > DurationSecondsMax {
		return fmt.Errorf("duration seconds is out of range %d, must be between %d and %d", secondsInt, DurationSecondsMin, DurationSecondsMax)
	}
	negative := secondsInt < 0
	// we subtract the min duration value to make sure secondsInt is always non-negative and starts at 0.
	secondsInt -= DurationSecondsMin
	err := encodeSeconds(secondsInt, w)
	if err != nil {
		return err
	}

	nanosInt := nanos.Int()
	if nanosInt == 0 {
		_, err = w.Write(timestampZeroNanosBz)
		return err
	}

	if negative {
		if nanosInt < DurationNanosMin || nanosInt > 0 {
			return fmt.Errorf("negative duration nanos is out of range %d, must be between %d and %d", nanosInt, DurationNanosMin, 0)
		}
		nanosInt = DurationNanosMax + nanosInt + 1
	} else if nanosInt < 0 || nanosInt > DurationNanosMax {
		return fmt.Errorf("duration nanos is out of range %d, must be between %d and %d", nanosInt, 0, DurationNanosMax)
	}

	return encodeNanosV1(nanosInt, w)
}

func (d DurationV1Codec) Decode(r Reader) (protoreflect.Value, error) {
	isNil, seconds, err := decodeSeconds(r)
	if isNil || err != nil {
		return protoreflect.Value{}, err
	}

	// we add the min duration value to get back the original value
	seconds += DurationSecondsMin

	negative := seconds < 0

	msg := durationMsgType.New()
	msg.Set(durationSecondsField, protoreflect.ValueOfInt64(seconds))

	nanos, err := decodeNanosV1(r)
	if err != nil {
		return protoreflect.Value{}, err
	}

	if nanos == 0 {
		return protoreflect.ValueOfMessage(msg), nil
	}

	if negative {
		nanos = nanos - DurationNanosMax - 1
	}

	msg.Set(durationNanosField, protoreflect.ValueOfInt32(nanos))
	return protoreflect.ValueOfMessage(msg), nil
}

func (d DurationV1Codec) Compare(v1, v2 protoreflect.Value) int {
	if !v1.IsValid() {
		if !v2.IsValid() {
			return 0
		}
		return 1
	}

	if !v2.IsValid() {
		return -1
	}

	s1, n1 := getDurationSecondsAndNanos(v1)
	s2, n2 := getDurationSecondsAndNanos(v2)
	c := compareInt(s1, s2)
	if c != 0 {
		return c
	}

	return compareInt(n1, n2)
}

func (d DurationV1Codec) IsOrdered() bool {
	return true
}

func (d DurationV1Codec) FixedBufferSize() int {
	return timestampDurationBufferSize
}

func (d DurationV1Codec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return timestampDurationBufferSize, nil
}
