package commitment

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/store/mapping"
)

var _ mapping.Value = value{}

type value struct {
	mapping.Value
}

func Value(base mapping.Base, key []byte) mapping.Value {
	return value{mapping.NewValue(base, key)}
}

func (v value) Is(ctx sdk.Context, value interface{}) bool {
	// CONTRACT: commitment.value must be used with commitment.Store as its underlying KVStore
	// TODO: enforce it

	v.Set(ctx, value)
	return v.Exists(ctx)
}

var _ mapping.Enum = enum{}

type enum struct {
	mapping.Enum
}

func Enum(v mapping.Value) mapping.Enum {
	return enum{mapping.NewEnum(v)}
}

func (v enum) Is(ctx sdk.Context, value byte) bool {
	// CONTRACT: commitment.enum must be used with commitment.Store as its underlying KVStore
	// TODO: enforce it

	v.Set(ctx, value)
	return v.Exists(ctx)
}

var _ mapping.Integer = integer{}

type integer struct {
	mapping.Integer
}

func Integer(v mapping.Value, enc mapping.IntEncoding) mapping.Integer {
	return integer{mapping.NewInteger(v, enc)}
}

func (v integer) Is(ctx sdk.Context, value uint64) bool {
	// CONTRACT: commitment.integer must be used with commitment.Store as its underlying KVStore
	// TODO: enforce it

	v.Set(ctx, value)
	return v.Exists(ctx)
}
