package codec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

// Marshaler defines the interface module codecs must implement in order to support
// backwards compatibility with Amino while allowing custom serialization. The
// implementing type can extend this with ProtoMarshaler to support and utilize
// protocol buffers.
type Marshaler interface {
	MarshalBinaryBare(o interface{}) ([]byte, error)
	MustMarshalBinaryBare(o interface{}) []byte

	MarshalBinaryLengthPrefixed(o interface{}) ([]byte, error)
	MustMarshalBinaryLengthPrefixed(o interface{}) []byte

	UnmarshalBinaryBare(bz []byte, ptr interface{}) error
	MustUnmarshalBinaryBare(bz []byte, ptr interface{})

	UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error
	MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{})

	MarshalJSON(o interface{}) ([]byte, error) // nolint: stdmethods
	MustMarshalJSON(o interface{}) []byte

	UnmarshalJSON(bz []byte, ptr interface{}) error // nolint: stdmethods
	MustUnmarshalJSON(bz []byte, ptr interface{})
}

// ProtoMarshaler defines an interface a type must implement as protocol buffer
// defined message.
type ProtoMarshaler interface {
	proto.Message // for JSON serialization

	Marshal() ([]byte, error)
	MarshalTo(data []byte) (n int, err error)
	MarshalToSizedBuffer(dAtA []byte) (int, error)
	Size() int
	Unmarshal(data []byte) error
}

// BaseCodec defines a codec implementing the Marshaler interface that allows
// for Protocol buffer and Amino-based serialization. Module-level codecs should
// use and extend this codec as necessary. If a module-level codec does not have
// a need for rich interfaces, then they may simply use a BaseCodec as-is.
type BaseCodec struct {
	amino *Codec
}

// NewBaseCodec returns a reference to a new BaseCodec that implements the Marshaler
// interface. A BaseCodec it used to support both Protocol Buffer and Amino
// encoding. If no Amino codec is provided, it is assumed all encoding is done
// via the type implementing the ProtoMarshaler interface (i.e. Protobuf).
func NewBaseCodec(amino *Codec) Marshaler {
	return &BaseCodec{amino}
}

// MarshalBinaryBare attempts to encode the provided type. If the type implements
// ProtoMarshaler, serialization is delegated to Marshal. Otherwise, encoding falls
// back on Amino.
func (bc *BaseCodec) MarshalBinaryBare(o interface{}) ([]byte, error) {
	m, ok := o.(ProtoMarshaler)
	if ok {
		return m.Marshal()
	}

	if bc.amino == nil {
		return nil, fmt.Errorf("type '%T' does not implement ProtoMarshaler and no Amino codec provided", o)
	}

	return bc.amino.MarshalBinaryBare(o)
}

// MustMarshalBinaryBare delegates a call to MarshalBinaryBare except it panics
// on error.
func (bc *BaseCodec) MustMarshalBinaryBare(o interface{}) []byte {
	bz, err := bc.MarshalBinaryBare(o)
	if err != nil {
		panic(err)
	}

	return bz
}

// MarshalBinaryLengthPrefixed behaves the same as MarshalBinaryBare except it
// length-prefixes the resulting serialization bytes.
func (bc *BaseCodec) MarshalBinaryLengthPrefixed(o interface{}) ([]byte, error) {
	m, ok := o.(ProtoMarshaler)
	if ok {
		bz, err := bc.MarshalBinaryBare(o)
		if err != nil {
			return nil, err
		}

		buf := new(bytes.Buffer)
		if err := encodeUvarint(buf, uint64(m.Size())); err != nil {
			return nil, err
		}

		if _, err := buf.Write(bz); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}

	if bc.amino == nil {
		return nil, fmt.Errorf("type '%T' does not implement ProtoMarshaler and no Amino codec provided", o)
	}

	return bc.amino.MarshalBinaryLengthPrefixed(o)
}

// MustMarshalBinaryLengthPrefixed delegates a call to MarshalBinaryLengthPrefixed
// except it panics on error.
func (bc *BaseCodec) MustMarshalBinaryLengthPrefixed(o interface{}) []byte {
	bz, err := bc.MarshalBinaryLengthPrefixed(o)
	if err != nil {
		panic(err)
	}

	return bz
}

// UnmarshalBinaryBare attempts to deserialize raw bytes into the provided type.
// If the given type implements the ProtoMarshaler interface, it delegates
// deserialization to it. Otherwise, it falls back on Amino encoding.
func (bc *BaseCodec) UnmarshalBinaryBare(bz []byte, ptr interface{}) error {
	m, ok := ptr.(ProtoMarshaler)
	if ok {
		return m.Unmarshal(bz)
	}

	if bc.amino == nil {
		return fmt.Errorf("type '%T' does not implement ProtoMarshaler and no Amino codec provided", ptr)
	}

	return bc.amino.UnmarshalBinaryBare(bz, ptr)
}

// MustUnmarshalBinaryBare delegates a call to UnmarshalBinaryBare except it
// panics on error.
func (bc *BaseCodec) MustUnmarshalBinaryBare(bz []byte, ptr interface{}) {
	if err := bc.UnmarshalBinaryBare(bz, ptr); err != nil {
		panic(err)
	}
}

// UnmarshalBinaryLengthPrefixed attempts to deserialize raw length-prefixed
// bytes into the provided type. If the given type implements the ProtoMarshaler
// interface, it delegates deserialization to it. Otherwise, it falls back on
// Amino encoding.
func (bc *BaseCodec) UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	m, ok := ptr.(ProtoMarshaler)
	if ok {
		size, n := binary.Uvarint(bz)
		if n < 0 {
			return fmt.Errorf("invalid number of bytes read from length-prefixed encoding: %d", n)
		}

		if size > uint64(len(bz)-n) {
			return fmt.Errorf("not enough bytes to read; want: %v, got: %v", size, len(bz)-n)
		} else if size < uint64(len(bz)-n) {
			return fmt.Errorf("too many bytes to read; want: %v, got: %v", size, len(bz)-n)
		}

		bz = bz[n:]
		return m.Unmarshal(bz)
	}

	if bc.amino == nil {
		return fmt.Errorf("type '%T' does not implement ProtoMarshaler and no Amino codec provided", ptr)
	}

	return bc.amino.UnmarshalBinaryLengthPrefixed(bz, ptr)
}

// MustUnmarshalBinaryLengthPrefixed delegates a call to UnmarshalBinaryLengthPrefixed
// except it panics on error.
func (bc *BaseCodec) MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) {
	if err := bc.UnmarshalBinaryLengthPrefixed(bz, ptr); err != nil {
		panic(err)
	}
}

// MarshalJSON attempts to JSON encode an object either using Protobuf JSON if
// the type implements ProtoMarshaler or falling back on amino JSON encoding.
func (bc *BaseCodec) MarshalJSON(o interface{}) ([]byte, error) { // nolint: stdmethods
	m, ok := o.(ProtoMarshaler)
	if ok {
		buf := new(bytes.Buffer)

		marshaler := &jsonpb.Marshaler{}
		if err := marshaler.Marshal(buf, m); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}

	return bc.amino.MarshalJSON(o)
}

// MustMarshalJSON delegates a call to MarshalJSON except it panics on error.
func (bc *BaseCodec) MustMarshalJSON(o interface{}) []byte {
	bz, err := bc.MarshalJSON(o)
	if err != nil {
		panic(err)
	}

	return bz
}

// UnmarshalJSON attempts to deserialize raw JSON-encoded bytes into an object
// reference. If the object reference implements ProtoMarshaler, then it is
// decoded using Protobuf-JSON, otherwise it falls back on Amino.
func (bc *BaseCodec) UnmarshalJSON(bz []byte, ptr interface{}) error { // nolint: stdmethods
	m, ok := ptr.(ProtoMarshaler)
	if ok {
		return jsonpb.Unmarshal(strings.NewReader(string(bz)), m)
	}

	return bc.amino.UnmarshalJSON(bz, ptr)
}

// MustUnmarshalJSON delegates a call to UnmarshalJSON except it panics on error.
func (bc *BaseCodec) MustUnmarshalJSON(bz []byte, ptr interface{}) {
	if err := bc.UnmarshalJSON(bz, ptr); err != nil {
		panic(err)
	}
}

func encodeUvarint(w io.Writer, u uint64) (err error) {
	var buf [10]byte

	n := binary.PutUvarint(buf[:], u)
	_, err = w.Write(buf[0:n])

	return err
}
