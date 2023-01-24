package types

import "cosmossdk.io/collections"

var (
	// AccAddressKey follows the same semantics of collections.BytesKey.
	// It just uses humanised format for the String() and EncodeJSON().
	AccAddressKey collections.KeyCodec[AccAddress] = genericAddressKey[AccAddress]{
		stringDecoder: AccAddressFromBech32,
		keyType:       "sdk.AccAddress",
	}

	// ValAddressKey follows the same semantics as AccAddressKey.
	ValAddressKey collections.KeyCodec[ValAddress] = genericAddressKey[ValAddress]{
		stringDecoder: ValAddressFromBech32,
		keyType:       "sdk.ValAddress",
	}

	// ConsAddressKey follows the same semantics as ConsAddressKey.
	ConsAddressKey collections.KeyCodec[ConsAddress] = genericAddressKey[ConsAddress]{
		stringDecoder: ConsAddressFromBech32,
		keyType:       "sdk.ConsAddress",
	}
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
