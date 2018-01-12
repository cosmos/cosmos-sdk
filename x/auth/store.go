package auth

import (
	"github.com/cosmos/cosmos-sdk/types"
)

// Implements types.AccountStore
type accountStore struct {
	key   *types.KVStoreKey
	codec types.Codec
}

func NewAccountStore(key *types.KVStoreKey, codec types.Codec) accountStore {
	return accountStore{
		key:   key,
		codec: codec,
	}
}

// Implements types.AccountStore
func (as accountStore) NewAccountWithAddress(ctx types.Context, addr crypto.Address) {
	acc := as.codec.Prototype().(types.Account)
	acc.SetAddress(addr)
	return acc
}

// Implements types.AccountStore
func (as accountStore) GetAccount(ctx types.Context, addr crypto.Address) types.Account {
	store := ctx.KVStore(as.key)
	bz := store.Get(addr)
	if bz == nil {
		return
	}
	o, err := as.codec.Decode(bz)
	if err != nil {
		panic(err)
	}
	return o.(types.Account)
}

// Implements types.AccountStore
func (as accountStore) SetAccount(ctx types.Context, acc types.Account) {
	addr := acc.GetAddress()
	store := ctx.KVStore(as.key)
	bz, err := as.codec.Encode(acc)
	if err != nil {
		panic(err)
	}
	store.Set(addr, bz)
}
