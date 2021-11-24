package ormfield

import (
	"bytes"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
)

// BytesCodec encodes bytes as raw bytes. It errors if the byte array is longer
// than 255 bytes.
type BytesCodec struct{}

func (b BytesCodec) FixedBufferSize() int {
	return -1
}

func (b BytesCodec) ComputeBufferSize(value protoreflect.Value) (int, error) {
	return bytesSize(value)
}

func bytesSize(value protoreflect.Value) (int, error) {
	bz := value.Bytes()
	n := len(bz)
	if n > 255 {
		return -1, ormerrors.BytesFieldTooLong
	}
	return n, nil
}

func (b BytesCodec) IsOrdered() bool {
	return false
}

func (b BytesCodec) Decode(r Reader) (protoreflect.Value, error) {
	bz, err := io.ReadAll(r)
	return protoreflect.ValueOfBytes(bz), err
}

func (b BytesCodec) Encode(value protoreflect.Value, w io.Writer) error {
	_, err := w.Write(value.Bytes())
	return err
}

func (b BytesCodec) Compare(v1, v2 protoreflect.Value) int {
	return bytes.Compare(v1.Bytes(), v2.Bytes())
}

// NonTerminalBytesCodec encodes bytes as raw bytes length prefixed by a single
// byte. It errors if the byte array is longer than 255 bytes.
type NonTerminalBytesCodec struct{}

func (b NonTerminalBytesCodec) FixedBufferSize() int {
	return -1
}

func (b NonTerminalBytesCodec) ComputeBufferSize(value protoreflect.Value) (int, error) {
	n, err := bytesSize(value)
	return n + 1, err
}

func (b NonTerminalBytesCodec) IsOrdered() bool {
	return false
}

func (b NonTerminalBytesCodec) Compare(v1, v2 protoreflect.Value) int {
	return bytes.Compare(v1.Bytes(), v2.Bytes())
}

func (b NonTerminalBytesCodec) Decode(r Reader) (protoreflect.Value, error) {
	n, err := r.ReadByte()
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
	bz := value.Bytes()
	n := len(bz)
	if n > 255 {
		return ormerrors.BytesFieldTooLong
	}
	_, err := w.Write([]byte{byte(n)})
	if err != nil {
		return err
	}
	_, err = w.Write(bz)
	return err
}
