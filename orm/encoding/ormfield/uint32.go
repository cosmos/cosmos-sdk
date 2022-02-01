package ormfield

import (
	"encoding/binary"
	"fmt"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// Uint32Codec encodes uint32 values as 4-byte big-endian integers.
type Uint32Codec struct{}

func (u Uint32Codec) FixedBufferSize() int {
	return 4
}

func (u Uint32Codec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return u.FixedBufferSize(), nil
}

func (u Uint32Codec) IsOrdered() bool {
	return true
}

func (u Uint32Codec) Compare(v1, v2 protoreflect.Value) int {
	return compareUint(v1, v2)
}

func (u Uint32Codec) Decode(r Reader) (protoreflect.Value, error) {
	var x uint32
	err := binary.Read(r, binary.BigEndian, &x)
	return protoreflect.ValueOfUint32(x), err
}

func (u Uint32Codec) Encode(value protoreflect.Value, w io.Writer) error {
	return binary.Write(w, binary.BigEndian, uint32(value.Uint()))
}

type CompactUint32Codec struct{}

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

func DecodeCompactU32(reader io.Reader) (uint32, error) {
	var buf [5]byte
	n, err := reader.Read(buf[:])
	if err != nil {
		return 0, err
	}
	if n < 2 {
		return 0, io.ErrUnexpectedEOF
	}

	switch buf[0] >> 6 {
	case 0:
		x := uint32(buf[0]) << 8
		x |= uint32(buf[1])
		return x, nil
	case 1:
		if n < 3 {
			return 0, io.ErrUnexpectedEOF
		}
		x := (uint32(buf[0]) & 0x3F) << 16
		x |= uint32(buf[1]) << 8
		x |= uint32(buf[2])
		return x, nil
	case 2:
		if n < 4 {
			return 0, io.ErrUnexpectedEOF
		}
		x := (uint32(buf[0]) & 0x3F) << 24
		x |= uint32(buf[1]) << 16
		x |= uint32(buf[2]) << 8
		x |= uint32(buf[3])
		return x, nil
	case 3:
		if n < 5 {
			return 0, io.ErrUnexpectedEOF
		}
		x := (uint32(buf[0]) & 0x3F) << 26
		x |= uint32(buf[1]) << 18
		x |= uint32(buf[2]) << 10
		x |= uint32(buf[3]) << 2
		x |= uint32(buf[4])
		return x, nil
	default:
		return 0, fmt.Errorf("unexpected case")
	}
}
