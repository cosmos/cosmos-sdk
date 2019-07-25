package state

import (
	//	"github.com/tendermint/tendermint/crypto/merkle"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Base is a state accessor base layer, consists of Codec, StoreKey, and prefix.
// StoreKey is used to get the KVStore, cdc is used to marshal/unmarshal the interfaces,
// and the prefix is prefixed to the key.
//
// Base has practically the same capability with the storeKey.
// It should not be passed to an untrusted actor.
type Base struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
	prefix   []byte
}

// NewBase() is the constructor for Base()
func NewBase(cdc *codec.Codec, key sdk.StoreKey, rootkey []byte) Base {
	return Base{
		storeKey: key,
		cdc:      cdc,
		prefix:   rootkey,
	}
}

func (base Base) store(ctx Context) KVStore {
	return prefix.NewStore(ctx.KVStore(base.storeKey), base.prefix)
}

func join(a, b []byte) (res []byte) {
	res = make([]byte, len(a)+len(b))
	copy(res, a)
	copy(res[len(a):], b)
	return
}

// Prefix() returns a copy of the Base with the updated prefix.
func (base Base) Prefix(prefix []byte) (res Base) {
	res = Base{
		storeKey: base.storeKey,
		cdc:      base.cdc,
		prefix:   join(base.prefix, prefix),
	}
	return
}

// Cdc() returns the codec of the base. It is safe to expose the codec.
func (base Base) Cdc() *codec.Codec {
	return base.cdc
}

func (base Base) key(key []byte) []byte {
	return join(base.prefix, key)
}

// StoreName() returns the name of the storeKey. It is safe to expose the store name.
// Used by the CLI side query operations.
func (base Base) StoreName() string {
	return base.storeKey.Name()
}

// PrefixBytes() returns the prefix bytes. It is safe to expsoe the prefix bytes.
// Used by the CLI side query operations.
func (base Base) PrefixBytes() (res []byte) {
	res = make([]byte, len(base.prefix))
	copy(res, base.prefix)
	return
}
