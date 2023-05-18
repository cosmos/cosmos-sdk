package keeper

import (
	"context"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// NewAccountWithAddress implements AccountKeeperI.
func (ak AccountKeeper) NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) (sdk.AccountI,error) {
	acc := ak.proto()
	err := acc.SetAddress(addr)
	if err != nil {
		return nil,err
	}

	return ak.NewAccount(ctx, acc)
}

// NewAccount sets the next account number to a given account interface
func (ak AccountKeeper) NewAccount(ctx context.Context, acc sdk.AccountI) (sdk.AccountI,error) {
	nextAccNo,_:=ak.NextAccountNumber(ctx)
	if err := acc.SetAccountNumber(nextAccNo); err != nil {
		return nil,err
	}
	
	return acc,nil
}

// HasAccount implements AccountKeeperI.
func (ak AccountKeeper) HasAccount(ctx context.Context, addr sdk.AccAddress) (bool,error) {
	store := ak.storeService.OpenKVStore(ctx)
	has, err := store.Has(types.AddressStoreKey(addr))
	if err != nil {
		return false,err
	}
	return has,nil
}

// HasAccountAddressByID checks account address exists by id.
func (ak AccountKeeper) HasAccountAddressByID(ctx context.Context, id uint64) (bool,error) {
	store := ak.storeService.OpenKVStore(ctx)
	has, err := store.Has(types.AccountNumberStoreKey(id))
	if err != nil {
		return false,err
	}
	return has,nil
}

// GetAccount implements AccountKeeperI.
func (ak AccountKeeper) GetAccount(ctx context.Context, addr sdk.AccAddress) (sdk.AccountI,error) {
	store := ak.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.AddressStoreKey(addr))
	if err != nil {
		return nil,err
	}

	if bz == nil {
		return nil,nil
	}

	return ak.decodeAccount(bz),nil
}

// GetAccountAddressById returns account address by id.
func (ak AccountKeeper) GetAccountAddressByID(ctx context.Context, id uint64) (string,error) {
	store := ak.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.AccountNumberStoreKey(id))
	if err != nil {
		return "",err
	}

	if bz == nil {
		return "",nil
	}
	return sdk.AccAddress(bz).String(),nil
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
func (ak AccountKeeper) SetAccount(ctx context.Context, acc sdk.AccountI) error {
	addr := acc.GetAddress()
	store := ak.storeService.OpenKVStore(ctx)

	bz, err := ak.MarshalAccount(acc)
	if err != nil {
		return err
	}

	store.Set(types.AddressStoreKey(addr), bz)
	store.Set(types.AccountNumberStoreKey(acc.GetAccountNumber()), addr.Bytes())
	return nil
}

// RemoveAccount removes an account for the account mapper store.
// NOTE: this will cause supply invariant violation if called
func (ak AccountKeeper) RemoveAccount(ctx context.Context, acc sdk.AccountI) error {
	addr := acc.GetAddress()
	store := ak.storeService.OpenKVStore(ctx)
	err := store.Delete(types.AddressStoreKey(addr))
	if err != nil {
		return err
	}

	err = store.Delete(types.AccountNumberStoreKey(acc.GetAccountNumber()))
	if err != nil {
		return err
	}
	return nil
}

// IterateAccounts iterates over all the stored accounts and performs a callback function.
// Stops iteration when callback returns true.
func (ak AccountKeeper) IterateAccounts(ctx context.Context, cb func(account sdk.AccountI) (stop bool)) error {
	store := ak.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.AddressStoreKeyPrefix, storetypes.PrefixEndBytes(types.AddressStoreKeyPrefix))
	if err != nil {
		return err
	}

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		account := ak.decodeAccount(iterator.Value())

		if cb(account) {
			break
		}
	}
	return nil
}
