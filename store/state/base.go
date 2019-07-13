package state

import (
	//	"github.com/tendermint/tendermint/crypto/merkle"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Base struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
	prefix   []byte
}

func EmptyBase() Base {
	return NewBase(nil, nil, nil)
}

func NewBase(cdc *codec.Codec, key sdk.StoreKey, rootkey []byte) Base {
	return Base{
		storeKey: key,
		cdc:      cdc,
		prefix:   rootkey,
	}
}

func (base Base) Store(ctx Context) KVStore {
	return prefix.NewStore(ctx.KVStore(base.storeKey), base.prefix)
}

func join(a, b []byte) (res []byte) {
	res = make([]byte, len(a)+len(b))
	copy(res, a)
	copy(res[len(a):], b)
	return
}

func (base Base) Prefix(prefix []byte) (res Base) {
	res = Base{
		storeKey: base.storeKey,
		cdc:      base.cdc,
		prefix:   join(base.prefix, prefix),
	}
	return
}

func (base Base) Cdc() *codec.Codec {
	return base.cdc
}

func (base Base) key(key []byte) []byte {
	return join(base.prefix, key)
}

func (base Base) StoreName() string {
	return base.storeKey.Name()
}

func (base Base) PrefixBytes() []byte {
	return base.prefix
}
