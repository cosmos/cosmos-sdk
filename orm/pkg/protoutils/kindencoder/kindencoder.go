package kindencoder

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type KindEncoder interface {
	Kind() protoreflect.Kind
	EncodeString(s string) (protoreflect.Value, error)
	EncodeInterface(i interface{}) (protoreflect.Value, error)
	EncodeValueToBytes(value protoreflect.Value) []byte
}

func NewKindEncoder(kind protoreflect.Kind) (KindEncoder, error) {
	ke, exists := kindEncoders[kind]
	if !exists {
		return nil, fmt.Errorf("kindencoder: unsupported kind %s", kind)
	}
	return ke, nil
}

var kindEncoders = map[protoreflect.Kind]KindEncoder{
	protoreflect.BoolKind:   kBool{},
	protoreflect.Int32Kind:  kInt32{},
	protoreflect.Uint32Kind: kUint32{},
	protoreflect.Int64Kind:  kInt64Kind{},
	protoreflect.Sint64Kind: kSint64Kind{},
	protoreflect.Uint64Kind: kUint64Kind{},
	protoreflect.StringKind: kString{},
	protoreflect.BytesKind:  kBytes{},
}

// kString is the protoreflect.StringKind encoder
type kString struct{}

func (k kString) Kind() protoreflect.Kind {
	return protoreflect.StringKind
}

func (k kString) EncodeString(s string) (protoreflect.Value, error) {
	return protoreflect.ValueOfString(s), nil
}

func (k kString) EncodeInterface(i interface{}) (protoreflect.Value, error) {
	v, ok := i.(string)
	if !ok {
		return protoreflect.Value{}, fmt.Errorf("invalid interface type %T expected string", i)
	}
	return protoreflect.ValueOfString(v), nil
}

func (k kString) EncodeValueToBytes(value protoreflect.Value) []byte {
	var b []byte
	// NOTE: skipping UTF8 checks, anyways marshalling would fail
	// if the string is invalid.
	// NOTE2: this prepends the string length which we can do without..
	b = protowire.AppendString(b, value.String())
	// NOTE3: this removes the length prefix which is not needed
	return b[1:]
}

// kBool is the KindEncoder for protoreflect.BoolKind
type kBool struct{}

func (k kBool) Kind() protoreflect.Kind {
	return protoreflect.BoolKind
}

func (k kBool) EncodeString(s string) (protoreflect.Value, error) {
	parsedBool, err := strconv.ParseBool(s)
	if err != nil {
		return protoreflect.Value{}, err
	}
	return protoreflect.ValueOfBool(parsedBool), nil
}

func (k kBool) EncodeInterface(i interface{}) (protoreflect.Value, error) {
	v, ok := i.(bool)
	if !ok {
		return protoreflect.Value{}, fmt.Errorf("invalid interface type %T expected bool", i)
	}
	return protoreflect.ValueOfBool(v), nil
}

func (k kBool) EncodeValueToBytes(value protoreflect.Value) []byte {
	var b []byte
	b = protowire.AppendVarint(b, protowire.EncodeBool(value.Bool()))
	return b
}

// kInt32 is the KindEncoder for protoreflect.Int32Kind
type kInt32 struct{}

func (k kInt32) Kind() protoreflect.Kind {
	return protoreflect.Int32Kind
}

func (k kInt32) EncodeString(s string) (protoreflect.Value, error) {
	parsedInt, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return protoreflect.Value{}, err
	}
	return protoreflect.ValueOfInt32(int32(parsedInt)), nil
}

func (k kInt32) EncodeInterface(i interface{}) (protoreflect.Value, error) {
	v, ok := i.(int32)
	if !ok {
		return protoreflect.Value{}, fmt.Errorf("invalid interface type %T expected int32", i)
	}
	return protoreflect.ValueOfInt32(v), nil
}

func (k kInt32) EncodeValueToBytes(value protoreflect.Value) []byte {
	var b []byte
	b = protowire.AppendVarint(b, uint64(int32(value.Int())))
	return b
}

type kUint32 struct{}

func (k kUint32) Kind() protoreflect.Kind {
	return protoreflect.Uint32Kind
}

func (k kUint32) EncodeString(s string) (protoreflect.Value, error) {
	parsedUint32, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return protoreflect.Value{}, err
	}
	return protoreflect.ValueOfUint32(uint32(parsedUint32)), nil
}

func (k kUint32) EncodeInterface(i interface{}) (protoreflect.Value, error) {
	v, ok := i.(uint32)
	if !ok {
		return protoreflect.Value{}, fmt.Errorf("invalid interface type %T expected uint32", i)
	}
	return protoreflect.ValueOfUint32(v), nil
}

func (k kUint32) EncodeValueToBytes(value protoreflect.Value) []byte {
	var b []byte
	b = protowire.AppendVarint(b, uint64(uint32(value.Uint())))
	return b
}

type kInt64Kind struct{}

func (k kInt64Kind) Kind() protoreflect.Kind {
	return protoreflect.Int64Kind
}

func (k kInt64Kind) EncodeString(s string) (protoreflect.Value, error) {
	parsedInt64, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return protoreflect.Value{}, err
	}
	return protoreflect.ValueOfInt64(parsedInt64), nil
}

func (k kInt64Kind) EncodeInterface(i interface{}) (protoreflect.Value, error) {
	v, ok := i.(int64)
	if !ok {
		return protoreflect.Value{}, fmt.Errorf("invalid interface type %T expected int64", i)
	}
	return protoreflect.ValueOfInt64(v), nil
}

func (k kInt64Kind) EncodeValueToBytes(value protoreflect.Value) []byte {
	var b []byte
	b = protowire.AppendVarint(b, uint64(value.Int()))
	return b
}

type kSint64Kind struct {
}

func (k kSint64Kind) Kind() protoreflect.Kind {
	panic("implement me")
}

func (k kSint64Kind) EncodeString(s string) (protoreflect.Value, error) {
	panic("implement me")
}

func (k kSint64Kind) EncodeInterface(i interface{}) (protoreflect.Value, error) {
	panic("implement me")
}

func (k kSint64Kind) EncodeValueToBytes(value protoreflect.Value) []byte {
	panic("implement me")
}

type kUint64Kind struct {
}

func (k kUint64Kind) Kind() protoreflect.Kind {
	return protoreflect.Uint64Kind
}

func (k kUint64Kind) EncodeString(s string) (protoreflect.Value, error) {
	parsedUint64, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return protoreflect.Value{}, err
	}
	return protoreflect.ValueOfUint64(parsedUint64), nil
}

func (k kUint64Kind) EncodeInterface(i interface{}) (protoreflect.Value, error) {
	v, ok := i.(uint64)
	if !ok {
		return protoreflect.Value{}, fmt.Errorf("invalid interface type %T expected uint64", i)
	}
	return protoreflect.ValueOfUint64(v), nil
}

func (k kUint64Kind) EncodeValueToBytes(value protoreflect.Value) []byte {
	var b []byte
	b = protowire.AppendVarint(b, value.Uint())
	return b
}

type kBytes struct {
}

func (k kBytes) Kind() protoreflect.Kind {
	return protoreflect.BytesKind
}

func (k kBytes) EncodeString(s string) (protoreflect.Value, error) {
	b, err := base64.RawStdEncoding.DecodeString(s)
	if err != nil {
		return protoreflect.Value{}, err
	}
	return protoreflect.ValueOfBytes(b), err
}

func (k kBytes) EncodeInterface(i interface{}) (protoreflect.Value, error) {
	v, ok := i.([]byte)
	if !ok {
		return protoreflect.Value{}, fmt.Errorf("invalid interface type %T expected []byte", i)
	}
	return protoreflect.ValueOfBytes(v), nil
}

func (k kBytes) EncodeValueToBytes(value protoreflect.Value) []byte {
	var b []byte
	b = protowire.AppendBytes(b, value.Bytes())
	// NOTE: removes the length prefix which is not needed
	return b[1:]
}
