package types

import (
	"context"
	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	"cosmossdk.io/math"
)

var (
	// AccAddressKey follows the same semantics of collections.BytesKey.
	// It just uses humanised format for the String() and EncodeJSON().
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

type coll[K, V any] interface {
	Iterate(ctx context.Context, ranger collections.Ranger[K]) (collections.Iterator[K, V], error)
}

// IterateAndCallBack is a helper function that applies one of the most common patterns in the sdk.
// It iterates over the entire collection and calls the provided function with the key and the value
// of the iterator. If the iteration needs to be terminated then the provided function should return
// true. The iterator is closed when this function exits.
func IterateAndCallBack[K, V any, C coll[K, V]](ctx context.Context, coll C, cb func(key K, value V) bool) error {
	iter, err := coll.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}
		if cb(kv.Key, kv.Value) {
			return nil
		}
	}
	return nil
}

// IterateAndCallBackWithRange applies the same logic of IterateAndCallBack, but it
// allows consumers to provide an extra argument, a Ranger, whose purpose is to
// set bounds to the iterator range.
func IterateAndCallBackWithRange[K, V any, C coll[K, V], R collections.Ranger[K]](
	ctx context.Context, coll C, ranger R, cb func(key K, value V) bool,
) error {
	iter, err := coll.Iterate(ctx, ranger)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}
		if cb(kv.Key, kv.Value) {
			return nil
		}
	}
	return nil
}

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
