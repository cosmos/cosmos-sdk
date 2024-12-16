package ormfield_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/reflect/protoreflect"
	"cosmossdk.io/orm/encoding/ormfield"
)

// MockReader implements the Reader interface for testing
type MockReader struct {
	*bytes.Reader
}

func NewMockReader(data []byte) *MockReader {
	return &MockReader{bytes.NewReader(data)}
}

// TestBytesCodecEncodeDecode tests encoding and decoding for BytesCodec
func TestBytesCodecEncodeDecode(t *testing.T) {
	codec := ormfield.BytesCodec{}

	original := protoreflect.ValueOfBytes([]byte("hello"))
	buffer := &bytes.Buffer{}

	// Encode
	err := codec.Encode(original, buffer)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello"), buffer.Bytes())

	// Decode
	decoded, err := codec.Decode(NewMockReader(buffer.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, original.Bytes(), decoded.Bytes())
}

// TestBytesCodecCompare tests comparison for BytesCodec
func TestBytesCodecCompare(t *testing.T) {
	codec := ormfield.BytesCodec{}

	v1 := protoreflect.ValueOfBytes([]byte("abc"))
	v2 := protoreflect.ValueOfBytes([]byte("xyz"))
	v3 := protoreflect.ValueOfBytes([]byte("abc"))

	assert.Equal(t, -1, codec.Compare(v1, v2))
	assert.Equal(t, 0, codec.Compare(v1, v3))
	assert.Equal(t, 1, codec.Compare(v2, v1))
}

// TestBytesCodecComputeBufferSize tests buffer size computation for BytesCodec
func TestBytesCodecComputeBufferSize(t *testing.T) {
	codec := ormfield.BytesCodec{}

	value := protoreflect.ValueOfBytes([]byte("hello"))
	size, err := codec.ComputeBufferSize(value)
	assert.NoError(t, err)
	assert.Equal(t, 5, size)
}

// TestNonTerminalBytesCodecEncodeDecode tests encoding and decoding for NonTerminalBytesCodec
func TestNonTerminalBytesCodecEncodeDecode(t *testing.T) {
	codec := ormfield.NonTerminalBytesCodec{}

	original := protoreflect.ValueOfBytes([]byte("world"))
	buffer := &bytes.Buffer{}

	// Encode
	err := codec.Encode(original, buffer)
	assert.NoError(t, err)

	// Decode
	decoded, err := codec.Decode(NewMockReader(buffer.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, original.Bytes(), decoded.Bytes())
}

// TestNonTerminalBytesCodecComputeBufferSize tests buffer size computation for NonTerminalBytesCodec
func TestNonTerminalBytesCodecComputeBufferSize(t *testing.T) {
	codec := ormfield.NonTerminalBytesCodec{}

	value := protoreflect.ValueOfBytes([]byte("protobuf"))
	size, err := codec.ComputeBufferSize(value)
	assert.NoError(t, err)

	// Length prefix is 1 byte, and "protobuf" is 8 bytes
	assert.Equal(t, 9, size)
}

// TestNonTerminalBytesCodecCompare tests comparison for NonTerminalBytesCodec
func TestNonTerminalBytesCodecCompare(t *testing.T) {
	codec := ormfield.NonTerminalBytesCodec{}

	v1 := protoreflect.ValueOfBytes([]byte("abc"))
	v2 := protoreflect.ValueOfBytes([]byte("xyz"))
	v3 := protoreflect.ValueOfBytes([]byte("abc"))

	assert.Equal(t, -1, codec.Compare(v1, v2))
	assert.Equal(t, 0, codec.Compare(v1, v3))
	assert.Equal(t, 1, codec.Compare(v2, v1))
}

// TestNonTerminalBytesCodecHandlesEmptyBytes tests handling of empty byte arrays
func TestNonTerminalBytesCodecHandlesEmptyBytes(t *testing.T) {
	codec := ormfield.NonTerminalBytesCodec{}
	buffer := &bytes.Buffer{}

	// Encode empty bytes
	err := codec.Encode(protoreflect.ValueOfBytes([]byte{}), buffer)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0}, buffer.Bytes())

	// Decode empty bytes
	decoded, err := codec.Decode(NewMockReader(buffer.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, []byte{}, decoded.Bytes())
}

// TestNonTerminalBytesCodecHandlesLargeInput tests handling of large input byte arrays
func TestNonTerminalBytesCodecHandlesLargeInput(t *testing.T) {
	codec := ormfield.NonTerminalBytesCodec{}
	buffer := &bytes.Buffer{}

	// Large input (length 300 bytes)
	input := make([]byte, 300)
	for i := range input {
		input[i] = byte(i % 256)
	}

	err := codec.Encode(protoreflect.ValueOfBytes(input), buffer)
	assert.NoError(t, err)

	decoded, err := codec.Decode(NewMockReader(buffer.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, input, decoded.Bytes())
}
