package commitment

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Mapping is key []byte -> value []byte mapping, possibly prefixed.
// Proof verification should be done over Value constructed from the Mapping.
type Mapping struct {
	cdc    *codec.Codec
	prefix []byte
}

// NewMapping() constructs a new Mapping.
// The KVStore accessor is fixed to the commitment store.
func NewMapping(cdc *codec.Codec, prefix []byte) Mapping {
	return Mapping{
		cdc:    cdc,
		prefix: prefix,
	}
}

func (m Mapping) store(ctx sdk.Context) Store {
	return NewPrefix(GetStore(ctx), m.prefix)
}

// Prefix() returns a new Mapping with the updated prefix
func (m Mapping) Prefix(prefix []byte) Mapping {
	return Mapping{
		cdc:    m.cdc,
		prefix: join(m.prefix, prefix),
	}
}

type Indexer struct {
	Mapping
	enc state.IntEncoding
}

func (m Mapping) Indexer(enc state.IntEncoding) Indexer {
	return Indexer{
		Mapping: m,
		enc:     enc,
	}
}

func (ix Indexer) Value(index uint64) Value {
	return ix.Mapping.Value(state.EncodeInt(index, ix.enc))
}

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

type String struct {
	Value
}

func (v Value) String() String {
	return String{v}
}

func (v String) Is(ctx sdk.Context, value string) bool {
	return v.Value.IsRaw(ctx, []byte(value))
}

type Boolean struct {
	Value
}

func (v Value) Boolean() Boolean {
	return Boolean{v}
}

func (v Boolean) Is(ctx sdk.Context, value bool) bool {
	return v.Value.Is(ctx, value)
}

// Integer is a uint64 types wrapper for Value.
type Integer struct {
	Value

	enc state.IntEncoding
}

// Integer() wraps the argument Value as Integer
func (v Value) Integer(enc state.IntEncoding) Integer {
	return Integer{v, enc}
}

// Is() proves the proof with the Integer's key and the provided value
func (v Integer) Is(ctx sdk.Context, value uint64) bool {
	return v.Value.IsRaw(ctx, state.EncodeInt(value, v.enc))
}