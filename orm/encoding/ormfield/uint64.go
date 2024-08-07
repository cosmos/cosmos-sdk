package ormfield

import (
	"encoding/binary"
	"errors"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// FixedUint64Codec encodes uint64 values as 8-byte big-endian integers.
type FixedUint64Codec struct{}

func (u FixedUint64Codec) FixedBufferSize() int {
	return 8
}

func (u FixedUint64Codec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return u.FixedBufferSize(), nil
}

func (u FixedUint64Codec) IsOrdered() bool {
	return true
}

func (u FixedUint64Codec) Compare(v1, v2 protoreflect.Value) int {
	return compareUint(v1, v2)
}

func (u FixedUint64Codec) Decode(r Reader) (protoreflect.Value, error) {
	var x uint64
	err := binary.Read(r, binary.BigEndian, &x)
	return protoreflect.ValueOfUint64(x), err
}

func (u FixedUint64Codec) Encode(value protoreflect.Value, w io.Writer) error {
	var x uint64
	if value.IsValid() {
		x = value.Uint()
	}
	return binary.Write(w, binary.BigEndian, x)
}

func compareUint(v1, v2 protoreflect.Value) int {
	var x, y uint64
	if v1.IsValid() {
		x = v1.Uint()
	}
	if v2.IsValid() {
		y = v2.Uint()
	}
	switch {
	case x == y:
		return 0
	case x < y:
		return -1
	default:
		return 1
	}
}

// CompactUint64Codec encodes uint64 values using EncodeCompactUint64.
type CompactUint64Codec struct{}

func (c CompactUint64Codec) Decode(r Reader) (protoreflect.Value, error) {
	x, err := DecodeCompactUint64(r)
	return protoreflect.ValueOfUint64(x), err
}

func (c CompactUint64Codec) Encode(value protoreflect.Value, w io.Writer) error {
	var x uint64
	if value.IsValid() {
		x = value.Uint()
	}
	_, err := w.Write(EncodeCompactUint64(x))
	return err
}

func (c CompactUint64Codec) Compare(v1, v2 protoreflect.Value) int {
	return compareUint(v1, v2)
}

func (c CompactUint64Codec) IsOrdered() bool {
	return true
}

func (c CompactUint64Codec) FixedBufferSize() int {
	return 9
}

func (c CompactUint64Codec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return c.FixedBufferSize(), nil
}

// EncodeCompactUint64 encodes uint64 values in 2,4,6 or 9 bytes.
// Unlike regular varints, this encoding is
// suitable for ordered prefix scans. The first two bits of the first byte
// indicate the length of the buffer - 00 for 2, 01 for 4, 10 for 6 and
// 11 for 9. The remaining bits are encoded with big-endian ordering.
// Values less than 2^14 fill fit in 2 bytes, values less than 2^30 will
// fit in 4, and values less than 2^46 will fit in 6.
func EncodeCompactUint64(x uint64) []byte {
	switch {
	case x < 16384: // 2^14
		buf := make([]byte, 2)
		buf[0] = byte(x >> 8)
		buf[1] = byte(x)
		return buf
	case x < 1073741824: // 2^30
		buf := make([]byte, 4)
		buf[0] = 0x40
		buf[0] |= byte(x >> 24)
		buf[1] = byte(x >> 16)
		buf[2] = byte(x >> 8)
		buf[3] = byte(x)
		return buf
	case x < 70368744177664: // 2^46
		buf := make([]byte, 6)
		buf[0] = 0x80
		buf[0] |= byte(x >> 40)
		buf[1] = byte(x >> 32)
		buf[2] = byte(x >> 24)
		buf[3] = byte(x >> 16)
		buf[4] = byte(x >> 8)
		buf[5] = byte(x)
		return buf
	default:
		buf := make([]byte, 9)
		buf[0] = 0xC0
		buf[0] |= byte(x >> 58)
		buf[1] = byte(x >> 50)
		buf[2] = byte(x >> 42)
		buf[3] = byte(x >> 34)
		buf[4] = byte(x >> 26)
		buf[5] = byte(x >> 18)
		buf[6] = byte(x >> 10)
		buf[7] = byte(x >> 2)
		buf[8] = byte(x) & 0x3
		return buf
	}
}

func DecodeCompactUint64(reader io.Reader) (uint64, error) {
	var buf [9]byte
	n, err := reader.Read(buf[:1])
	if err != nil {
		return 0, err
	}
	if n < 1 {
		return 0, io.ErrUnexpectedEOF
	}

	switch buf[0] >> 6 {
	case 0:
		n, err := reader.Read(buf[1:2])
		if err != nil {
			return 0, err
		}
		if n < 1 {
			return 0, io.ErrUnexpectedEOF
		}

		x := uint64(buf[0]) << 8
		x |= uint64(buf[1])
		return x, nil
	case 1:
		n, err := reader.Read(buf[1:4])
		if err != nil {
			return 0, err
		}
		if n < 3 {
			return 0, io.ErrUnexpectedEOF
		}

		x := (uint64(buf[0]) & 0x3F) << 24
		x |= uint64(buf[1]) << 16
		x |= uint64(buf[2]) << 8
		x |= uint64(buf[3])
		return x, nil
	case 2:
		n, err := reader.Read(buf[1:6])
		if err != nil {
			return 0, err
		}
		if n < 5 {
			return 0, io.ErrUnexpectedEOF
		}

		x := (uint64(buf[0]) & 0x3F) << 40
		x |= uint64(buf[1]) << 32
		x |= uint64(buf[2]) << 24
		x |= uint64(buf[3]) << 16
		x |= uint64(buf[4]) << 8
		x |= uint64(buf[5])
		return x, nil
	case 3:
		n, err := reader.Read(buf[1:9])
		if err != nil {
			return 0, err
		}
		if n < 8 {
			return 0, io.ErrUnexpectedEOF
		}

		x := (uint64(buf[0]) & 0x3F) << 58
		x |= uint64(buf[1]) << 50
		x |= uint64(buf[2]) << 42
		x |= uint64(buf[3]) << 34
		x |= uint64(buf[4]) << 26
		x |= uint64(buf[5]) << 18
		x |= uint64(buf[6]) << 10
		x |= uint64(buf[7]) << 2
		x |= uint64(buf[8])
		return x, nil
	default:
		return 0, errors.New("unexpected case")
	}
}
