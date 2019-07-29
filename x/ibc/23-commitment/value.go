package commitment

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Mapping struct {
	cdc    *codec.Codec
	prefix []byte
}

func NewMapping(cdc *codec.Codec, prefix []byte) Mapping {
	return Mapping{
		cdc:    cdc,
		prefix: prefix,
	}
}

func (m Mapping) store(ctx sdk.Context) Store {
	return NewPrefix(GetStore(ctx), m.prefix)
}

func (m Mapping) Prefix(prefix []byte) Mapping {
	return Mapping{
		cdc:    m.cdc,
		prefix: join(m.prefix, prefix),
	}
}

// Value is for proving commitment proof on a speicifc key-value point in the other state
// using the already initialized commitment store.
type Value struct {
	m   Mapping
	key []byte
}

func (m Mapping) Value(key []byte) Value {
	return Value{m, key}
}

// Is() proves the proof with the Value's key and the provided value.
func (v Value) Is(ctx sdk.Context, value interface{}) bool {
	return v.m.store(ctx).Prove(v.key, v.m.cdc.MustMarshalBinaryBare(value))
}

// IsRaw() proves the proof with the Value's key and the provided raw value bytes.
func (v Value) IsRaw(ctx sdk.Context, value []byte) bool {
	return v.m.store(ctx).Prove(v.key, value)
}

// Enum is a byte typed wrapper for Value.
// Except for the type checking, it does not alter the behaviour.
type Enum struct {
	Value
}

// Enum() wraps the argument Value as Enum
func (v Value) Enum() Enum {
	return Enum{v}
}

// Is() proves the proof with the Enum's key and the provided value
func (v Enum) Is(ctx sdk.Context, value byte) bool {
	return v.Value.IsRaw(ctx, []byte{value})
}

// Integer is a uint64 types wrapper for Value.
type Integer struct {
	Value

	enc state.IntEncoding
}

func (v Value) Integer(enc state.IntEncoding) Integer {
	return Integer{v, enc}
}

func (v Integer) Is(ctx sdk.Context, value uint64) bool {
	return v.Value.IsRaw(ctx, state.EncodeInt(value, v.enc))
}
