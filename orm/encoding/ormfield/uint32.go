package ormfield

import (
	"encoding/binary"
	"errors"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// FixedUint32Codec encodes uint32 values as 4-byte big-endian integers.
type FixedUint32Codec struct{}

func (u FixedUint32Codec) FixedBufferSize() int {
	return 4
}

func (u FixedUint32Codec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return u.FixedBufferSize(), nil
}

func (u FixedUint32Codec) IsOrdered() bool {
	return true
}

func (u FixedUint32Codec) Compare(v1, v2 protoreflect.Value) int {
	return compareUint(v1, v2)
}

func (u FixedUint32Codec) Decode(r Reader) (protoreflect.Value, error) {
	var x uint32
	err := binary.Read(r, binary.BigEndian, &x)
	return protoreflect.ValueOfUint32(x), err
}

func (u FixedUint32Codec) Encode(value protoreflect.Value, w io.Writer) error {
	var x uint64
	if value.IsValid() {
		x = value.Uint()
	}
	return binary.Write(w, binary.BigEndian, uint32(x))
}

// CompactUint32Codec encodes uint32 values using EncodeCompactUint32.
type CompactUint32Codec struct{}

func (c CompactUint32Codec) Decode(r Reader) (protoreflect.Value, error) {
	x, err := DecodeCompactUint32(r)
	return protoreflect.ValueOfUint32(x), err
}

func (c CompactUint32Codec) Encode(value protoreflect.Value, w io.Writer) error {
	var x uint64
	if value.IsValid() {
		x = value.Uint()
	}
	_, err := w.Write(EncodeCompactUint32(uint32(x)))
	return err
}

func (c CompactUint32Codec) Compare(v1, v2 protoreflect.Value) int {
	return compareUint(v1, v2)
}

func (c CompactUint32Codec) IsOrdered() bool {
	return true
}

func (c CompactUint32Codec) FixedBufferSize() int {
	return 5
}

func (c CompactUint32Codec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return c.FixedBufferSize(), nil
}

// EncodeCompactUint32 encodes uint32 values in 2,3,4 or 5 bytes.
// Unlike regular varints, this encoding is
// suitable for ordered prefix scans. The length of the output + 2 is encoded
// in the first 2 bits of the first byte and the remaining bits encoded with
// big-endian ordering.
// Values less than 2^14 fill fit in 2 bytes, values less than 2^22 will
// fit in 3, and values less than 2^30 will fit in 4.
func EncodeCompactUint32(x uint32) []byte {
	switch {
	case x < 16384: // 2^14
		buf := make([]byte, 2)
		buf[0] = byte(x >> 8)
		buf[1] = byte(x)
		return buf
	case x < 4194304: // 2^22
		buf := make([]byte, 3)
		buf[0] = 0x40
		buf[0] |= byte(x >> 16)
		buf[1] = byte(x >> 8)
		buf[2] = byte(x)
		return buf
	case x < 1073741824: // 2^30
		buf := make([]byte, 4)
		buf[0] = 0x80
		buf[0] |= byte(x >> 24)
		buf[1] = byte(x >> 16)
		buf[2] = byte(x >> 8)
		buf[3] = byte(x)
		return buf
	default:
		buf := make([]byte, 5)
		buf[0] = 0xC0
		buf[0] |= byte(x >> 26)
		buf[1] = byte(x >> 18)
		buf[2] = byte(x >> 10)
		buf[3] = byte(x >> 2)
		buf[4] = byte(x) & 0x3
		return buf
	}
}

// DecodeCompactUint32 decodes a uint32 encoded with EncodeCompactU32.
func DecodeCompactUint32(reader io.Reader) (uint32, error) {
	var buf [5]byte

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

		x := uint32(buf[0]) << 8
		x |= uint32(buf[1])
		return x, nil
	case 1:
		n, err := reader.Read(buf[1:3])
		if err != nil {
			return 0, err
		}
		if n < 2 {
			return 0, io.ErrUnexpectedEOF
		}

		x := (uint32(buf[0]) & 0x3F) << 16
		x |= uint32(buf[1]) << 8
		x |= uint32(buf[2])
		return x, nil
	case 2:
		n, err := reader.Read(buf[1:4])
		if err != nil {
			return 0, err
		}
		if n < 3 {
			return 0, io.ErrUnexpectedEOF
		}

		x := (uint32(buf[0]) & 0x3F) << 24
		x |= uint32(buf[1]) << 16
		x |= uint32(buf[2]) << 8
		x |= uint32(buf[3])
		return x, nil
	case 3:
		n, err := reader.Read(buf[1:5])
		if err != nil {
			return 0, err
		}
		if n < 4 {
			return 0, io.ErrUnexpectedEOF
		}

		x := (uint32(buf[0]) & 0x3F) << 26
		x |= uint32(buf[1]) << 18
		x |= uint32(buf[2]) << 10
		x |= uint32(buf[3]) << 2
		x |= uint32(buf[4])
		return x, nil
	default:
		return 0, errors.New("unexpected case")
	}
}
