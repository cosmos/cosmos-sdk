package auth

import (
	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Implements sdk.AccountStore
type accountStore struct {
	key   sdk.SubstoreKey
	codec sdk.Codec
}

func NewAccountStore(key sdk.SubstoreKey, codec sdk.Codec) accountStore {
	return accountStore{
		key:   key,
		codec: codec,
	}
}

// Implements sdk.AccountStore
func (as accountStore) NewAccountWithAddress(ctx sdk.Context, addr crypto.Address) sdk.Account {
	acc := as.codec.Prototype().(sdk.Account)
	acc.SetAddress(addr)
	return acc
}

// Implements sdk.AccountStore
func (as accountStore) GetAccount(ctx sdk.Context, addr crypto.Address) sdk.Account {
	store := ctx.KVStore(as.key)
	bz := store.Get(addr)
	if bz == nil {
		return nil // XXX
	}
	o, err := as.codec.Decode(bz)
	if err != nil {
		panic(err)
	}
	return o.(sdk.Account)
}

// Implements sdk.AccountStore
func (as accountStore) SetAccount(ctx sdk.Context, acc sdk.Account) {
	addr := acc.GetAddress()
	store := ctx.KVStore(as.key)
	bz, err := as.codec.Encode(acc)
	if err != nil {
		panic(err)
	}
	store.Set(addr, bz)
}
