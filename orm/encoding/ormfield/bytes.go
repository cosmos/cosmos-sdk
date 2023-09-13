package ormfield

import (
	"bytes"
	"encoding/binary"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// BytesCodec encodes bytes as raw bytes. It errors if the byte array is longer
// than 255 bytes.
type BytesCodec struct{}

func (b BytesCodec) FixedBufferSize() int {
	return -1
}

// ComputeBufferSize returns the bytes size of the value.
func (b BytesCodec) ComputeBufferSize(value protoreflect.Value) (int, error) {
	return bytesSize(value), nil
}

func bytesSize(value protoreflect.Value) int {
	if !value.IsValid() {
		return 0
	}
	return len(value.Bytes())
}

func (b BytesCodec) IsOrdered() bool {
	return false
}

func (b BytesCodec) Decode(r Reader) (protoreflect.Value, error) {
	bz, err := io.ReadAll(r)
	return protoreflect.ValueOfBytes(bz), err
}

func (b BytesCodec) Encode(value protoreflect.Value, w io.Writer) error {
	if !value.IsValid() {
		return nil
	}
	_, err := w.Write(value.Bytes())
	return err
}

func (b BytesCodec) Compare(v1, v2 protoreflect.Value) int {
	return compareBytes(v1, v2)
}

// NonTerminalBytesCodec encodes bytes as raw bytes length prefixed by a single
// byte. It errors if the byte array is longer than 255 bytes.
type NonTerminalBytesCodec struct{}

func (b NonTerminalBytesCodec) FixedBufferSize() int {
	return -1
}

// ComputeBufferSize returns the bytes size of the value plus the length of the
// varint length-prefix.
func (b NonTerminalBytesCodec) ComputeBufferSize(value protoreflect.Value) (int, error) {
	n := bytesSize(value)
	prefixLen := 1
	// we use varint, if the first bit of a byte is 1 then we need to signal continuation
	for n >= 0x80 {
		prefixLen++
		n >>= 7
	}
	return n + prefixLen, nil
}

func (b NonTerminalBytesCodec) IsOrdered() bool {
	return false
}

func (b NonTerminalBytesCodec) Compare(v1, v2 protoreflect.Value) int {
	return compareBytes(v1, v2)
}

func (b NonTerminalBytesCodec) Decode(r Reader) (protoreflect.Value, error) {
	n, err := binary.ReadUvarint(r)
	if err != nil {
		return protoreflect.Value{}, err
	}

	if n == 0 {
		return protoreflect.ValueOfBytes([]byte{}), nil
	}

	bz := make([]byte, n)
	_, err = r.Read(bz)
	return protoreflect.ValueOfBytes(bz), err
}

func (b NonTerminalBytesCodec) Encode(value protoreflect.Value, w io.Writer) error {
	var bz []byte
	if value.IsValid() {
		bz = value.Bytes()
	}
	n := len(bz)
	var prefix [binary.MaxVarintLen64]byte
	prefixLen := binary.PutUvarint(prefix[:], uint64(n))
	_, err := w.Write(prefix[:prefixLen])
	if err != nil {
		return err
	}
	_, err = w.Write(bz)
	return err
}

func compareBytes(v1, v2 protoreflect.Value) int {
	var bz1, bz2 []byte
	if v1.IsValid() {
		bz1 = v1.Bytes()
	}
	if v2.IsValid() {
		bz2 = v2.Bytes()
	}
	return bytes.Compare(bz1, bz2)
}
