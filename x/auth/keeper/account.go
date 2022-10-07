package keeper

import (
	store2 "github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// NewAccountWithAddress implements AccountKeeperI.
func (ak AccountKeeper) NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) types.AccountI {
	acc := ak.proto()
	err := acc.SetAddress(addr)
	if err != nil {
		panic(err)
	}

	return ak.NewAccount(ctx, acc)
}

// NewAccount sets the next account number to a given account interface
func (ak AccountKeeper) NewAccount(ctx sdk.Context, acc types.AccountI) types.AccountI {
	if err := acc.SetAccountNumber(ak.GetNextAccountNumber(ctx)); err != nil {
		panic(err)
	}

	return acc
}

// HasAccount implements AccountKeeperI.
func (ak AccountKeeper) HasAccount(ctx sdk.Context, addr sdk.AccAddress) bool {
	store := ctx.KVStore(ak.storeKey)
	return store.Has(types.AddressStoreKey(addr))
}

// HasAccountAddressByID checks account address exists by id.
func (ak AccountKeeper) HasAccountAddressByID(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(ak.storeKey)
	return store.Has(types.AccountNumberStoreKey(id))
}

// GetAccount implements AccountKeeperI.
func (ak AccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI {
	store := ctx.KVStore(ak.storeKey)
	acc, err := store2.GetAndDecode(store, ak.decodeAccount, types.AddressStoreKey(addr))
	if err != nil {
		panic(err)
	}
	return acc
}

func decodeAccAddr(bz []byte) (string, error) {
	if bz == nil {
		return "", nil
	}
	res := sdk.AccAddress(bz)
	return res.String(), nil
}

// GetAccountAddressById returns account address by id.
func (ak AccountKeeper) GetAccountAddressByID(ctx sdk.Context, id uint64) string {
	store := ctx.KVStore(ak.storeKey)
	addr, err := store2.GetAndDecode(store, decodeAccAddr, types.AccountNumberStoreKey(id))
	if err != nil {
		panic(err)
	}
	return addr
}

// GetAllAccounts returns all accounts in the accountKeeper.
func (ak AccountKeeper) GetAllAccounts(ctx sdk.Context) (accounts []types.AccountI) {
	ak.IterateAccounts(ctx, func(acc types.AccountI) (stop bool) {
		accounts = append(accounts, acc)
		return false
	})

	return accounts
}

func (ak AccountKeeper) getStore(ctx sdk.Context) store2.StoreAPI {
	return store2.NewStoreAPI(ctx.KVStore(ak.storeKey))
}

// SetAccount implements AccountKeeperI.
func (ak AccountKeeper) SetAccount(ctx sdk.Context, acc types.AccountI) {
	addr := acc.GetAddress()
	store := ak.getStore(ctx)

	bz, err := ak.MarshalAccount(acc)
	if err != nil {
		panic(err)
	}

	store.Set(types.AddressStoreKey(addr), bz)
	store.Set(types.AccountNumberStoreKey(acc.GetAccountNumber()), addr.Bytes())
}

// RemoveAccount removes an account for the account mapper store.
// NOTE: this will cause supply invariant violation if called
func (ak AccountKeeper) RemoveAccount(ctx sdk.Context, acc types.AccountI) {
	addr := acc.GetAddress()
	store := ak.getStore(ctx)
	store.Delete(types.AddressStoreKey(addr))
	store.Delete(types.AccountNumberStoreKey(acc.GetAccountNumber()))
}

// IterateAccounts iterates over all the stored accounts and performs a callback function.
// Stops iteration when callback returns true.
func (ak AccountKeeper) IterateAccounts(ctx sdk.Context, cb func(account types.AccountI) (stop bool)) {
	store := ctx.KVStore(ak.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.AddressStoreKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		account, err := ak.decodeAccount(iterator.Value())
		if err != nil {
			panic(err)
		}
		if cb(account) {
			break
		}
	}
}
