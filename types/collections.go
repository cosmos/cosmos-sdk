package types

import (
	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	"cosmossdk.io/math"
)

var (
	// AccAddressKey follows the same semantics of collections.BytesKey.
	// It just uses humanized format for the String() and EncodeJSON().
	AccAddressKey collcodec.KeyCodec[AccAddress] = genericAddressKey[AccAddress]{
		stringDecoder: AccAddressFromBech32,
		keyType:       "sdk.AccAddress",
	}

	// ValAddressKey follows the same semantics as AccAddressKey.
	ValAddressKey collcodec.KeyCodec[ValAddress] = genericAddressKey[ValAddress]{
		stringDecoder: ValAddressFromBech32,
		keyType:       "sdk.ValAddress",
	}

	// ConsAddressKey follows the same semantics as ConsAddressKey.
	ConsAddressKey collcodec.KeyCodec[ConsAddress] = genericAddressKey[ConsAddress]{
		stringDecoder: ConsAddressFromBech32,
		keyType:       "sdk.ConsAddress",
	}

	// IntValue represents a collections.ValueCodec to work with Int.
	IntValue collcodec.ValueCodec[math.Int] = intValueCodec{}
)

type addressUnion interface {
	AccAddress | ValAddress | ConsAddress
	String() string
}

type genericAddressKey[T addressUnion] struct {
	stringDecoder func(string) (T, error)
	keyType       string
}

func (a genericAddressKey[T]) Encode(buffer []byte, key T) (int, error) {
	return collections.BytesKey.Encode(buffer, key)
}

func (a genericAddressKey[T]) Decode(buffer []byte) (int, T, error) {
	return collections.BytesKey.Decode(buffer)
}

func (a genericAddressKey[T]) Size(key T) int {
	return collections.BytesKey.Size(key)
}

func (a genericAddressKey[T]) EncodeJSON(value T) ([]byte, error) {
	return collections.StringKey.EncodeJSON(value.String())
}

func (a genericAddressKey[T]) DecodeJSON(b []byte) (v T, err error) {
	s, err := collections.StringKey.DecodeJSON(b)
	if err != nil {
		return
	}
	v, err = a.stringDecoder(s)
	return
}

func (a genericAddressKey[T]) Stringify(key T) string {
	return key.String()
}

func (a genericAddressKey[T]) KeyType() string {
	return a.keyType
}

func (a genericAddressKey[T]) EncodeNonTerminal(buffer []byte, key T) (int, error) {
	return collections.BytesKey.EncodeNonTerminal(buffer, key)
}

func (a genericAddressKey[T]) DecodeNonTerminal(buffer []byte) (int, T, error) {
	return collections.BytesKey.DecodeNonTerminal(buffer)
}

func (a genericAddressKey[T]) SizeNonTerminal(key T) int {
	return collections.BytesKey.SizeNonTerminal(key)
}

// Deprecated: genericAddressIndexKey is a special key codec used to retain state backwards compatibility
// when a generic address key (be: AccAddress, ValAddress, ConsAddress), is used as an index key.
// More docs can be found in the AddressKeyAsIndexKey function.
type genericAddressIndexKey[T addressUnion] struct {
	collcodec.KeyCodec[T]
}

func (g genericAddressIndexKey[T]) Encode(buffer []byte, key T) (int, error) {
	return g.EncodeNonTerminal(buffer, key)
}

func (g genericAddressIndexKey[T]) Decode(buffer []byte) (int, T, error) {
	return g.DecodeNonTerminal(buffer)
}

func (g genericAddressIndexKey[T]) Size(key T) int { return g.SizeNonTerminal(key) }

func (g genericAddressIndexKey[T]) KeyType() string { return "index_key/" + g.KeyCodec.KeyType() }

// Deprecated: AddressKeyAsIndexKey implements an SDK backwards compatible indexing key encoder
// for addresses.
// The status quo in the SDK is that address keys are length prefixed even when they're the
// last part of a composite key. This should never be used unless to retain state compatibility.
// For example, a composite key composed of `[string, address]` in theory would need you only to
// define a way to understand when the string part finishes, we usually do this by appending a null
// byte to the string, then when you know when the string part finishes, it's logical that the
// part which remains is the address key. In the SDK instead we prepend to the address key its
// length too.
func AddressKeyAsIndexKey[T addressUnion](keyCodec collcodec.KeyCodec[T]) collcodec.KeyCodec[T] {
	return genericAddressIndexKey[T]{
		keyCodec,
	}
}

// Collection Codecs

type intValueCodec struct{}

func (i intValueCodec) Encode(value math.Int) ([]byte, error) {
	return value.Marshal()
}

func (i intValueCodec) Decode(b []byte) (math.Int, error) {
	v := new(Int)
	err := v.Unmarshal(b)
	if err != nil {
		return Int{}, err
	}
	return *v, nil
}

func (i intValueCodec) EncodeJSON(value math.Int) ([]byte, error) {
	return value.MarshalJSON()
}

func (i intValueCodec) DecodeJSON(b []byte) (Int, error) {
	v := new(Int)
	err := v.UnmarshalJSON(b)
	if err != nil {
		return Int{}, err
	}
	return *v, nil
}

func (i intValueCodec) Stringify(value Int) string {
	return value.String()
}

func (i intValueCodec) ValueType() string {
	return "math.Int"
}
