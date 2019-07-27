package commitment

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Base is a proof store accessor, consists of Codec and prefix.
// The base uses the commitment store which is expected to be filled with the proofs.
type Base struct {
	cdc    *codec.Codec
	prefix []byte
}

func NewBase(cdc *codec.Codec) Base {
	return Base{
		cdc: cdc,
	}
}

func (base Base) Store(ctx sdk.Context) Store {
	return NewPrefix(GetStore(ctx), base.prefix)
}

func (base Base) Prefix(prefix []byte) Base {
	return Base{
		cdc:    base.cdc,
		prefix: join(base.prefix, prefix),
	}
}

type Mapping struct {
	base Base
}

func NewMapping(base Base, prefix []byte) Mapping {
	return Mapping{
		base: base.Prefix(prefix),
	}
}

// Value is for proving commitment proof on a speicifc key-value point in the other state
// using the already initialized commitment store.
type Value struct {
	base Base
	key  []byte
}

func NewValue(base Base, key []byte) Value {
	return Value{base, key}
}

// Is() proves the proof with the Value's key and the provided value.
func (v Value) Is(ctx sdk.Context, value interface{}) bool {
	return v.base.Store(ctx).Prove(v.key, v.base.cdc.MustMarshalBinaryBare(value))
}

// IsRaw() proves the proof with the Value's key and the provided raw value bytes.
func (v Value) IsRaw(ctx sdk.Context, value []byte) bool {
	return v.base.Store(ctx).Prove(v.key, value)
}

// Enum is a byte typed wrapper for Value.
// Except for the type checking, it does not alter the behaviour.
type Enum struct {
	Value
}

// NewEnum() wraps the argument Value as Enum
func NewEnum(v Value) Enum {
	return Enum{v}
}

// Is() proves the proof with the Enum's key and the provided value
func (v Enum) Is(ctx sdk.Context, value byte) bool {
	return v.Value.IsRaw(ctx, []byte{value})
}

// Integer is a uint64 types wrapper for Value.
type Integer struct {
	Value
}

func NewInteger(v Value) Integer {
	return Integer{v}
}

func (v Integer) Is(ctx sdk.Context, value uint64) bool {
	return v.Value.IsRaw(ctx, state.EncodeInt(value, v.enc))
}
