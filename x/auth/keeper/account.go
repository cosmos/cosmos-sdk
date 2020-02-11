package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// NewAccountWithAddress implements sdk.AccountKeeper.
func (ak AccountKeeper) NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) exported.AccountI {
	acc := ak.proto()
	err := acc.SetAddress(addr)
	if err != nil {
		panic(err)
	}

	return ak.NewAccount(ctx, acc)
}

// NewAccount sets the next account number to a given account interface
func (ak AccountKeeper) NewAccount(ctx sdk.Context, acc exported.AccountI) exported.AccountI {
	if err := acc.SetAccountNumber(ak.GetNextAccountNumber(ctx)); err != nil {
		panic(err)
	}

	return acc
}

// GetAccount implements sdk.AccountKeeper.
func (ak AccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) exported.AccountI {
	store := ctx.KVStore(ak.key)
	bz := store.Get(types.AddressStoreKey(addr))
	if bz == nil {
		return nil
	}

	return ak.decodeAccount(bz)
}

// GetAllAccounts returns all accounts in the accountKeeper.
func (ak AccountKeeper) GetAllAccounts(ctx sdk.Context) (accounts []exported.AccountI) {
	ak.IterateAccounts(ctx, func(acc exported.AccountI) (stop bool) {
		accounts = append(accounts, acc)
		return false
	})

	return accounts
}

// SetAccount implements sdk.AccountKeeper.
func (ak AccountKeeper) SetAccount(ctx sdk.Context, acc exported.AccountI) {
	addr := acc.GetAddress()
	store := ctx.KVStore(ak.key)

	bz, err := ak.cdc.MarshalAccount(acc)
	if err != nil {
		panic(err)
	}

	store.Set(types.AddressStoreKey(addr), bz)
}

// RemoveAccount removes an account for the account mapper store.
// NOTE: this will cause supply invariant violation if called
func (ak AccountKeeper) RemoveAccount(ctx sdk.Context, acc exported.AccountI) {
	addr := acc.GetAddress()
	store := ctx.KVStore(ak.key)
	store.Delete(types.AddressStoreKey(addr))
}

// IterateAccounts iterates over all the stored accounts and performs a callback function
func (ak AccountKeeper) IterateAccounts(ctx sdk.Context, cb func(account exported.AccountI) (stop bool)) {
	store := ctx.KVStore(ak.key)
	iterator := sdk.KVStorePrefixIterator(store, types.AddressStoreKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		account := ak.decodeAccount(iterator.Value())

		if cb(account) {
			break
		}
	}
}
