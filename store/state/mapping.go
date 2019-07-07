package state

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Mapping is key []byte -> value []byte mapping using a base(possibly prefixed).
// All store accessing operations are redirected to the Value corresponding to the key argument
type Mapping struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
	prefix   []byte
}

// NewMapping() constructs a Mapping with a provided prefix
func NewMapping(storeKey sdk.StoreKey, cdc *codec.Codec, prefix []byte) Mapping {
	return Mapping{
		storeKey: storeKey,
		cdc:      cdc,
		prefix:   prefix,
	}
}

// Value() returns the Value corresponding to the provided key
func (m Mapping) Value(key []byte) Value {
	return NewValue(m, key)
}

// Get() unmarshales and sets the stored value to the pointer if it exists.
// It will panic if the value exists but not unmarshalable.
func (m Mapping) Get(ctx Context, key []byte, ptr interface{}) {
	m.Value(key).Get(ctx, ptr)
}

// GetSafe() unmarshales and sets the stored value to the pointer.
// It will return an error if the value does not exist or unmarshalable.
func (m Mapping) GetSafe(ctx Context, key []byte, ptr interface{}) error {
	return m.Value(key).GetSafe(ctx, ptr)
}

// Set() marshales and sets the argument to the state.
// Calls Delete() if the argument is nil.
func (m Mapping) Set(ctx Context, key []byte, o interface{}) {
	if o == nil {
		m.Delete(ctx, key)
		return
	}
	m.Value(key).Set(ctx, o)
}

func (m Mapping) SetRaw(ctx Context, key []byte, value []byte) {
	m.Value(key).SetRaw(ctx, value)
}

// Has() returns true if the stored value is not nil
func (m Mapping) Has(ctx Context, key []byte) bool {
	return m.Value(key).Exists(ctx)
}

// Delete() deletes the stored value.
func (m Mapping) Delete(ctx Context, key []byte) {
	m.Value(key).Delete(ctx)
}

func (m Mapping) Cdc() *codec.Codec {
	return m.cdc
}

func (m Mapping) StoreName() string {
	return m.storeKey.Name()
}

func (m Mapping) PrefixBytes() (res []byte) {
	res = make([]byte, len(m.prefix))
	copy(res, m.prefix)
	return
}

func (m Mapping) KeyBytes(key []byte) (res []byte) {
	return join(m.prefix, key)
}

func join(a, b []byte) (res []byte) {
	res = make([]byte, len(a)+len(b))
	copy(res, a)
	copy(res[len(a):], b)
	return
}

// Prefix() returns a new mapping with the updated prefix.
func (m Mapping) Prefix(prefix []byte) Mapping {
	return Mapping{
		storeKey: m.storeKey,
		cdc:      m.cdc,
		prefix:   join(m.prefix, prefix),
	}
}

