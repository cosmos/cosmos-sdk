package commitment

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

func join(a, b []byte) (res []byte) {
	res = make([]byte, len(a)+len(b))
	copy(res, a)
	copy(res[len(a):], b)
	return
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

type Value struct {
	base Base
	key  []byte
}

func NewValue(base Base, key []byte) Value {
	return Value{base, key}
}

func (v Value) Is(ctx sdk.Context, value interface{}) bool {
	return v.base.Store(ctx).Prove(v.key, v.base.cdc.MustMarshalBinaryBare(value))
}

func (v Value) IsRaw(ctx sdk.Context, value []byte) bool {
	return v.base.Store(ctx).Prove(v.key, value)
}

type Enum struct {
	Value
}

func NewEnum(v Value) Enum {
	return Enum{v}
}

func (v Enum) Is(ctx sdk.Context, value byte) bool {
	return v.Value.IsRaw(ctx, []byte{value})
}

type Integer struct {
	Value

	enc state.IntEncoding
}

func NewInteger(v Value, enc state.IntEncoding) Integer {
	return Integer{v, enc}
}

func (v Integer) Is(ctx sdk.Context, value uint64) bool {
	return v.Value.IsRaw(ctx, state.EncodeInt(value, v.enc))
}
