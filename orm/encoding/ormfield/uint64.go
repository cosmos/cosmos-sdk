package ormfield

import (
	"encoding/binary"
	"fmt"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// Uint64Codec encodes uint64 values as 8-byte big-endian integers.
type Uint64Codec struct{}

func (u Uint64Codec) FixedBufferSize() int {
	return 8
}

func (u Uint64Codec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return u.FixedBufferSize(), nil
}

func (u Uint64Codec) IsOrdered() bool {
	return true
}

func (u Uint64Codec) Compare(v1, v2 protoreflect.Value) int {
	return compareUint(v1, v2)
}

func (u Uint64Codec) Decode(r Reader) (protoreflect.Value, error) {
	var x uint64
	err := binary.Read(r, binary.BigEndian, &x)
	return protoreflect.ValueOfUint64(x), err
}

func (u Uint64Codec) Encode(value protoreflect.Value, w io.Writer) error {
	return binary.Write(w, binary.BigEndian, value.Uint())
}

func compareUint(v1, v2 protoreflect.Value) int {
	x := v1.Uint()
	y := v2.Uint()
	if x == y {
		return 0
	} else if x < y {
		return -1
	} else {
		return 1
	}
}

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

func DecodeCompactU64(reader io.Reader) (uint64, error) {
	var buf [9]byte
	n, err := reader.Read(buf[:])
	if err != nil {
		return 0, err
	}
	if n < 2 {
		return 0, io.ErrUnexpectedEOF
	}

	switch buf[0] >> 6 {
	case 0:
		x := uint64(buf[0]) << 8
		x |= uint64(buf[1])
		return x, nil
	case 1:
		if n < 4 {
			return 0, io.ErrUnexpectedEOF
		}
		x := (uint64(buf[0]) & 0x3F) << 24
		x |= uint64(buf[1]) << 16
		x |= uint64(buf[2]) << 8
		x |= uint64(buf[3])
		return x, nil
	case 2:
		if n < 6 {
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
		if n < 9 {
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
		return 0, fmt.Errorf("unexpected case")
	}
}
