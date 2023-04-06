package keeper

import (
	"context"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// NewAccountWithAddress implements AccountKeeperI.
func (ak AccountKeeper) NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	acc := ak.proto()
	err := acc.SetAddress(addr)
	if err != nil {
		panic(err)
	}

	return ak.NewAccount(ctx, acc)
}

// NewAccount sets the next account number to a given account interface
func (ak AccountKeeper) NewAccount(ctx context.Context, acc sdk.AccountI) sdk.AccountI {
	if err := acc.SetAccountNumber(ak.NextAccountNumber(ctx)); err != nil {
		panic(err)
	}

	return acc
}

// HasAccount implements AccountKeeperI.
func (ak AccountKeeper) HasAccount(ctx context.Context, addr sdk.AccAddress) bool {
	store := ak.storeService.OpenKVStore(ctx)
	has, err := store.Has(types.AddressStoreKey(addr))
	if err != nil {
		panic(err)
	}
	return has
}

// HasAccountAddressByID checks account address exists by id.
func (ak AccountKeeper) HasAccountAddressByID(ctx context.Context, id uint64) bool {
	store := ak.storeService.OpenKVStore(ctx)
	has, err := store.Has(types.AccountNumberStoreKey(id))
	if err != nil {
		panic(err)
	}
	return has
}

// GetAccount implements AccountKeeperI.
func (ak AccountKeeper) GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	store := ak.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.AddressStoreKey(addr))
	if err != nil {
		panic(err)
	}

	if bz == nil {
		return nil
	}

	return ak.decodeAccount(bz)
}

// GetAccountAddressById returns account address by id.
func (ak AccountKeeper) GetAccountAddressByID(ctx context.Context, id uint64) string {
	store := ak.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.AccountNumberStoreKey(id))
	if err != nil {
		panic(err)
	}

	if bz == nil {
		return ""
	}
	return sdk.AccAddress(bz).String()
}

// GetAllAccounts returns all accounts in the accountKeeper.
func (ak AccountKeeper) GetAllAccounts(ctx context.Context) (accounts []sdk.AccountI) {
	ak.IterateAccounts(ctx, func(acc sdk.AccountI) (stop bool) {
		accounts = append(accounts, acc)
		return false
	})

	return accounts
}

// SetAccount implements AccountKeeperI.
func (ak AccountKeeper) SetAccount(ctx context.Context, acc sdk.AccountI) {
	addr := acc.GetAddress()
	store := ak.storeService.OpenKVStore(ctx)

	bz, err := ak.MarshalAccount(acc)
	if err != nil {
		panic(err)
	}

	store.Set(types.AddressStoreKey(addr), bz)
	store.Set(types.AccountNumberStoreKey(acc.GetAccountNumber()), addr.Bytes())
}

// RemoveAccount removes an account for the account mapper store.
// NOTE: this will cause supply invariant violation if called
func (ak AccountKeeper) RemoveAccount(ctx context.Context, acc sdk.AccountI) {
	addr := acc.GetAddress()
	store := ak.storeService.OpenKVStore(ctx)
	err := store.Delete(types.AddressStoreKey(addr))
	if err != nil {
		panic(err)
	}

	err = store.Delete(types.AccountNumberStoreKey(acc.GetAccountNumber()))
	if err != nil {
		panic(err)
	}
}

// IterateAccounts iterates over all the stored accounts and performs a callback function.
// Stops iteration when callback returns true.
func (ak AccountKeeper) IterateAccounts(ctx context.Context, cb func(account sdk.AccountI) (stop bool)) {
	store := ak.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.AddressStoreKeyPrefix, storetypes.PrefixEndBytes(types.AddressStoreKeyPrefix))
	if err != nil {
		panic(err)
	}

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		account := ak.decodeAccount(iterator.Value())

		if cb(account) {
			break
		}
	}
}
